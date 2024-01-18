package api

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/constant"
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
	router.POST("/api_key", func(ctx *gin.Context) {

		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		authMutex.Lock()
		defer authMutex.Unlock()

		u, err := query.FindUserByUID(authPayload.UserID)
		if err != nil {
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

	router.GET("/api_key", func(c *gin.Context) {

		authPayload := c.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		p, err := query.FindApiKeyByUserID(authPayload.UserID)
		if err != nil {
			AbortEntityNotFound(c)
			return
		}
		c.JSON(http.StatusOK, p)
	})

}
