package v1

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/gin-gonic/gin"
)

type InvestCodeRequest struct {
	IP            string `json:"ip"`
	Email         string `json:"email"`
	SocialNetwork string `json:"social_network"`
}

func InvestPostData(router *gin.RouterGroup) {

	router.POST("/invest", func(ctx *gin.Context) {

		code := ctx.Query("code")

		var request InvestCodeRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "body request not found or incomplete"})
			return
		}

		var investCode entity.InvestCode
		result := db.Db().Preload("InvestAccounts").Where("code = ?", code).First(&investCode)

		if result.RowsAffected > 0 {
			newInvestAccount := entity.InvestAccount{IP: request.IP, Code: code}
			db.Db().Create(&newInvestAccount)

			ctx.JSON(http.StatusOK, gin.H{
				"isSuccess": true,
				"message":   investCode,
			})
		} else {
			//Creacion de Invest_codes

			// newInvestCode := entity.InvestCode{Code: code, Email: request.Email, SocialNetwork: request.SocialNetwork}
			// db.Db().Create(&newInvestCode)

			//Creacion de Invest_account

			// newInvestAccount := entity.InvestAccount{IP: request.IP, Code: code}
			// db.Db().Create(&newInvestAccount)

			ctx.JSON(http.StatusOK, gin.H{
				"isSuccess": false,
				"message":   "The code is not found",
			})
		}

	})

}
