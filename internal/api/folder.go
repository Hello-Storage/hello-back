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
	"gorm.io/gorm"
)

var folderMutex = sync.Mutex{}

func ShareWithUserHandler(tx *gorm.DB, formget form.SharedFolder, parentRoot string, authPayload *token.Payload, shareType string, accountIdentifier string, ctx *gin.Context, shareWithUser *entity.User) {

	// Find the existing folder in the database
	foundFolder, err := query.FindFolderByUID(formget.Uid)
	fmt.Println(foundFolder)

	if err != nil {
		AbortNotFound(ctx)
		return
	}

	// Update the folder title
	if err := foundFolder.UpdateTitle(formget.Title); err != nil {
		AbortBadRequest(ctx)
		return
	}

	// Update the folder's encryption status to 'public'
	if err := foundFolder.UpdateEncryptionStatusAndCID(entity.Public, fmt.Sprintf("%d", authPayload.UserID)); err != nil {
		tx.Rollback()
		AbortBadRequest(ctx)
		return
	}

	folder := entity.Folder{
		Title:            formget.Title,
		Root:             parentRoot,
		EncryptionStatus: entity.Public,
		CID:              fmt.Sprintf("%d", authPayload.UserID) + foundFolder.UID,
		IsInPool:         true,
	}

	if err := folder.TxCreate(tx); err != nil {
		tx.Rollback()
		AbortBadRequest(ctx)
		return
	}

	folder_user := entity.FolderUser{
		FolderID:   folder.ID,
		UserID:     shareWithUser.ID,
		Permission: entity.SharedPermission,
	}

	if err := folder_user.TxCreate(tx); err != nil {
		tx.Rollback()
		log.Errorf("error when creating folder_user: %v", err)
		AbortBadRequest(ctx)
		return
	}

	// Find files in the folder based on its UID
	filesInFolder, _ := query.FindFilesByRoot(foundFolder.UID)

	// Validate that the number of files in the folder matches the number of files in the request payload
	if len(filesInFolder) != len(formget.Files) {
		Abort(ctx, http.StatusBadRequest, "files in folder don't match with files in the database")
		return
	}

	// Print a debug message indicating the start of the sharing process
	fmt.Println("Sharing folder:\n uid:", formget.Uid, "\n user:", authPayload.UserID)

	// Start the transaction

	// Iterate through each file in the request payload
	for _, file := range formget.Files {

		// Check if the file exists
		f, err := query.FindFileByUID(file.UID)
		if err != nil {
			log.Errorf("failed to get file: %s", err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			tx.Rollback()
			return
		}

		// Create a new file with the same metadata
		newFile := CreateNewFileFromMetadata(f, file)
		newFile.Root = folder.UID
		if err := newFile.TxCreate(tx); err != nil {
			log.Errorf("create file in folder from metadata: %s", err)
			tx.Rollback()
			AbortInternalServerError(ctx)
			return
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
			return
		}

		// delete the file share state user shared in case it exists
		CIDOriginalDecrypted := query.DeleteFileShareStatesUserShared(tx, f.UID, shareWithUser.ID)
		shareState, err := query.CreateShareStateUserShared(tx, newFile, shareWithUser.ID)
		if err != nil {
			log.Errorf("failed to create a new share state user shared: %s", err)
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share state"})
			return
		}

		//if CIDOriginalDecrypted is not empty, set eit to selectedShareFile.CIDOriginalDecrypted
		if CIDOriginalDecrypted != "" {
			file.CIDOriginalEncrypted = CIDOriginalDecrypted
		}

		// PublishFile crea un nuevo PublicFile y lo devuelve
		publicFile, err := query.PublishFileUserShared(tx, shareState, file)
		if err != nil {
			tx.Rollback()
			log.Errorf("failed to publish file: %s", err)
			// Devuelve un mensaje de error al cliente
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish file"})
			return
		}

		// Update the shareState with the new PublicFile
		shareState.PublicFileUserShared = *publicFile

		// Save the updated shareState.PublicFile
		err = shareState.PublicFileUserShared.TxSave(tx)
		if err != nil {
			tx.Rollback()
			log.Errorf("failed to save share state: %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save share state"})
			return
		}

	}

	for _, child := range formget.Folders {
		ShareWithUserHandler(tx, child.Folder, folder.UID, authPayload, shareType, accountIdentifier, ctx, shareWithUser)
	}
}
func ShareFolderHandler(formget form.SharedFolder, shareType string, ctx *gin.Context) {
	// Find the existing folder in the database
	foundFolder, err := query.FindFolderByUID(formget.Uid)

	if err != nil {
		AbortNotFound(ctx)
		return
	}

	tx := db.Db().Begin()

	// Update the folder title
	if err := foundFolder.UpdateTitle(formget.Title); err != nil {
		log.Errorf("failed to update folder title: %s", err)
		tx.Rollback()
		AbortBadRequest(ctx)
		return
	}

	// Update the folder's encryption status to 'public'
	if err := foundFolder.UpdateEncryptionStatus(entity.Public); err != nil {
		log.Errorf("failed to update folder encryption status: %s", err)
		tx.Rollback()
		AbortBadRequest(ctx)
		return
	}

	// Find files in the folder based on its UID
	filesInFolder, _ := query.FindFilesByRoot(foundFolder.UID)

	// Validate that the number of files in the folder matches the number of files in the request payload
	if len(filesInFolder) != len(formget.Files) {
		log.Errorf("files in folder don't match with files in the database")
		tx.Rollback()
		Abort(ctx, http.StatusBadRequest, "files in folder don't match with files in the database")
		return
	}

	// Iterate through each file in the request payload
	for _, file := range formget.Files {
		// Find the file in the database based on its UID
		f, err := query.FindFileByUID(file.UID)
		if err != nil {
			log.Errorf("failed to get file: %s", err)
		}

		query.DeleteFileShareState(tx, f.UID)

		// Find or create a sharing state for the file
		shareState, err := query.FindShareStateByFileUID(file.UID)
		if err != nil {
			shareState, err = query.CreateShareState(tx, f)
			if err != nil {
				tx.Rollback()
				log.Errorf("failed to create share state: %s", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share state"})
				return
			}
		}

		// Publish the file and get the corresponding PublicFile instance
		publicFile, err := query.PublishFile(tx, shareState, file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish file"})
			return
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

		// Save the updated shareState.PublicFile
		err = shareState.PublicFile.Save()

		if err != nil {
			fmt.Println("failed to save share state: ", err)
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save share state"})
			return
		}
	}

	if formget.Folders != nil && len(formget.Folders) > 0 {
		for _, childF := range formget.Folders {
			ShareFolderHandler(childF.Folder, shareType, ctx)
		}
	}

	tx.Commit()

}

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

		fmt.Println("Sharing folder\n user:", authPayload.UserID)

		// Lock to ensure exclusive access to shared resources
		folderMutex.Lock()
		defer folderMutex.Unlock()

		ShareFolderHandler(formget, shareType, ctx)

		// Respond
		ctx.JSON(http.StatusOK, "success")
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

		tx := db.Db().Begin()

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
			tx.Rollback()
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share type"})
			return
		}

		// Check if the user was found
		if shareWithUser == nil {
			tx.Rollback()
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		ShareWithUserHandler(tx, formget, "/", authPayload, shareType, accountIdentifier, ctx, shareWithUser)

		tx.Commit()

		// Respond
		ctx.JSON(http.StatusOK, "success")
	})

	router.GET("/folder/shared-uid/:uid", func(ctx *gin.Context) {
		// Extract authorization payload from the context
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		// Get the uid from the URL parameters
		uid := ctx.Param("uid")

		// Lock to ensure exclusive access to shared resources
		folderMutex.Lock()
		defer folderMutex.Unlock()

		fmt.Println("Getting shared folder content:\n uid:", uid, "\n user:", authPayload.UserID)

		// check if the folder exists
		foundFolder, err := query.FindFolderByUID(uid)
		if err != nil {
			AbortNotFound(ctx)
			return
		}

		// get files in the folder
		filesInFolder, _ := query.FindFilesByRoot(foundFolder.UID)

		publicfilesInFolder, _ := query.FindPublicFilesByRoot(foundFolder.UID)

		publicfilesUserSharedInFolder, _ := query.FindPublicFilesUserSharedByRoot(foundFolder.UID, authPayload.UserID)

		var resFiles []entity.File

		// check the size of the results
		if len(filesInFolder) != len(publicfilesInFolder) {
			
			for i, file := range filesInFolder {
				publicFileUserShared := publicfilesUserSharedInFolder[i]
				resFiles = append(resFiles, CreateFileForSharedFile(file, nil, &publicFileUserShared))

			}
		} else {
			for i, file := range filesInFolder {
				publicFile := publicfilesInFolder[i]
				resFiles = append(resFiles, CreateFileForSharedFile(file, &publicFile, nil))

				// Check additional conditions
				if publicFile.HasBeenOpened != nil && *publicFile.HasBeenOpened {
					// If HasBeenOpened is true, the file has been accessed before, and access is not allowed
					ctx.JSON(http.StatusForbidden, gin.H{"error": "File has already been accessed"})
					return
				}

				if publicFile.ExpireAt != nil && !publicFile.ExpireAt.IsZero() && time.Now().After(*publicFile.ExpireAt) {
					// If ExpireAt is a past date, the file has expired, and access is not allowed
					ctx.JSON(http.StatusForbidden, gin.H{"error": "File has expired"})
					return
				}

				// Update HasBeenOpened if necessary
				if publicFile.HasBeenOpened != nil && !*publicFile.HasBeenOpened {
					hasBeenOpened := true
					publicFile.HasBeenOpened = &hasBeenOpened
					// You could also update the access date here if necessary
					publicFile.UpdatedAt = time.Now()
					publicFile.Save()
				}

			}
		}

		// get folders in the folder
		foldersInFolder, _ := query.FoldersByRoot(foundFolder.UID)

		res := form.SharedFolderContentRes{
			Uid:     foundFolder.UID,
			Title:   foundFolder.Title,
			Files:   resFiles,
			Folders: foldersInFolder,
		}

		// Respond with the updated folder information
		ctx.JSON(http.StatusOK, res)
	})
}

