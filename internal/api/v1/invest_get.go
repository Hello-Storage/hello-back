package v1

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/gin-gonic/gin"
)

func InvestGetDataByCode(router *gin.RouterGroup) {

	router.GET("/invest/all", func(ctx *gin.Context) {

		var investAccount []entity.InvestAccount
		result := db.Db().Find(&investAccount)

		if result.RowsAffected > 0 {

			ctx.JSON(http.StatusOK, gin.H{
				"isSuccess": "true",
				"message":   investAccount,
			})
		} else {

			ctx.JSON(http.StatusNotFound, gin.H{
				"isSuccess": "false",
				"message":   "No data found",
			})
		}

	})

	router.GET("/invest", func(ctx *gin.Context) {

		var investCodes []entity.InvestCode
		db.Db().Preload("InvestAccounts").Find(&investCodes)

		ctx.JSON(http.StatusOK, gin.H{
			"isSuccess": "true",
			"message":   investCodes,
		})

	})

}
