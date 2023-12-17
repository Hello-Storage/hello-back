package api

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

// ApiKey
//
// POST /api/api_key
func ApiKey(router *gin.RouterGroup, tokenMaker token.Maker) {
	router.POST("/api/api_key", func(ctx *gin.Context) {
		var f form.CreateApiKeyRequest
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

		// create api key
		apikey, accessPayload, err := tokenMaker.CreateApiKey(
			u.ID,
			u.UID,
			u.Name,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		rsp := form.CreateApiKeyResponse{
			ApiKey:          apikey,
			ApiKeyExpiresAt: accessPayload.ExpiredAt,
		}

		apikeyEntity := &entity.ApiKey{
			UserID: u.ID,
			ApiKey: apikey,
		}

		if err := apikeyEntity.Create(); err != nil {
			log.Errorf("failed to create api-key: %v", err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"status": "fail", "message": err.Error()},
			)
			return
		}
		ctx.JSON(http.StatusOK, rsp)
	})

}
