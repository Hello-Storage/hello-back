package api

import (
	"net/http"
	"sync"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/Hello-Storage/hello-back/pkg/web3"
	"github.com/gin-gonic/gin"
)

var authMutex = sync.Mutex{}

// LoadUser
//
// GET /api/load
func LoadUser(router *gin.RouterGroup) {
	router.GET("/load", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		u := query.FindUserWithWallet(authPayload.UserID)

		if u == nil {
			log.Errorf("user not found: %d", authPayload.UserID)
			ctx.JSON(http.StatusNotFound, "user not found")
			return
		}

		log.Infof("user: %v", u.Detail)

		var resp = struct {
			UID           string `json:"uid"`
			Name          string `json:"name"`
			Role          string `json:"role"`
			WalletAddress string `json:"walletAddress"`
			Signature     string `json:"signature"`
		}{
			UID:           u.UID,
			Name:          u.Name,
			Role:          string(u.Role),
			WalletAddress: u.Wallet.Address,
			Signature:     u.Wallet.Signature,
		}

		ctx.JSON(http.StatusOK, resp)
	})
}

// LoginUser
//
// POST /api/login
func LoginUser(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.POST("/login", func(ctx *gin.Context) {
		var f form.LoginUserRequest
		if err := ctx.BindJSON(&f); err != nil {
			AbortBadRequest(ctx)
			return
		}

		authMutex.Lock()
		defer authMutex.Unlock()

		u := query.FindUserByWalletAddress(f.WalletAddress)
		if u == nil {
			Abort(ctx, http.StatusNotFound, "user not exists!")
			return
		}

		// retrieve nonce
		nonce, err := u.RetrieveNonce(false)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		log.Infof("nonce: %s", nonce)

		// validate signature
		result := web3.ValidateMessageSignature(
			f.WalletAddress,
			f.Signature,
			constant.BuildLoginMessage(nonce),
		)
		if !result {
			ctx.JSON(http.StatusBadRequest, "invalide signature")
			return
		}

		user_detail := query.FindUserDetailByUserID(u.ID)
		if user_detail.ReferredBy == 0 && f.Referral != "" {
			// check if referral code is valid
			referrer_id, _ := query.CheckReferralCode(f.Referral)

			log.Infof("referrer_id %d", referrer_id)
			// initialize user detail

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

		// TO-DO create session part

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

// RequestNonce
// POST /api/nonce
func RequestNonce(router *gin.RouterGroup) {
	router.POST("/nonce", func(ctx *gin.Context) {
		var req struct {
			WalletAddress string `json:"wallet_address" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		u := entity.User{
			Wallet: &entity.Wallet{
				Address: req.WalletAddress,
			},
		}

		log.Info("renew", u)
		nonce, err := u.RetrieveNonce(true)
		if err != nil {
			ctx.JSON(
				http.StatusInternalServerError,
				ErrorResponse(err),
			)
			return
		}
		ctx.JSON(http.StatusOK, nonce)
	})
}
