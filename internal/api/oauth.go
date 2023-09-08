package api

import (
	"fmt"
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/oauth"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/Hello-Storage/hello-back/pkg/crypto"
	"github.com/gin-gonic/gin"
)

// OAuthGoogle
//
// GET /api/oauth/google
func OAuthGoogle(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.GET("/oauth/google", func(ctx *gin.Context) {

		

		code := ctx.Query("code")


		if code == "" {
			log.Errorf("Authorization code not provided!")
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"status": "fail", "message": "Authorization code not provided!"},
			)
			return
		}

		google_user, err := oauth.GetGoogleUser(code)

		if err != nil {
			log.Errorf("failed to get google user: %v", err)
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		u := query.FindUserByEmail(google_user.Email)
		if u == nil {

			var req struct {
				WalletAddress string `json:"wallet_address" binding:"required"`
				PrivateKey   string `json:"private_key" binding:"required"`
			}

			req.WalletAddress = ctx.Query("wallet_address")
			req.PrivateKey = ctx.Query("private_key")

			isValidEthereumAddress := crypto.IsValidEthereumAddress(req.WalletAddress)
			isValidEthereumPrivateKey := crypto.IsValidEthereumPrivateKey(req.PrivateKey)
			if !isValidEthereumAddress || !isValidEthereumPrivateKey {
				log.Errorf("invalid ethereum address or private key")
				ctx.JSON(
					http.StatusBadRequest,
					gin.H{"status": "fail", "message": "invalid ethereum address or private key"},
				)
				return
			}

			encryptedPrivateKey, err := crypto.Encrypt(req.PrivateKey)
			if err != nil {
				log.Errorf("failed to encrypt private key: %v", err)
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
				return
			}

			fmt.Printf("Request: %+v\n", req)

			// create new user
			new := entity.User{
				Name: google_user.Name,
				Email: entity.Email{
					Email: google_user.Email,
				},
				Wallet: entity.Wallet{
					Address:    req.WalletAddress,
					PrivateKey: encryptedPrivateKey,
					Type: string(entity.Google),
				},
			}

			//decrypt andn print private key
			//decryptedPrivateKey, err := crypto.Decrypt(encryptedPrivateKey)
			if err := new.Create(); err != nil {
				log.Errorf("failed to create user: %v", err)
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
				return
			}

			// initialize user detail
			user_detail := entity.UserDetail{
				StorageUsed: 0,
				UserID:      new.ID,
			}

			if err := user_detail.Create(); err != nil {
				log.Errorf("failed to create user detail: %v", err)
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}

			u = &new
		}

		// authorization token
		accessToken, accessPayload, err := tokenMaker.CreateToken(
			u.ID,
			u.UID,
			u.Name,
			config.Env().AccessTokenDuration,
		)
		if err != nil {
			log.Errorf("failed to create access token: %v", err)
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
			log.Errorf("failed to create refresh token: %v", err)
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

// OAuthGithub
//
// GET /api/oauth/github
func OAuthGithub(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.GET("/oauth/github", func(ctx *gin.Context) {
		code := ctx.Query("code")

		if code == "" {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"status": "fail", "message": "Authorization code not provided!"},
			)
			return
		}

		token, err := oauth.GetGithubOAuthToken(code)

		if err != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		github_user, err := oauth.GetGithubUser(token)
		if err != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		u := query.FindUserByGithub(github_user.ID)
		if u == nil {
			new := entity.User{
				Name: github_user.Name,
				Github: entity.Github{
					GithubID: github_user.ID,
					Name:     github_user.Name,
					Avatar:   github_user.Avatar,
				},
				Detail: entity.UserDetail{
					StorageUsed: 0,
				},
			}

			if err := new.Create(); err != nil {
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}

			// initialize user detail
			user_detail := entity.UserDetail{
				StorageUsed: 0,
				UserID:      new.ID,
			}

			if err := user_detail.Create(); err != nil {
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}

			u = &new
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
