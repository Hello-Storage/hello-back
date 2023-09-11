package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Hello-Storage/hello-back/internal/query"
)

// FetchReferredUsers returns referred users for a given referral code.
//
// GET /api/referral/:referral_code
// Params:
// - referral_code string
func FetchReferredUsers(router *gin.RouterGroup) {
	router.GET("/referral/:referral_code", func(ctx *gin.Context) {
		
		referral_code := ctx.Param("referral_code")

		if referral_code == "" {
			log.Errorf("referral code not provided!")
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"status": "fail", "message": "referral code not provided!"},
			)
			return
		}

		users, err := query.FindReferredUsers(referral_code)

		if err != nil {
			log.Errorf("failed to get referred users: %v", err)
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": users})
	})
}
