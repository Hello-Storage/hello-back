package api

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/oauth"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

// Internal func
func CreateNewWallet() (string, string, string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", "", err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyString := hexutil.Encode(privateKeyBytes)[2:]
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", "", errors.New("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	data := []byte(
		fmt.Sprintf(
			"https://hello.storage/\nPersonal signature\n\nWallet address:\n%s",
			address,
		),
	)
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", "", "", err
	}

	return privateKeyString, address, hexutil.Encode(signature), nil
}

// OAuthGoogle
//
// GET /api/oauth/google
func OAuthGoogle(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.GET("/oauth/google", func(ctx *gin.Context) {
		code := ctx.Query("code")
		referral := ctx.Query("referral")

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

		// Start a new transaction
		tx := db.Db().Begin()

		if u == nil {
			privateKey, publicKey, signature, err := CreateNewWallet()

			if err != nil {
				tx.Rollback()
				ctx.JSON(
					http.StatusInternalServerError,
					"can't create wallet",
				)
				return
			}

			// create new user
			new := entity.User{
				Name: google_user.Name,
				Email: &entity.Email{
					Email: google_user.Email,
				},
				Wallet: &entity.Wallet{
					Address:    publicKey,
					PrivateKey: privateKey,
					Signature:  signature,
				},
			}

			//decrypt and print private key
			//decryptedPrivateKey, err := crypto.Decrypt(encryptedPrivateKey)
			if err := new.TxCreate(tx); err != nil {
				log.Errorf("failed to create user: %v", err)
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
				return
			}

			// check if referral code is valid
			referrer_id, err := query.CheckReferralCode(referral)
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

			if err == nil {
				referral := entity.Referral{
					ReferrerID:   referrer_id,
					ReferredID:   new.ID,
					UserDetailID: user_detail.ID,
				}

				if err := referral.TxCreate(tx); err != nil {
					log.Errorf("failed to create referral: %v", err)
					tx.Rollback()
					ctx.JSON(
						http.StatusInternalServerError,
						gin.H{"status": "fail", "message": err.Error()},
					)
					return
				}
			}

			if err := query.UpdateReferralStorage(referrer_id); err != nil {
				log.Errorf("failed to create referral: %v", err)
				tx.Rollback()
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
			tx.Rollback()
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

// OAuthGithub
//
// GET /api/oauth/github
func OAuthGithub(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.GET("/oauth/github", func(ctx *gin.Context) {
		code := ctx.Query("code")
		referral := ctx.Query("referral")

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
			privateKey, publicKey, signature, err := CreateNewWallet()

			if err != nil {
				tx.Rollback()
				ctx.JSON(
					http.StatusInternalServerError,
					"can't create wallet",
				)
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
					Address:    publicKey,
					PrivateKey: privateKey,
					Signature:  signature,
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
			referrer_id, err := query.CheckReferralCode(referral)

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

			if err == nil {
				referral := entity.Referral{
					ReferrerID:   referrer_id,
					ReferredID:   new.ID,
					UserDetailID: user_detail.ID,
				}

				if err := referral.TxCreate(tx); err != nil {
					log.Errorf("failed to create referral: %v", err)
					tx.Rollback()
					ctx.JSON(
						http.StatusInternalServerError,
						gin.H{"status": "fail", "message": err.Error()},
					)
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
