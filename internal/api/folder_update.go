package api

import (
	"net/http"
	"strconv"

	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

// UpdateFolderRoot updates folder root on database
//
// POST /api/folder/update/root
// formData: form.UpdateFolder
// @return 200 {string} string "ok"
func UpdateFolderRoot(router *gin.RouterGroup) {
	router.PUT("/folder/update/root", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		var form form.UpdateFolder

		if err := ctx.BindJSON(&form); err != nil {
			log.Error("ctx.BindJSON")
			log.Error(ctx.Errors)
			AbortBadRequest(ctx)
			return
		}
		uintID64, tmpError := strconv.ParseUint(form.Id, 10, 32)
		log.Info("uint ", tmpError)
		uintID := uint(uintID64)

		folderMutex.Lock()
		defer folderMutex.Unlock()

		// Check if the current user is the owner of the folder
		isOwner, err := entity.IsFolderOwner(uintID, authPayload.UserID)
		if err != nil {
			log.Error("entity.IsFolderOwner != nil")
			log.Error(ctx.Errors)
			AbortBadRequest(ctx)
			return
		}
		if !isOwner {
			// Forbidden: The user is not the owner of the folder
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this folder"})
			log.Error("entity.IsFolderOwner != isOwner")
			log.Error(ctx)
			return
		}

		folder := entity.Folder{
			ID:   uintID,
			UID:  form.Uid,
			Root: form.Root,
		}

		if err := folder.UpdateRootOnly(); err != nil {
			log.Error("folder.UpdateRootOnly")
			log.Error(ctx.Errors)
			AbortBadRequest(ctx)
			return
		}
		/*
			folder_user := entity.FolderUser{
				FolderID:   folder.ID,
				UserID:     authPayload.UserID,
				Permission: entity.OwnerPermission,
			}

			// Devulve un error, hay que investigar porque
			// hello-back\internal\entity\folder_user.go
			if err := folder_user.Update(); err != nil {
				log.Error("folder_user.Update")
				log.Error(err)
				AbortBadRequest(ctx)
				return
			}*/

		ctx.JSON(http.StatusOK, folder)
	})
}
