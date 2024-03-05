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

// ShareWithUserHandler takes in the necessary parameters and shares a folder and its contents with a specified user.
func ShareWithUserHandler(formget form.SharedFolder, parentRoot string, authPayload *token.Payload, shareType string, accountIdentifier string, shareWithUser *entity.User) bool {
	tx := db.Db().Begin() // Start a new transaction

	// Find the existing folder in the database
	foundFolder, err := query.FindFolderByUID(formget.Uid)

	if err != nil {
		return false
	}

	// Update the folder title
	if err := foundFolder.UpdateTitle(formget.Title); err != nil {
		return false
	}

	// Update the folder's encryption status to 'public'
	if err := foundFolder.UpdateEncryptionStatusAndCID(entity.Public, fmt.Sprintf("%d", authPayload.UserID)); err != nil {
		return false
	}

	// Create a new folder
	folder := entity.Folder{
		Title:            formget.Title,
		Root:             parentRoot,
		EncryptionStatus: entity.Public,
		CID:              fmt.Sprintf("%d", authPayload.UserID) + foundFolder.UID,
		IsInPool:         true,
	}

	if err := folder.TxCreate(tx); err != nil {
		tx.Rollback()
		return false
	}

	tx.Commit() // Commit the transaction after creating the new folder

	// Start a new transaction
	tx = db.Db().Begin()

	// Share the folder with the user
	folder_user := entity.FolderUser{
		FolderID:   folder.ID,
		UserID:     shareWithUser.ID,
		Permission: entity.SharedPermission,
	}

	if err := folder_user.TxCreate(tx); err != nil {
		tx.Rollback()
		log.Errorf("error when creating folder_user: %v", err)
		return false
	}

	tx.Commit() // Commit the transaction after sharing the folder with the user

	// Find files in the folder based on its UID
	filesInFolder, err := query.FindFilesByRoot(foundFolder.UID)
	if err != nil {
		return false
	}

	// Validate that the number of files in the folder matches the number of files in the request payload
	if len(filesInFolder) != len(formget.Files) {
		log.Errorf("number of files in folder does not match the number of files in request payload")
		return false
	}

	// Iterate through each file in the request payload
	for _, file := range formget.Files {
		// Start a new transaction
		tx = db.Db().Begin()

		// Check if the file exists
		f, err := query.FindFileByUID(file.UID)
		if err != nil {
			log.Errorf("failed to get file: %s", err)
			return false
		}

		// Create a new file with the same metadata
		newFile := CreateNewFileFromMetadata(f, file)
		newFile.Root = folder.UID
		if err := newFile.TxCreate(tx); err != nil {
			log.Errorf("create file in folder from metadata: %s", err)
			return false
		}

		tx.Commit() // Commit the transaction after creating the new file

		// Start a new transaction
		tx = db.Db().Begin()

		// Create a FilesUsers entry to share the file with the specified user
		fileUser := &entity.FileUser{
			FileID:     newFile.ID,
			UserID:     shareWithUser.ID,
			Permission: entity.SharedPermission,
		}

		if err := fileUser.TxCreate(tx); err != nil {
			log.Errorf("create file_user relation: %s", err)
			tx.Rollback()
			return false
		}

		tx.Commit() // Commit the transaction after sharing the file with the user

		// Start a new transaction
		tx = db.Db().Begin()

		// Delete the file share state user shared in case it exists
		query.DeleteFileShareStatesUserShared(db.Db(), f.UID, shareWithUser.ID)

		tx.Commit() // Commit the transaction after deleting the file share state

		// Start a new transaction
		tx = db.Db().Begin()

		// Create a share state for the file shared with the user
		shareState, err := query.CreateShareStateUserShared(tx, newFile, shareWithUser.ID)
		if err != nil {
			log.Errorf("failed to create a new share state user shared: %s", err)
			tx.Rollback()
			return false
		}

		tx.Commit() // Commit the transaction after creating the new file

		tx = db.Db().Begin() // Start a new transaction

		// PublishFile crea un nuevo PublicFile y lo devuelve
		publicFile, err := query.PublishFileUserShared(tx, shareState, file)
		if err != nil {
			tx.Rollback()
			log.Errorf("failed to publish file: %s", err)
			return false
		}

		tx.Commit() // Commit the transaction after creating the new file

		tx = db.Db().Begin() // Start a new transaction

		// Update the shareState with the new PublicFile
		shareState.PublicFileUserShared = *publicFile

		// Save the updated shareState.PublicFile
		err = shareState.PublicFileUserShared.Save()
		if err != nil {
			tx.Rollback()
			log.Errorf("failed to save share state: %s", err)
			return false
		}

		tx.Commit() // Commit the transaction after creating the new file

	}

	for _, child := range formget.Folders {
		res := ShareWithUserHandler(child.Folder, folder.UID, authPayload, shareType, accountIdentifier, shareWithUser)

		if !res {
			return false
		}
	}

	return true
}
// ShareFolderHandler handles the sharing of a folder.
//
// It takes a form.SharedFolder and a shareType string as parameters and returns a boolean.
func ShareFolderHandler(formget form.SharedFolder, shareType string) bool {

	// Find the existing folder in the database
	foundFolder, err := query.FindFolderByUID(formget.Uid)

	if err != nil {
		log.Errorf("failed to get folder: %s", err)
		return false
	}

	// Update the folder title
	if err := foundFolder.UpdateTitle(formget.Title); err != nil {
		log.Errorf("failed to update folder title: %s", err)
		return false
	}

	// Update the folder's encryption status to 'public'
	if err := foundFolder.UpdateEncryptionStatus(entity.Public); err != nil {
		log.Errorf("failed to update folder encryption status: %s", err)
		return false
	}

	tx := db.Db().Begin()

	// Find files in the folder based on its UID
	filesInFolder, _ := query.FindFilesByRoot(foundFolder.UID)

	// Validate that the number of files in the folder matches the number of files in the request payload
	if len(filesInFolder) != len(formget.Files) {
		log.Errorf("files in folder don't match with files in the database")
		return false
	}

	tx.Commit() // Commit the transaction

	// Iterate through each file in the request payload
	for _, file := range formget.Files {
		tx = db.Db().Begin() // Start a new transaction
		// Find the file in the database based on its UID
		f, err := query.FindFileByUID(file.UID)
		if err != nil {
			log.Errorf("failed to get file: %s", err)
			continue
		}
		query.DeleteFileShareState(tx, f.UID)
		tx.Commit()          // Commit the transaction
		tx = db.Db().Begin() // Start a new transaction

		// Find or create a sharing state for the file
		shareState, _ := query.CreateShareState(tx, f) // errors prohibited here

		tx.Commit()          // Commit the transaction
		tx = db.Db().Begin() // Start a new transaction

		// Publish the file and get the corresponding PublicFile instance
		publicFile, err := query.PublishFile(tx, *shareState, file)
		if err != nil {
			log.Errorf("failed to publish file: %s", err)
			return false
		}

		tx.Commit()          // Commit the transaction
		tx = db.Db().Begin() // Start a new transaction

		// Set values based on the sharing type
		var expireDate *time.Time
		switch shareType {
		case "public":
			// No changes needed
		case "one-time":
			hasBeenOpened := false
			publicFile.HasBeenOpened = &hasBeenOpened
		case "monthly":
			tmpExpireDate := time.Now().AddDate(0, 1, 0)
			expireDate = &tmpExpireDate
			publicFile.HasBeenOpened = nil
		}

		// Set ExpireAt outside of the switch
		publicFile.ExpireAt = expireDate
		// Save the updated shareState.PublicFile
		publicFile.Save()

		// Update the sharing state with the new PublicFile
		shareState.PublicFile = *publicFile
		// Save the updated shareState
		shareState.Save()

		tx.Commit() // Commit the transaction
	}

	// Iterate through each folder in the request payload
	if formget.Folders != nil && len(formget.Folders) > 0 {
		for _, childF := range formget.Folders {
			res := ShareFolderHandler(childF.Folder, shareType)

			if !res {
				return false
			}
		}
	}

	tx.Commit()

	return true
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

		res := ShareFolderHandler(formget, shareType)

		if !res {
			AbortBadRequest(ctx)
			return
		}

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

		res := ShareWithUserHandler(formget, "/", authPayload, shareType, accountIdentifier, shareWithUser)

		if !res {
			AbortBadRequest(ctx)
			return
		}

		// Respond
		ctx.JSON(http.StatusOK, "success")
	})

	router.GET("/folder/shared-uid/:uid", func(ctx *gin.Context) {

		// Get the uid from the URL parameters
		uid := ctx.Param("uid")

		// Lock to ensure exclusive access to shared resources
		folderMutex.Lock()
		defer folderMutex.Unlock()

		// check if the folder exists
		foundFolder, err := query.FindFolderByUID(uid)
		if err != nil {
			AbortNotFound(ctx)
			return
		}

		// get files in the folder
		filesInFolder, err := query.FindFilesByRoot(foundFolder.UID)
		if err != nil {
			AbortNotFound(ctx)
			return
		}

		publicfilesInFolder, err := query.FindPublicFilesByRoot(foundFolder.UID)
		if err != nil {
			AbortNotFound(ctx)
			return
		}

		publicfilesUserSharedInFolder, err := query.FindPublicFilesUserSharedByRoot(foundFolder.UID)
		if err != nil {
			AbortNotFound(ctx)
			return
		}

		// Create a slice of shared files
		var resFiles []entity.File

		// check the size of the results
		if len(filesInFolder) != len(publicfilesInFolder) {

			for i, file := range filesInFolder {
				publicFileUserShared := publicfilesUserSharedInFolder[i]
				filledFile := CreateFileForSharedFile(file, nil, &publicFileUserShared)
				if filledFile != nil {
					resFiles = append(resFiles, *filledFile)
				}

			}
		} else {
			for i, file := range filesInFolder {
				publicFile := publicfilesInFolder[i]
				filledFile := CreateFileForSharedFile(file, &publicFile, nil)
				if filledFile != nil {
					resFiles = append(resFiles, *CreateFileForSharedFile(file, &publicFile, nil))
				}

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

func CreateFileForSharedFile(
	originalFile entity.File,
	publicFile *entity.PublicFile,
	publicFileUserShared *entity.PublicFileUserShared,
) *entity.File {
	var isInPool = true
	shareState, shareStateUserShared, err := query.FindShareStateByFileUID(originalFile.UID)

	if err != nil {
		return nil
	}

	if publicFile != nil && shareState != nil {
		return createFile(originalFile, publicFile, isInPool, shareState)
	}

	if publicFileUserShared != nil && shareStateUserShared != nil {
		return createFileWithUserShared(originalFile, publicFileUserShared, publicFileUserShared, isInPool, shareStateUserShared)
	}

	return nil
}

func createFile(originalFile entity.File,
	publicFile *entity.PublicFile, isInPool bool,
	shareState *entity.FileShareState) *entity.File {
	return &entity.File{
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
		FileShareState:       *shareState,
	}
}

func createFileWithUserShared(originalFile entity.File,
	publicFile *entity.PublicFileUserShared,
	publicFileUserShared *entity.PublicFileUserShared, isInPool bool,
	shareStateUserShared *entity.FileShareStatesUserShared) *entity.File {
	return &entity.File{
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
		CreatedAt:            originalFile.CreatedAt,
		UpdatedAt:            originalFile.UpdatedAt,
		IsInPool:             &isInPool,
		DeletedAt:            originalFile.DeletedAt,
		Path:                 originalFile.Path,
		FileShareState: entity.FileShareState{
			ID:      shareStateUserShared.ID,
			FileUID: shareStateUserShared.FileUID,
			PublicFile: entity.PublicFile{
				ID:                   1,
				Name:                 publicFileUserShared.Name,
				Mime:                 publicFileUserShared.Mime,
				Size:                 publicFileUserShared.Size,
				ShareHash:            publicFileUserShared.ShareHash,
				CIDOriginalDecrypted: publicFileUserShared.CIDOriginalDecrypted,
			},
		},
	}
}
