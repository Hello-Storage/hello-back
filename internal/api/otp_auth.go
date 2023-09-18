package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/mg"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// OTP Auth (one-time-passcode auth)
//
// POST /api/otp/start
func StartOTP(router *gin.RouterGroup) {
	router.POST("/otp/start", func(ctx *gin.Context) {
		var f struct {
			Email string `json:"email" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&f); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "joinhello.app",
			AccountName: f.Email,
			Period:      30 * 60,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		u := query.FindUserByEmail(f.Email)

		if u == nil {
			privateKey, publicKey, signature, err := CreateNewWallet()
			if err != nil {
				ctx.JSON(
					http.StatusInternalServerError,
					"can't create wallet",
				)
				return
			}

			// create new user
			u = &entity.User{
				Name: strings.Split(f.Email, "@")[0],
				Email: &entity.Email{
					Email:  f.Email,
					Secret: key.Secret(),
				},
				Wallet: &entity.Wallet{
					Address:    publicKey,
					PrivateKey: privateKey,
					Signature:  signature,
				},
				Detail: &entity.UserDetail{
					StorageUsed: 0,
				},
			}

			if err := u.Create(); err != nil {
				log.Errorf("failed to create user: %v", err)
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
				return
			}

		} else {
			email := u.Email
			email.Secret = key.Secret()

			if err := email.Save(); err != nil {
				log.Errorf("failed to save secret: %v", err)
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
				return
			}
		}

		code, err := totp.GenerateCode(key.Secret(), time.Now())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		mg := mg.Mailgun{
			Domain: "joinhello.app",
			ApiKey: config.Env().MailGunApiKey,
		}

		mg.Init()
		id, err := mg.SendEmail(
			"noreply@joinhello.app",
			f.Email,
			"Log in to JoinHello",
			"magic-code",
			map[string]interface{}{
				"code": code,
			},
		)

		log.Infof("id: %s", id)

		ctx.JSON(http.StatusOK, "success")
	})
}

// OTP Auth (one-time-passcode auth)
//
// POST /api/otp/start
func VerifyOTP(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.POST("/otp/verify", func(ctx *gin.Context) {
		var f struct {
			Email    string `json:"email" binding:"required"`
			Code     string `json:"code" binding:"required"`
			Referral string `json:"referral"`
		}

		if err := ctx.ShouldBindJSON(&f); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		u := query.FindUserByEmail(f.Email)
		if u == nil {
			ctx.JSON(http.StatusNotFound, "user not found")
			return
		}

		user_detail := query.FindUserDetailByUserID(u.ID)
		if user_detail.ReferredBy == 0 {
			referrer_id, _ := query.CheckReferralCode(f.Referral)
			user_detail.ReferredBy = referrer_id

			if err := user_detail.Save(); err != nil {
				log.Errorf("failed to update user detail: %v", err)
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}

			referral := &entity.Referral{
				ReferrerID:   referrer_id,
				ReferredID:   u.ID,
				UserDetailID: user_detail.ID,
			}

			if err := referral.Create(); err != nil {
				log.Errorf("failed to create referral: %v", err)
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}

			if err := query.UpdateReferralStorage(referrer_id); err != nil {
				log.Errorf("failed to create referral: %v", err)
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}
		}

		result := totp.Validate(f.Code, u.Email.Secret)
		log.Infof("code: %s, secret: %s", f.Code, u.Email.Secret)
		if !result {
			ctx.JSON(http.StatusBadRequest, "invalide code")
			return
		}

		// authorization token
		accessToken, accessPayload, err := tokenMaker.CreateToken(
			u.ID,
			u.UID,
			u.Name,
			config.Env().AccessTokenDuration,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		refreshToken, refreshPayload, err := tokenMaker.CreateToken(
			u.ID,
			u.UID,
			u.Name,
			config.Env().RefreshTokenDuration,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		rsp := form.LoginUserResponse{
			// SessionID:             session.ID,
			AccessToken:           accessToken,
			AccessTokenExpiresAt:  accessPayload.ExpiredAt,
			RefreshToken:          refreshToken,
			RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		}
		ctx.JSON(http.StatusOK, rsp)
	})
}
