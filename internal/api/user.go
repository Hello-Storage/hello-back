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

// GetUserCount returns the total number of registered users and referred users.

// GET /api/user/count
func GetUserCount(router *gin.RouterGroup) {
	router.GET("/user/count", func(ctx *gin.Context) {
	//use gin to count user
	var count int64

	var user entity.User

	count, err := user.Count()

	if err != nil {
		AbortInternalServerError(ctx)
		return
	}

	var referredCount int64

	var referredUser entity.ReferredUser

	referredCount, err = referredUser.Count()

	if err != nil {
		AbortInternalServerError(ctx)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"user_count":      count,
			"ns":  referredCount,
		},
	)
})
}