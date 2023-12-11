package api

import (
	"net/http"
	"time"

	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SharedListUser struct {
	SharedWithMe entity.Files
	SharedByMe   entity.Files
}

// UpdateUser updates the profile information of the currently authenticated user.
//
// GET /api/user/:uid
func GetUserDetail(router *gin.RouterGroup) {
	router.Use(cors.Default())
	router.GET("/user/detail", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		user_detail := query.FindUserDetailByUserID(authPayload.UserID)

		user := query.FindUser(entity.User{ID: authPayload.UserID})

		if user == nil {
			ctx.JSON(http.StatusNotFound, "user not found")
			return
		}

		if user_detail == nil {
			ctx.JSON(http.StatusNotFound, "user detail not found")
			return
		}

		userLogin := &entity.UserLogin{
			LoginDate:  time.Now(),
			WalletAddr: user.Wallet.Address, //this is the line that is giving the panic
		}

		if err := userLogin.Create(); err != nil {
			log.Errorf("failed to create user login: %v", err)
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"status": "fail", "message": err.Error()},
			)
			return
		}

		ctx.JSON(http.StatusOK, user_detail)
	})

	router.GET("/user/shared/general", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		user := query.FindUser(entity.User{ID: authPayload.UserID})
		if user == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		filesUser, err := query.GetFilesUserFromUser(user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching user files"})
			return
		}

		var sharedwithUser entity.Files
		var sharedByUser entity.Files

		for _, fileUser := range filesUser {
			file, err := query.FindFileByID(fileUser.FileID)
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching file"})
					return
				}
				continue
			}

			if fileUser.Permission == entity.SharedPermission {
				sharedwithUser = append(sharedwithUser, *file)
			} else {
				usersWithFile, err := query.FindUsersByFileCID(file.CID)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching users with file"})
					return
				}
				if fileUser.Permission == entity.OwnerPermission && len(usersWithFile) > 1 {
					sharedByUser = append(sharedByUser, *file)
				}
			}
		}

		response := SharedListUser{
			SharedWithMe: sharedwithUser,
			SharedByMe:   sharedByUser,
		}

		ctx.JSON(http.StatusOK, response)
	})

}
