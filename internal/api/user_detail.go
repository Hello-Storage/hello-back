package api

import (
	"net/http"
	"time"

	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

type sharedListUser struct {
	sharedWithMe []form.FileResponse
	sharedByMe   []form.FileResponse
}

// UpdateUser updates the profile information of the currently authenticated user.
//
// GET /api/user/:uid
func GetUserDetail(router *gin.RouterGroup) {
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
			ctx.JSON(http.StatusNotFound, "user not found")
			return
		}

		filesUser, err := query.GetFilesUserFromUser(user.ID)

		if err != nil {
			ctx.JSON(http.StatusNotFound, "files not found")
			return
		}

		var sharedwithUser []form.FileResponse
		var sharedByUser []form.FileResponse

		for _, fileUser := range filesUser {
			file, err := query.FindFileByID(fileUser.FileID)
			formatedFileR := form.FileResponse{
				ID:                   file.ID,
				Name:                 file.Name,
				UID:                  file.UID,
				Root:                 file.Root,
				CID:                  file.CID,
				CIDOriginalEncrypted: file.CIDOriginalEncrypted,
				Mime:                 file.Mime,
				Size:                 file.Size,
				EnryptionStatus:      file.EncryptionStatus,
				IsInPool:             file.IsInPool,
				CreatedAt:            file.CreatedAt.String(),
				UpdatedAt:            file.UpdatedAt.String(),
			}

			if err != nil {
				ctx.JSON(http.StatusNotFound, "file not found")
				return
			}

			if fileUser.Permission == entity.SharedPermission {
				sharedwithUser = append(sharedwithUser, formatedFileR)
			} else {

				// Check if other users have the file
				usersWithFile, err := query.FindUsersByFileCID(file.CID)

				if err != nil {
					ctx.JSON(http.StatusNotFound, "file not found")
					return
				}
				if fileUser.Permission == entity.OwnerPermission {
					if len(usersWithFile) > 1 {
						sharedByUser = append(sharedByUser, formatedFileR)
					}
				}
			}
		}
		response := sharedListUser{
			sharedWithMe: sharedwithUser,
			sharedByMe:   sharedByUser,
		}

		ctx.JSON(http.StatusOK, response)
	})
}
