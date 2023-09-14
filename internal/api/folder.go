package api

import (
	"net/http"
	"sync"

	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

var folderMutex = sync.Mutex{}

// CreateFolder returns bool
//
// POST /api/folder/create
// formData: form.CreateFolder
func CreateFolder(router *gin.RouterGroup) {
	router.POST("/folder/create", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		var form form.CreateFolder

		if err := ctx.BindJSON(&form); err != nil {
			AbortBadRequest(ctx)
			return
		}

		folderMutex.Lock()
		defer folderMutex.Unlock()

		folder := entity.Folder{
			Title: form.Title,
			Root:  form.Root,
		}

		if err := folder.Create(); err != nil {
			AbortBadRequest(ctx)
			return
		}

		folder_user := entity.FolderUser{
			FolderID:   folder.ID,
			UserID:     authPayload.UserID,
			Permission: entity.OwnerPermission,
		}

		if err := folder_user.Create(); err != nil {
			AbortBadRequest(ctx)
			return
		}

		ctx.JSON(http.StatusOK, folder)
	})
}
