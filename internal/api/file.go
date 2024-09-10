package api

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/mg"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

// GetFile returns file details as JSON.
//
// GET /api/file/info/:uid
// Params:
// - uid
func GetFile(router *gin.RouterGroup) {

	router.GET("/info/:uid", func(c *gin.Context) {
		// To Do check access grant
		uid := c.Param("uid")

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		p, err := query.FindFileByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})

	router.GET("/apikey", func(c *gin.Context) {
		authPayload := c.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		pageNumber := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("pageSize", "10")

		pageNumberInt, err := strconv.Atoi(pageNumber)
		if err != nil || pageNumberInt < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil || pageSizeInt < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
			return
		}

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		allFiles, err := query.GetApiFiles(authPayload.UserID)
		if err != nil {
			fmt.Print(err)
		}

		totalItems := len(allFiles)
		totalPages := int(math.Ceil(float64(totalItems) / float64(pageSizeInt)))

		startIndex := (pageNumberInt - 1) * pageSizeInt
		endIndex := pageNumberInt * pageSizeInt

		if startIndex >= totalItems {
			c.JSON(http.StatusOK, gin.H{"files": []form.FileResponse{}, "totalItems": totalItems, "totalPages": totalPages})
			return
		}

		if endIndex > totalItems {
			endIndex = totalItems
		}

		paginatedFiles := allFiles[startIndex:endIndex]

		var fileResponses []form.FileResponse

		for _, f := range paginatedFiles {
			fResponse := form.FileResponse{
				ID:              f.ID,
				Name:            f.Name,
				UID:             f.UID,
				Root:            f.Root,
				CID:             f.CID,
				Mime:            f.Mime,
				Size:            f.Size,
				EnryptionStatus: f.EncryptionStatus,
				CreatedAt:       f.CreatedAt.String(),
				UpdatedAt:       f.UpdatedAt.String(),
			}

			fileResponses = append(fileResponses, fResponse)
		}

		c.JSON(http.StatusOK, gin.H{
			"files":      fileResponses,
			"totalItems": totalItems,
			"totalPages": totalPages,
		})
	})

}

// GetShareState returns file share state based on file id.
//
// GET /share/state
// Params:
// - file_uid
func GetShareState(router *gin.RouterGroup) {
	router.GET("/share/state", func(c *gin.Context) {
		//get file id from params
		file_uid := c.Query("file_uid")

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		//check if file exists
		f, err := query.FindFileByUID(file_uid)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		tx := db.Db().Begin()

		//get share state, if doesn't exist, create it
		share_state, _, err := query.FindShareStateByFileUID(file_uid)
		if err != nil {
			//log.Errorf("Error finding share state: %s", err)
			share_state, err = query.CreateShareState(tx, f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				tx.Rollback()
				AbortEntityNotFound(c)
				return
			}
		}

		tx.Commit()

		c.JSON(http.StatusOK, share_state)
	})

	router.GET("/share/states", func(c *gin.Context) {
		// Get file UIDs from query params
		fileUIDs := c.QueryArray("file_uids")

		if len(fileUIDs) == 0 {
			fmt.Println("No file UIDs received")
			AbortEntityNotFound(c)
		}

		tx := db.Db().Begin()

		var shareStates []entity.FileShareState

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		// Iterate through file UIDs
		for _, fileUID := range fileUIDs {
			// check if file exists
			f, err := query.FindFileByUID(fileUID)
			if err != nil {
				log.Errorf("cannot get file: %s", err)
				// skip this file if it doesn't exist
				continue
			}

			// get share state, if doesn't exist, create it
			shareState, _, err := query.FindShareStateByFileUID(fileUID)
			if err != nil {
				//log.Errorf("Error finding share state: %s", err)
				shareState, err = query.CreateShareState(tx, f)
				if err != nil {
					log.Errorf("cannot create share state: %s", err)
					// skip this file if share state creation fails
					continue

				}
			}

			shareStates = append(shareStates, *shareState)
		}

		tx.Commit()
		c.JSON(http.StatusOK, shareStates)
	})
}

