package api

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

var fileMutex = sync.Mutex{}

// UpadteFileRoot updates file root on database
//
// POST /api/file/update/root
// formData: form.UpdateFile
// @return 200 {string} string "ok"
func UpdateFileRoot(router *gin.RouterGroup) {
	router.PUT("/update/root", func(ctx *gin.Context) {
		fmt.Println("UpdateFileRoot")
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		var form form.UpdateFileRoot

		if err := ctx.BindJSON(&form); err != nil {
			log.Errorf("ctx.BindJSON: %s", err)
			AbortBadRequest(ctx)
			return
		}

		uintID64, _ := strconv.ParseUint(form.Id, 10, 32)

		uintID := uint(uintID64)

		fileMutex.Lock()
		defer fileMutex.Unlock()

		// Check if the current user is the owner of the folder
		isOwner, err := entity.IsFileOwner(uintID, authPayload.UserID)
		if err != nil {
			log.Errorf("entity.IsFileOwner != nil: %s", err)
			AbortBadRequest(ctx)
			return
		}
		if !isOwner {
			log.Errorf("entity.IsFileOwner != isOwner: %s", err)
			// Forbidden: The user is not the owner of the folder
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this file"})
			return
		}

		file := entity.File{}

		// Query the database to populate the 'File' entity with existing data.
		if err := db.Db().Where("UID = ?", form.Uid).First(&file).Error; err != nil {
			log.Errorf("db.Db().Where != nil: %s", err)
			AbortBadRequest(ctx)
			return
		}
		file.ID = uintID
		file.UID = form.Uid
		file.Root = form.Root

		if err := file.UpdateRootOnly(); err != nil {
			log.Errorf("file.UpdateRootOnly != nil: %s", err)
			AbortBadRequest(ctx)
			return
		}

		// Add select db for return id of file

		//file_user := entity.FileUser{
		//FileID:     file.ID,
		//UserID:     authPayload.UserID,
		//Permission: entity.OwnerPermission,
		//}

		//if err := file_user.Update(); err != nil {
		//AbortBadRequest(ctx)
		//return
		//}

		ctx.JSON(http.StatusOK, file)
	})
}
