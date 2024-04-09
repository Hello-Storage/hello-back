package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/crypto"
	"github.com/Hello-Storage/hello-back/pkg/oauth"
	"github.com/Hello-Storage/hello-back/pkg/token"
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
				ErrorResponse(errors.New("code not provided"), "/oauth/google:00000001"),
			)
			return
		}

		google_user, err := oauth.GetGoogleUser(code)

		if err != nil {
			log.Errorf("failed to get google user: %v", err)
			ctx.JSON(http.StatusBadGateway, 
				ErrorResponse(err, "/oauth/google:00000002"),
			)
			return
		}

		u := query.FindUserByEmail(google_user.Email)

		// Start a new transaction
		tx := db.Db().Begin()

		if u == nil {

			var req struct {
				WalletAddress string `json:"wallet_address" binding:"required"`
				PrivateKey    string `json:"private_key" binding:"required"`
				ReferralCode  string `json:"referrer_code" binding:"required"`
			}

			req.WalletAddress = ctx.Query("wallet_address")
			req.PrivateKey = ctx.Query("private_key")
			req.ReferralCode = ctx.Query("referrer_code")

			isValidEthereumAddress := crypto.IsValidEthereumAddress(req.WalletAddress)
			isValidEthereumPrivateKey := crypto.IsValidEthereumPrivateKey(req.PrivateKey)
			if !isValidEthereumAddress || !isValidEthereumPrivateKey {
				log.Errorf("invalid ethereum address or private key")
				ctx.JSON(
					http.StatusBadRequest,
					ErrorResponse(errors.New("invalid wallet address or private key"), "/oauth/google:00000003"),
				)
				return
			}

			encryptedPrivateKey, err := crypto.Encrypt(req.PrivateKey)
			if err != nil {
				log.Errorf("failed to encrypt private key: %v", err)
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err,"/oauth/google:00000004"))
				return
			}

			// create new user
			new := entity.User{
				Name: google_user.Name,
				Email: &entity.Email{
					Email: google_user.Email,
				},
				Wallet: &entity.Wallet{
					Address:     req.WalletAddress,
					PrivateKey:  encryptedPrivateKey,
					AccountType: string(entity.Google),
				},
			}

			//decrypt and print private key
			//decryptedPrivateKey, err := crypto.Decrypt(encryptedPrivateKey)
			if err := new.Create(); err != nil {
				log.Errorf("failed to create user: %v", err)
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err,"/oauth/google:00000005"))
				return
			}

			// check if referral code is valid
			if req.ReferralCode == "ns" {
				referral := entity.ReferredUser{
					ReferredID:  new.ID,
					Referrer:   req.ReferralCode,
				}
				if err := referral.TxCreate(tx); err != nil {
					log.Errorf("failed to save referral: %v", err)
					tx.Rollback()
					ctx.JSON(http.StatusInternalServerError, ErrorResponse(err,"/oauth/google:00000006"))
					return
				}
			}

			referrer_id, err := query.CheckReferralCode(req.ReferralCode)

			if err != nil {
				log.Errorf("failed to check referral code: %v", err)
			}

			// initialize user detail
			user_detail := entity.UserDetail{
				StorageUsed: 0,
				UserID:      new.ID,
				ReferredBy:  referrer_id,
			}

			if err := user_detail.TxCreate(tx); err != nil {
				log.Errorf("failed to create user detail: %v", err)
				tx.Rollback()
				ctx.JSON(
					http.StatusInternalServerError,
					ErrorResponse(err,"/oauth/google:00000007"),
				)
				return
			}

			if err == nil && referrer_id != 0 && user_detail.ID != 0 && new.ID != 0 {
				err := query.CreateReferral(referrer_id, new.ID, user_detail.ID)

				if err != nil {
					log.Errorf("failed to create referral: %v", err)
					ctx.JSON(http.StatusInternalServerError,ErrorResponse(err,"/otp/start:00000008"))
					return
				}
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
			tx.Rollback()
			log.Errorf("failed to create access token: %v", err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err,"/oauth/google:00000010"))
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
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err,"/oauth/google:00000011"))
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

		if err := userLogin.TxCreate(tx); err != nil {
			log.Errorf("failed to create user login: %v", err)
			tx.Rollback()
			ctx.JSON(
				http.StatusInternalServerError,
				ErrorResponse(err,"/oauth/google:00000012"),
			)
			return
		}

		tx.Commit()
		ctx.JSON(http.StatusOK, rsp)
	})
}