// PublishFile publishes a file.
//
// POST /share/publish
// Params:
// - selectedShareFile: form.CustomFileMeta
func PublishFile(router *gin.RouterGroup) {
	router.POST("/share/group", func(c *gin.Context) {
		var request struct {
			ShareHashes []string `json:"share_hashes" binding:"required"`
		}

		// Bind JSON request body to the request struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON format"})
			return
		}

		if len(request.ShareHashes) == 0 {
			fmt.Println("No share hashes received")
			c.JSON(400, gin.H{"error": "No share hashes received"})
			return
		}

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		tx := db.Db().Begin()

		// Create a new share group
		shareGroup := entity.ShareGroup{}
		if err := shareGroup.TxCreate(tx); err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Failed to create share group"})
			return
		}

		// Associate each share_hash with the newly created share group
		for _, shareHash := range request.ShareHashes {
			publicFileShareGroup := entity.PublicFileShareGroup{
				ShareGroupHash: shareGroup.Hash,
				ShareHash:      shareHash,
			}

			if err := publicFileShareGroup.TxCreate(tx); err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": "Failed to associate share_hash with share group"})
				return
			}
		}
		tx.Commit()
		c.JSON(200, gin.H{"share_group": shareGroup.Hash})
	})

	router.GET("/share/group/:shareGroupHash", func(c *gin.Context) {
		shareGroupHash := c.Param("shareGroupHash")

		// Check if the share group Hash is provided and valid
		if shareGroupHash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No share group Hash provided"})
			return
		}

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		// Perform the query using the queryShareGroup function
		shareHashes, err := query.QueryShareGroupByHash(shareGroupHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve share hashes"})
			return
		}

		// Respond with the list of share hashes
		c.JSON(http.StatusOK, gin.H{"share_hashes": shareHashes})
	})

	router.POST("/share/custom-type/:shareType", func(c *gin.Context) {
		// Get the sharing type from the parameters
		shareType := c.Param("shareType")

		// Validate that the sharing type is valid
		if shareType != "public" && shareType != "one-time" && shareType != "monthly" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share type"})
			return
		}

		// Get the file ID from the parameters
		var selectedShareFile form.CustomFileMeta
		if err := c.BindJSON(&selectedShareFile); err != nil {
			log.Errorf("failed to bind JSON: %s", err)
			AbortEntityNotFound(c)
			return
		}

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		// Check if the file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("failed to get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		tx := db.Db().Begin()

		// Get the sharing state; if it doesn't exist, create it
		shareState, _, err := query.FindShareStateByFileUID(selectedShareFile.UID)
		if err != nil {
			shareState, err = query.CreateShareState(tx, f)
			if err != nil {
				log.Errorf("failed to create new share state: %s", err)
				// Devuelve un mensaje de error al cliente
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share state"})
				return
			}
		}

		// PublishFile crea un nuevo PublicFile y lo devuelve
		publicFile, err := query.PublishFile(tx, *shareState, selectedShareFile)
		if err != nil {
			log.Errorf("failed to publish file: %s", err)
			// Devuelve un mensaje de error al cliente
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish file"})
			return
		}

		// Update the shareState with the new PublicFile
		shareState.PublicFile = *publicFile

		// Set the values based on the sharing type
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
		err = shareState.PublicFile.TxSave(tx)
		if err != nil {
			tx.Rollback()
			log.Errorf("failed to save share state: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save share state"})
			return
		}

		tx.Commit()

		c.JSON(http.StatusOK, shareState)
	})

	router.POST("/share/:shareType/:user", func(ctx *gin.Context) {
		// Get the share type from the parameters
		shareType := ctx.Param("shareType")

		// Get the account identifier (email or wallet) from the parameters
		accountIdentifier := ctx.Param("user")
		//get auth payload
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		// Check if the user is sharing with themselves
		if shareType == "wallet" && accountIdentifier == authPayload.UserName {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot share with yourself"})
			return
		}

		// Get the file metadata from the request body
		var selectedShareFile form.CustomFileMeta
		if err := ctx.BindJSON(&selectedShareFile); err != nil {
			log.Errorf("failed to bind JSON: %s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to bind JSON"})
			return
		}

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

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

		//cancel if user not found
		receiverNil := shareWithUser == nil || shareWithUser.ID == 0
		if receiverNil {
			log.Errorf("user not found: %s", accountIdentifier)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Check if the file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("failed to get file: %s", err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// Start the transaction
		tx := db.Db().Begin()

		// Create a new file with the same metadata
		newFile := CreateNewFileFromMetadata(f, selectedShareFile)
		if err := newFile.TxCreate(tx); err != nil {
			log.Errorf("create file: %s", err)
			AbortInternalServerError(ctx)
			return
		}
		tx.Commit() // Commit the transaction after creating the new file

		tx = db.Db().Begin() // Start a new transaction

		// Create a FilesUsers entry to share the file with the specified user
		if !receiverNil {
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
		}

		// delete the file share state user shared in case it exists
		query.DeleteFileShareStatesUserShared(db.Db(), f.UID, shareWithUser.ID)
		// create a new share state user shared
		shareState, err := query.CreateShareStateUserShared(tx, newFile, shareWithUser.ID)
		if err != nil {
			log.Errorf("failed to create a new share state user shared: %s", err)
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share state"})
			return
		}

		// PublishFile crea un nuevo PublicFile y lo devuelve
		publicFile, err := query.PublishFileUserShared(tx, shareState, selectedShareFile)
		if err != nil {
			log.Errorf("failed to publish file: %s", err)
			tx.Rollback()
			// Devuelve un mensaje de error al cliente
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish file"})
			return
		}

		// Update the shareState with the new PublicFile
		shareState.PublicFileUserShared = *publicFile

		// Send email with the file link to the user if the share type is email
		if shareType == "email" {
			// Send email with the file link to the user and pass also the sender user's email
			sendEmailLinkToUser(authPayload.UserName, shareWithUser, accountIdentifier, newFile, publicFile)
		}

		// Save the updated shareState.PublicFile
		err = shareState.PublicFileUserShared.TxSave(tx)
		if err != nil {
			tx.Rollback()
			log.Errorf("failed to save share state: %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save share state"})
			return
		}

		// Commit the transaction
		tx.Commit()

		ctx.JSON(http.StatusOK, gin.H{"message": "File shared successfully",
			"data": gin.H{"file": newFile, "shareState": shareState, "publicFile": publicFile}})
	})

}

