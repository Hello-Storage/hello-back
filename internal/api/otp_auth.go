package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/crypto"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

const ETHERMAIL_API_URL = "https://hub-gateway.ethermail.io/v1/transactional-emails"

var ETHERMAIL_API_KEY = config.Env().EthermailApiKey
var ETHERMAIL_API_KEY_SECRET = config.Env().EthermailApiSecret

type EmailRequest struct {
	SenderName  string `json:"sender_name"`
	SenderEmail string `json:"sender_email"`
	ToName      string `json:"to_name"`
	ToEmail     string `json:"to_email"`
	Subject     string `json:"subject"`
	ReplyTo     string `json:"reply_to"`
	TemplateID  string `json:"template_id"`
	MergeData   any    `json:"merge_data"`
}

// OTP Auth (one-time-passcode auth)
//
// POST /api/otp/start
func StartOTP(router *gin.RouterGroup) {
	router.POST("/otp/start", func(ctx *gin.Context) {
		var f struct {
			Email         string `json:"email" binding:"required"`
			ReferrerCode  string `json:"referrer_code"`
			WalletAddress string `json:"wallet_address"`
			PrivateKey    string `json:"private_key"`
		}

		if err := ctx.ShouldBindJSON(&f); err != nil {
			log.Errorf("failed to bind json: %v", err)
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err, "/otp/start:00000001"))
			return
		}

		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "hello.app",
			AccountName: f.Email,
			Period:      30 * 60,
		})
		if err != nil {
			log.Errorf("failed to generate key: %v", err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err, "/otp/start:00000002"))
			return
		}

		uSearch := query.FindUserByEmail(f.Email)

		tx := db.Db().Begin()

		if uSearch == nil {

			isValidEthereumAddress := crypto.IsValidEthereumAddress(f.WalletAddress)
			isValidEthereumPrivateKey := crypto.IsValidEthereumPrivateKey(f.PrivateKey)
			if !isValidEthereumAddress || !isValidEthereumPrivateKey {
				log.Errorf("invalid wallet address or private key")
				ctx.JSON(http.StatusBadRequest,
					ErrorResponse(errors.New("invalid wallet address or private key"), "/otp/start:00000003"))
				return
			}

			encryptedPrivateKey, err := crypto.Encrypt(f.PrivateKey)
			if err != nil {
				log.Errorf("failed to encrypt private key: %v", err)
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err, "/otp/start:00000004"))
				return
			}

			// create new user
			u := entity.User{
				Name: strings.Split(f.Email, "@")[0],
				Email: &entity.Email{
					Email:  f.Email,
					Secret: key.Secret(),
				},
				Wallet: &entity.Wallet{
					Address:     f.WalletAddress,
					PrivateKey:  encryptedPrivateKey,
					AccountType: string(entity.Mail),
				},
			}

			if err := u.Create(); err != nil {
				log.Errorf("failed to create user: %v", err)
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err, "/otp/start:00000005"))
				return
			}

			// check if referral code is valid
			if f.ReferrerCode == "ns" {
				referral := entity.ReferredUser{
					ReferredID: u.ID,
					Referrer:   f.ReferrerCode,
				}
				if err := referral.TxCreate(tx); err != nil {
					log.Errorf("failed to save referral: %v", err)
					tx.Rollback()
					ctx.JSON(http.StatusInternalServerError, ErrorResponse(err, "/otp/start:00000006"))
					return
				}
			}

			referrer_id, err := query.CheckReferralCode(f.ReferrerCode)
			if err != nil {
				log.Errorf("failed to check referral code: %v", err)
			}
			// initialize user detail
			user_detail := entity.UserDetail{
				StorageUsed: 0,
				UserID:      u.ID,
				ReferredBy:  referrer_id,
			}

			if err := user_detail.Create(); err != nil {
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err, "/otp/start:00000007"))
				return
			}

			if err == nil && referrer_id != 0 && user_detail.ID != 0 && u.ID != 0 {
				err := query.CreateReferral(referrer_id, u.ID, user_detail.ID)

				if err != nil {
					log.Errorf("failed to create referral: %v", err)
					ctx.JSON(http.StatusInternalServerError, ErrorResponse(err, "/otp/start:00000008"))
					return
				}
			}
			uSearch = &u

		} else {
			email := uSearch.Email
			email.Secret = key.Secret()

			if err := email.Save(); err != nil {
				log.Errorf("failed to save email: %v", err)
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err, "/otp/start:00000009"))
				return
			}
		}

		userLogin := &entity.UserLogin{
			LoginDate:  time.Now(),
			WalletAddr: uSearch.Wallet.Address,
		}

		if err := userLogin.TxCreate(tx); err != nil {
			log.Errorf("failed to create user login: %v", err)
			tx.Rollback()
			ctx.JSON(
				http.StatusInternalServerError,
				ErrorResponse(err, "/otp/start:00000010"),
			)
			return
		}

		code, err := totp.GenerateCode(key.Secret(), time.Now())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err, "/otp/start:00000011"))
			return
		}

		email := EmailRequest{
			SenderName:  "hello.app",
			SenderEmail: "noreply@hello.app",
			ToName:      f.Email,
			ToEmail:     f.Email,
			Subject:     "Hello from hello.app!",
			ReplyTo:     f.Email,
			TemplateID:  "6807a4239d805166a3cebc12",
			MergeData: map[string]interface{}{
				"code":          code,
				"walletAddress": f.WalletAddress,
			},
		}

		payloadBuf := new(bytes.Buffer)
		if err := json.NewEncoder(payloadBuf).Encode(email); err != nil {
			log.Errorf("failed to encode payload: %v", err)
			return
		}

		req, err := http.NewRequest("POST", ETHERMAIL_API_URL, payloadBuf)
		if err != nil {
			log.Errorf("failed to create request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", ETHERMAIL_API_KEY)
		req.Header.Set("x-api-secret", ETHERMAIL_API_KEY_SECRET)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Errorf("failed to send request: %v", err)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusCreated {
			log.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			return
		}

		tx.Commit()

		ctx.JSON(http.StatusOK, "success")
	})
}

// OTP Auth (one-time-passcode auth)
//
// POST /api/otp/verify
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
