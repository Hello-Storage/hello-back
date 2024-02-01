package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware creates a gin middleware for API key authorization
func APIKeyAuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKeyHeader := ctx.GetHeader(constant.APIKeyHeaderKey)
		fmt.Println("API Key Header constant: " + constant.APIKeyHeaderKey)
		fmt.Println("API Key Header: ")
		spew.Dump(ctx.Request.Header)
		if len(apiKeyHeader) == 0 {
			err := errors.New("1: API key header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse(err))
			return
		}

		payload, err := tokenMaker.VerifyApiKey(apiKeyHeader)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse(err))
			return
		}

		ctx.Set(constant.APIKeyHeaderKey, payload)
		ctx.Next()
	}
}
