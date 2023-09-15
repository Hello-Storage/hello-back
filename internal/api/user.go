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
// PUT /api/user/signature
func UpdateUserSignature(router *gin.RouterGroup) {
	router.POST("/user/signature", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		u := query.FindUserWithWallet(authPayload.UserID)
		if u == nil {
			log.Errorf("user not found: %d", authPayload.UserID)
			ctx.JSON(http.StatusNotFound, "user not found")
			return
		}

		var f struct {
			Signature string `json:"signature" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&f); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		u.Wallet.Signature = f.Signature

		log.Infof("signature", u.Wallet.Signature)
		if err := u.Save(); err != nil {
			log.Errorf("failed to save user signature: %v", err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		ctx.JSON(
			http.StatusOK,
			"success",
		)
	})
}
