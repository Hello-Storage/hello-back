package api

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/gin-gonic/gin"
)

// FetchReferredUsers returns referred users for a given referral code.
//
// GET /api/referral/:referral_code
// Params:
// - referral_code string
func FetchReferredUsers(router *gin.RouterGroup) {
	router.GET("/referrals/:referral_code", func(ctx *gin.Context) {

		referralCode := ctx.Param("referral_code")
		if referralCode == "" {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"status": "fail", "message": "referral code not provided!"},
			)
			return
		}

		addresses, err := query.FindReferredUsers(referralCode)

		if err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"status": "fail", "message": err.Error()},
			)
			return
		}


		referredByAddress := query.FindReferrerFromAddress(referralCode)

		response := gin.H{"status": "success"}
		if len(addresses) > 0 {
			response["referredAddresses"] = addresses
		}
		if referredByAddress != "" {
			response["referredBy"] = referredByAddress
		}
		ctx.JSON(http.StatusOK, response)

	})
}
