package v1

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/gin-gonic/gin"
)

func InvestGetDataByCode(router *gin.RouterGroup) {

	router.GET("/invest/:code", func(ctx *gin.Context) {

		code := ctx.Param("code")

		var investCode entity.InvestCode
		result := db.Db().Preload("InvestAccounts").Where("code = ?", code).First(&investCode)

		if result.RowsAffected > 0 {

			ctx.JSON(http.StatusOK, gin.H{
				"isSuccess": "true",
				"message":   investCode,
			})
		} else {

			ctx.JSON(http.StatusNotFound, gin.H{
				"isSuccess": "false",
				"message":   "code was not found",
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
