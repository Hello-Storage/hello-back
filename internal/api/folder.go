package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

var folderMutex = sync.Mutex{}
var ShareWithUserHandler func(formget form.SharedFolder, authPayload *token.Payload, shareType string, accountIdentifier string, ctx *gin.Context, shareWithUser *entity.User) *entity.Folder
var ShareFolderHandler func(formget form.SharedFolder, authPayload *token.Payload, shareType string, ctx *gin.Context) *entity.Folder

// CreateFolder returns bool
//
// POST /api/folder/create
// formData: form.CreateFolder
func CreateFolder(router *gin.RouterGroup) {

	ShareWithUserHandler := func(formget form.SharedFolder, authPayload *token.Payload, shareType string, accountIdentifier string, ctx *gin.Context, shareWithUser *entity.User) *entity.Folder {

		// Find the existing folder in the database
		foundFolder, err := query.FindFolderByUID(formget.Uid)

		if err != nil {
			AbortNotFound(ctx)
			return nil
		}

		// Update the folder title
		if err := foundFolder.UpdateTitle(formget.Title); err != nil {
			AbortBadRequest(ctx)
			return nil
		}

		// Update the folder's encryption status to 'public'
		if err := foundFolder.UpdateEncryptionStatus(entity.Public); err != nil {
			AbortBadRequest(ctx)
			return nil
		}

		folder_user := entity.FolderUser{
			FolderID:   foundFolder.ID,
			UserID:     shareWithUser.ID,
			Permission: entity.SharedPermission,
		}

		if err := folder_user.Create(); err != nil {
			AbortBadRequest(ctx)
			return nil
		}

		// Find files in the folder based on its UID
		filesInFolder, err := query.FindFilesByRoot(foundFolder.UID)
		if err != nil {
			Abort(ctx, http.StatusNoContent, "folder is empty")
			return nil
		}

		// Validate that the number of files in the folder matches the number of files in the request payload
		if len(filesInFolder) != len(formget.Files) {
			Abort(ctx, http.StatusNoContent, "files in folder don't match with files in the database")
			return nil
		}

		// Print a debug message indicating the start of the sharing process
		fmt.Println("Sharing folder:\n uid:", formget.Uid, "\n user:", authPayload.UserID)

		// Start the transaction
		tx := db.Db().Begin()

		// Iterate through each file in the request payload
		for _, file := range formget.Files {

			// Check if the file exists
			f, err := query.FindFileByUID(file.UID)
			if err != nil {
				log.Errorf("failed to get file: %s", err)
				ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
				tx.Rollback()
				return nil
			}

			// Create a new file with the same metadata
			newFile := CreateNewFileFromMetadata(f, file)
			if err := newFile.TxCreate(tx); err != nil {
				log.Errorf("create file: %s", err)
				tx.Rollback()
				AbortInternalServerError(ctx)
				return nil
			}

			// Create a FilesUsers entry to share the file with the specified user
			fileUser := &entity.FileUser{
				FileID:     newFile.ID,
				UserID:     shareWithUser.ID,
				Permission: entity.SharedPermission,
			}

			if err := fileUser.TxCreate(tx); err != nil {
				log.Errorf("create file_user relation: %s", err)
				tx.Rollback()
				AbortInternalServerError(ctx)
				return nil
			}

			// Print a debug message indicating the sharing details
			fmt.Printf("Sharing file:\nshareType: %s\nfile UID: %s\nshared with: %s\n", entity.SharedPermission, file.UID, accountIdentifier)

		}

		// Commit the transaction
		tx.Commit()

		for _, child := range formget.Folders {
			ShareWithUserHandler(child.Folder, authPayload, shareType, accountIdentifier, ctx, shareWithUser)
		}

		return foundFolder
	}

	ShareFolderHandler := func(formget form.SharedFolder, authPayload *token.Payload, shareType string, ctx *gin.Context) *entity.Folder {

		// Find the existing folder in the database
		foundFolder, err := query.FindFolderByUID(formget.Uid)

		if err != nil {
			AbortNotFound(ctx)
			return nil
		}

		// Update the folder title
		if err := foundFolder.UpdateTitle(formget.Title); err != nil {
			AbortBadRequest(ctx)
			return nil
		}

		// Update the folder's encryption status to 'public'
		if err := foundFolder.UpdateEncryptionStatus(entity.Public); err != nil {
			AbortBadRequest(ctx)
			return nil
		}

		// Find files in the folder based on its UID
		filesInFolder, err := query.FindFilesByRoot(foundFolder.UID)
		if err != nil {
			Abort(ctx, http.StatusNoContent, "folder is empty")
			return nil
		}

		// Validate that the number of files in the folder matches the number of files in the request payload
		if len(filesInFolder) != len(formget.Files) {
			Abort(ctx, http.StatusNoContent, "files in folder don't match with files in the database")
			return nil
		}

		// Print a debug message indicating the start of the sharing process
		fmt.Println("Sharing folder:\n uid:", formget.Uid, "\n user:", authPayload.UserID)

		// Iterate through each file in the request payload
		for _, file := range formget.Files {
			// Find the file in the database based on its UID
			f, err := query.FindFileByUID(file.UID)
			if err != nil {
				log.Errorf("failed to get file: %s", err)
			}

			// Find or create a sharing state for the file
			shareState, err := query.FindShareStateByFileUID(file.UID)
			if err != nil {
				shareState, err = query.CreateShareState(f)
				if err != nil {
					log.Errorf("failed to create share state: %s", err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share state"})
					return nil
				}
			}

			// Publish the file and get the corresponding PublicFile instance
			publicFile, err := query.PublishFile(shareState, file)
			if err != nil {
				log.Errorf("failed to publish file: %s", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish file"})
				return nil
			}

			// Update the sharing state with the new PublicFile
			shareState.PublicFile = *publicFile

			// Set values based on the sharing type
			var expireDate *time.Time
			switch shareType {
			case "public":
				// No changes needed
			case "one-time":
				hasBeenOpened := false
				shareState.PublicFile.HasBeenOpened = &hasBeenOpened
			case "monthly":
				tmpExpireDate := time.Now().AddDate(0, 1, 0)
				expireDate = &tmpExpireDate
				shareState.PublicFile.HasBeenOpened = nil
			}

			// Set ExpireAt outside of the switch
			shareState.PublicFile.ExpireAt = expireDate

			// Debugging: Print the current state of shareState.PublicFile
			fmt.Printf("After switch: ExpireAt: %v, HasBeenOpened: %v\n", shareState.PublicFile.ExpireAt, shareState.PublicFile.HasBeenOpened)

			// Save the updated shareState.PublicFile
			err = shareState.PublicFile.Save()
			if err != nil {
				log.Errorf("failed to save share state: %s", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save share state"})
				return nil
			}
		}

		for _, child := range formget.Folders {
			ShareFolderHandler(child.Folder, authPayload, shareType, ctx)
		}
		return foundFolder

	}

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
			Title:            form.Title,
			Root:             form.Root,
			EncryptionStatus: form.EncryptionStatus,
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

	router.POST("/folder/share/:shareType", func(ctx *gin.Context) {
		// Extract authorization payload from the context
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		// Get the sharing type from the URL parameters
		shareType := ctx.Param("shareType")

		// Parse the request payload into a SharedFolder form
		var formget form.SharedFolder
		if err := ctx.BindJSON(&formget); err != nil {
			AbortBadRequest(ctx)
			return
		}

		// Lock to ensure exclusive access to shared resources
		folderMutex.Lock()
		defer folderMutex.Unlock()

		foundFolder := ShareFolderHandler(formget, authPayload, shareType, ctx)

		res := form.SharedFolderRes{
			Folder: *foundFolder,
		}

		// Respond with the updated folder information
		ctx.JSON(http.StatusOK, res)
	})

	router.POST("/folder/share/:shareType/:user", func(ctx *gin.Context) {
		// Extract authorization payload from the context
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		// Get the share type from the parameters
		shareType := ctx.Param("shareType")

		// Get the account identifier (email or wallet) from the parameters
		accountIdentifier := ctx.Param("user")

		// Parse the request payload into a SharedFolder form
		var formget form.SharedFolder
		if err := ctx.BindJSON(&formget); err != nil {
			AbortBadRequest(ctx)
			return
		}

		// Lock to ensure exclusive access to shared resources
		folderMutex.Lock()
		defer folderMutex.Unlock()

		// Determine the appropriate function to retrieve the user based on the share type
		var shareWithUser *entity.User
		switch shareType {
		case "email":
			shareWithUser = query.FindUserByEmail(accountIdentifier)
		case "wallet":
			shareWithUser = query.FindUserByWalletAddress(accountIdentifier)
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share type"})
			return
		}

		// Check if the user was found
		if shareWithUser == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		foundFolder := ShareWithUserHandler(formget, authPayload, shareType, accountIdentifier, ctx, shareWithUser)

		res := form.SharedFolderRes{
			Folder: *foundFolder,
		}

		// Respond with the updated folder information
		ctx.JSON(http.StatusOK, res)
	})

}