func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func sendEmailLinkToUser(username string, user *entity.User, email string, file *entity.File, publicFile *entity.PublicFileUserShared) {

	mg := mg.Mailgun{
		Domain: "hello.app",
		ApiKey: config.Env().MailGunApiKey,
	}

	mg.Init()
	id, err := mg.SendEmail(
		"noreply@hello.app",
		email,
		"hello.app | Received file named "+file.Name+"",
		"received-file",
		map[string]interface{}{
			"filename":   file.Name,
			"filesize":   formatBytes(file.Size),
			"filelink":   "https://hello.app/space/shared/public/" + publicFile.ShareHash,
			"sendername": username,
			"username":   email,
		},
	)

	log.Infof("id: %s", id)

	if err != nil {
		log.Errorf("failed to send email: %v", err)
	}

}

func CreateNewFileFromMetadata(originalFile *entity.File, metadata form.CustomFileMeta) *entity.File {
	var isInPool bool = true

	return &entity.File{
		Name:                 metadata.Name,
		Root:                 "/",
		CID:                  metadata.CID,
		CIDOriginalEncrypted: nil,
		Mime:                 metadata.MimeType,
		Size:                 metadata.Size,
		EncryptionStatus:     entity.Public,
		CreatedAt:            time.Now(),
		IsInPool:             &isInPool,
		UpdatedAt:            time.Now(),
	}
}