func CreateFileForSharedFile(originalFile entity.File, publicFile *entity.PublicFile, publicFileUserShared *entity.PublicFileUserShared) entity.File {
	var isInPool bool = true
	shareState, _ := query.FindShareStateByFileUID(originalFile.UID)

	if publicFile != nil {
		return entity.File{
			ID:                   originalFile.ID,
			UID:                  originalFile.UID,
			CID:                  originalFile.CID,
			CIDOriginalEncrypted: nil,
			Name:                 publicFile.Name,
			Root:                 "",
			Mime:                 publicFile.Mime,
			Size:                 publicFile.Size,
			MediaType:            originalFile.MediaType,
			EncryptionStatus:     entity.Public,
			CreatedAt:            publicFile.CreatedAt,
			UpdatedAt:            publicFile.UpdatedAt,
			IsInPool:             &isInPool,
			DeletedAt:            originalFile.DeletedAt,
			Path:                 originalFile.Path,
			FileShareState:       shareState,
		}
	} else {
		return entity.File{
			ID:                   originalFile.ID,
			UID:                  originalFile.UID,
			CID:                  originalFile.CID,
			CIDOriginalEncrypted: nil,
			Name:                 publicFileUserShared.Name,
			Root:                 "",
			Mime:                 publicFileUserShared.Mime,
			Size:                 publicFileUserShared.Size,
			MediaType:            originalFile.MediaType,
			EncryptionStatus:     entity.Public,
			CreatedAt:            originalFile.CreatedAt,
			UpdatedAt:            originalFile.UpdatedAt,
			IsInPool:             &isInPool,
			DeletedAt:            originalFile.DeletedAt,
			Path:                 originalFile.Path,
			FileShareState:       shareState,
		}
	}
}
