package api

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/gin-gonic/gin"
)

// UpdateUser updates the profile information of the currently authenticated user.
//
// PUT /api/user/:uid
func UpdateUser(router *gin.RouterGroup) {
	router.PUT("/user", func(ctx *gin.Context) {
		user := entity.User{
			Name: "abc",
		}

		err := user.Create()

		if err != nil {
			AbortInternalServerError(ctx)
			return
		}

		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "pong",
			},
		)
	})
}