// UnpublishFile unpublishes a file.
//
// POST /share/unpublish
// Params:
// - selectedShareFile: form.CustomFileMeta
func UnpublishFile(router *gin.RouterGroup) {
	router.POST("/share/unpublish", func(c *gin.Context) {
		// get file id from params
		var selectedShareFile form.CustomFileMeta
		err := c.BindJSON(&selectedShareFile)
		if err != nil {
			log.Errorf("cannot bind json: %s", err)
			AbortEntityNotFound(c)
			return
		}

		tx := db.Db().Begin()

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		// check if file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		// get share state, if doesn't exist, create it
		shareState, _, err := query.FindShareStateByFileUID(selectedShareFile.UID)
		if err != nil {
			shareState, err = query.CreateShareState(tx, f)
			if err != nil {
				tx.Rollback()
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		// update share state
		// delete public file from db
		// update share state
		if shareState.PublicFile.ID != 0 { // Check if the PublicFile has a valid ID
			// Delete the corresponding record in publicFileShareGroups
			err = query.DeletePublicFileShareGroupByShareHash(tx, shareState.PublicFile.ShareHash)
			if err != nil {
				tx.Rollback()
				log.Errorf("cannot delete public file share group: %s", err)
				AbortEntityNotFound(c)
				return
			}

			// Delete the PublicFile
			err = shareState.PublicFile.TxDelete(tx)
			if err != nil {
				tx.Rollback()
				log.Errorf("cannot delete public file: %s", err)
				AbortEntityNotFound(c)
				return
			}

			// Check if the ShareGroup is empty after deleting the PublicFile
			err = query.DeleteEmptyShareGroup(tx, shareState.PublicFile.ShareHash)
			if err != nil {
				tx.Rollback()
				log.Errorf("cannot delete empty share group: %s", err)
				AbortEntityNotFound(c)
				return
			}
		} else {
			log.Info("No existing public file to delete.")
		}

		// Clear the association with PublicFile
		shareState.PublicFile = entity.PublicFile{}

		// Update the share state
		shareState.UpdatedAt = time.Now()

		err = shareState.TxSave(tx)

		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			tx.Rollback()
			AbortEntityNotFound(c)
			return
		}

		tx.Commit()

		c.JSON(http.StatusOK, shareState)
	})
}

// GetPublishedFile publishes a file.
//
// GET /share/published/:hash
// Params:
// - hash: form.PublicFile.ShareHash
func GetPublishedFile(router *gin.RouterGroup) {
	router.GET("/share/published/:hash", func(c *gin.Context) {
		// Get the hash from the parameters
		hash := c.Param("hash")

		tx := db.Db().Begin()

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		// Get the public file
		publicFile, publicFileUserShared, err := query.FindPublicFileByHash(hash)
		if err != nil {
			tx.Rollback()
			log.Errorf("cannot get public file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		var f *entity.File
		var res entity.File

		if publicFile.ID != 0 {
			// Check additional conditions
			if publicFile.HasBeenOpened != nil && *publicFile.HasBeenOpened {
				// If HasBeenOpened is true, the file has been accessed before, and access is not allowed
				tx.Rollback()
				c.JSON(http.StatusForbidden, gin.H{"error": "File has already been accessed"})
				return
			}

			if publicFile.ExpireAt != nil && !publicFile.ExpireAt.IsZero() && time.Now().After(*publicFile.ExpireAt) {
				// If ExpireAt is a past date, the file has expired, and access is not allowed
				tx.Rollback()
				c.JSON(http.StatusForbidden, gin.H{"error": "File has expired"})
				return
			}

			// Update HasBeenOpened if necessary
			if publicFile.HasBeenOpened != nil && !*publicFile.HasBeenOpened {
				hasBeenOpened := true
				publicFile.HasBeenOpened = &hasBeenOpened
				// You could also update the access date here if necessary
				publicFile.UpdatedAt = time.Now()
				err = publicFile.TxSave(tx)
				if err != nil {
					log.Errorf("failed to update public file: %s", err)
					tx.Rollback()
					AbortEntityNotFound(c)
					return
				}
			}

			// get the original file
			f, err = query.FindFileByUID(publicFile.FileUID)
			if err != nil {
				log.Errorf("cannot get publicFile's file: %s", err)
				tx.Rollback()
				AbortEntityNotFound(c)
				return
			}
			aux := CreateFileForSharedFile(*f, publicFile, nil)
			if aux == nil {
				tx.Rollback()
				AbortEntityNotFound(c)
				return
			}
			res = *aux
		} else {
			// get the original file
			f, err = query.FindFileByUID(publicFileUserShared.FileUID)
			if err != nil {
				log.Errorf("cannot get publicFileUserShared's file: %s", err)
				tx.Rollback()
				AbortEntityNotFound(c)
				return
			}

			aux := CreateFileForSharedFile(*f, nil, publicFileUserShared)
			if aux == nil {
				tx.Rollback()
				AbortEntityNotFound(c)
				return
			}
			res = *aux
		}

		tx.Commit()

		c.JSON(http.StatusOK, res)
	})
}

// GetPublishedFileName returns a published file name.
//
// GET /share/published/name/:hash
// Params:
// - hash: form.PublicFile.ShareHash
func GetPublishedFileName(router *gin.RouterGroup) {
	router.GET("/share/published/name/:hash", func(c *gin.Context) {
		//get file id from params
		hash := c.Param("hash")

		fileMutex := sync.Mutex{}
		fileMutex.Lock()
		defer fileMutex.Unlock()

		//get public file
		public_file, public_file_user_shared, err := query.FindPublicFileByHash(hash)
		if err != nil {
			log.Errorf("cannot get public file: %s", err)
			AbortEntityNotFound(c)
			return
		}
		// get the original file
		fileUid := ""

		if public_file != nil && public_file.ID != 0 {
			fileUid = public_file.FileUID
		} else if public_file_user_shared != nil && public_file_user_shared.ID != 0 {
			fileUid = public_file_user_shared.FileUID
		} else {
			log.Errorf("cannot get published file, published file is null: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//print public file and public file user shared
		f, err := query.FindFileByUID(fileUid)
		if err != nil {
			log.Errorf("cannot get published file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		res := CreateFileForSharedFile(*f, public_file, public_file_user_shared)

		c.JSON(http.StatusOK, res)
	})
}