// OAuthGithub
//
// GET /api/oauth/github
func OAuthGithub(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.GET("/oauth/github", func(ctx *gin.Context) {
		log.Printf("github oauth call")
		code := ctx.Query("code")

		if code == "" {
			log.Errorf("Authorization code not provided!")
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
		// Start a new transaction
		tx := db.Db().Begin()
		if u == nil {
			var req struct {
				WalletAddress string `json:"wallet_address" binding:"required"`
				PrivateKey    string `json:"private_key" binding:"required"`
				ReferralCode  string `json:"referral_code" binding:"required"`
			}

			req.WalletAddress = ctx.Query("wallet_address")
			req.PrivateKey = ctx.Query("private_key")
			req.ReferralCode = ctx.Query("referrer_code")

			isValidEthereumAddress := crypto.IsValidEthereumAddress(req.WalletAddress)
			isValidEthereumPrivateKey := crypto.IsValidEthereumPrivateKey(req.PrivateKey)
			if !isValidEthereumAddress || !isValidEthereumPrivateKey {
				log.Errorf("invalid ethereum address or private key")
				tx.Rollback()
				ctx.JSON(
					http.StatusBadRequest,
					gin.H{"status": "fail", "message": "invalid ethereum address or private key"},
				)
				return
			}

			encryptedPrivateKey, err := crypto.Encrypt(req.PrivateKey)
			if err != nil {
				log.Errorf("failed to encrypt private key: %v", err)
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
				return
			}

			// create new user
			new := entity.User{
				Name: github_user.Name,
				Github: &entity.Github{
					GithubID: github_user.ID,
					Name:     github_user.Name,
					Avatar:   github_user.Avatar,
				},
				Detail: &entity.UserDetail{
					StorageUsed: 0,
				},
				Wallet: &entity.Wallet{
					Address:     req.WalletAddress,
					PrivateKey:  encryptedPrivateKey,
					AccountType: string(entity.GitHub),
				},
			}

			if err := new.TxCreate(tx); err != nil {
				log.Errorf("failed to create user: %v", err)
				tx.Rollback()
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}

			// check if referral code is valid
			if req.ReferralCode == "ns" {
				referral := entity.ReferredUser{
					ReferredID: u.ID,
					Referrer:   req.ReferralCode,
				}
				if err := referral.TxCreate(tx); err != nil {
					log.Errorf("failed to save referral: %v", err)
					tx.Rollback()
					ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
					return
				}
			}
			referrer_id, err := query.CheckReferralCode(req.ReferralCode)

			// initialize user detail
			user_detail := entity.UserDetail{
				StorageUsed: 0,
				UserID:      new.ID,
				ReferredBy:  referrer_id,
			}

			if err := user_detail.TxCreate(tx); err != nil {
				log.Errorf("failed to create user detail: %v", err)
				tx.Rollback()
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "fail", "message": err.Error()},
				)
				return
			}


			if err == nil && referrer_id != 0 && user_detail.ID != 0 && new.ID != 0 {
				err := query.CreateReferral(referrer_id, new.ID, user_detail.ID)

				if err != nil {
					log.Errorf("failed to create referral: %v", err)
					ctx.JSON(http.StatusInternalServerError,ErrorResponse(err,"/otp/start:00000008"))
					return
				}
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
			tx.Rollback()
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
			tx.Rollback()
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
		tx.Commit()
		ctx.JSON(http.StatusOK, rsp)
	})
}
