package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/crypto"
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

		var privateKey *string
		if u.Wallet.AccountType != string(entity.Provider) {

			decryptedKey, err := crypto.Decrypt(u.Wallet.PrivateKey)

			if err != nil {
				log.Errorf("failed to decrypt private key: %s", err)
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
				return
			}

			privateKey = &decryptedKey
		}

		var resp = struct {
			UID              string  `json:"uid"`
			Name             string  `json:"name"`
			Role             string  `json:"role"`
			WalletAddress    string  `json:"walletAddress"`
			WalletPrivateKey *string `json:"walletPrivateKey"`
		}{
			UID:              u.UID,
			Name:             u.Name,
			Role:             string(u.Role),
			WalletAddress:    u.Wallet.Address,
			WalletPrivateKey: privateKey,
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
		nonce, err := u.RetrieveNonce(false, f.Referral)
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
		userLogin := &entity.UserLogin{
			LoginDate:  time.Now(),
			WalletAddr: u.Wallet.Address,
		}

		if err := userLogin.Create(); err != nil {
			log.Errorf("failed to create user login: %v", err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"status": "fail", "message": err.Error()},
			)
			return
		}
		ctx.JSON(http.StatusOK, rsp)
	})

}

// RegisterUser
//
// POST /api/register
func RegisterUser(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.POST("/register", func(ctx *gin.Context) {
		var f form.RegisterUserRequest
		if err := ctx.BindJSON(&f); err != nil {
			AbortBadRequest(ctx)
			return
		}

		authMutex.Lock()
		defer authMutex.Unlock()

		u := entity.User{
			Name: f.Name,
		}

		// TO-DO check exists user info, if
		if user := query.FindUser(u); user != nil {
			Abort(ctx, http.StatusBadRequest, "user already exists!")
		}

		if err := u.Create(); err != nil {
			AbortInternalServerError(ctx)
			return
		}

		ctx.JSON(
			http.StatusOK,
			"user created!",
		)
	})
}

// RequestNonce
// POST /api/nonce
func RequestNonce(router *gin.RouterGroup) {
	router.POST("/nonce", func(ctx *gin.Context) {
		var req struct {
			WalletAddress string `json:"wallet_address" binding:"required"`
			ReferrerCode  string `json:"referral"`
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

		nonce, err := u.RetrieveNonce(true, req.ReferrerCode)
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
