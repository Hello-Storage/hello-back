package api

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

// UpdateUser updates the profile information of the currently authenticated user.
//
// GET /api/user/:uid
func GetUserDetail(router *gin.RouterGroup) {
	router.GET("/user/detail", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		user_detail := query.FindUserDetailByUserID(authPayload.UserID)

		if user_detail == nil {
			ctx.JSON(http.StatusNotFound, "user detail not found")
			return
		}

		ctx.JSON(http.StatusOK, user_detail)
	})
}
