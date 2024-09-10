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

type SharedNode struct {
	Files   entity.Files
	Folders entity.Folders
}

type SharedListUser struct {
	SharedWithMe SharedNode
	SharedByMe   SharedNode
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

		// get user from the db to check if it exists
		user := query.FindUser(entity.User{ID: authPayload.UserID})
		if user == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// get filesUser from the table "file_user" (where permission != deleted)
		filesUser, err := query.GetFilesUserFromUser(user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching user files"})
			return
		}

		var sharedWithUser entity.Files
		var sharedByUser entity.Files
		sharedByUserMap := make(map[uint]struct{})
		sharedWithUserMap := make(map[string]struct{})

		// Create a channel to fetch files concurrently
		type fileResult struct {
			file *entity.File
			err  error
		}
		fileChan := make(chan fileResult, len(filesUser))

		// Fetch files concurrently using goroutines
		for _, fileUser := range filesUser {
			go func(fileUser entity.FileUser) {
				file, err := query.FindFileByID(fileUser.FileID)
				fileChan <- fileResult{file: file, err: err}
			}(fileUser)
		}

		for _, fileUser := range filesUser {
			fileResult := <-fileChan
			if fileResult.err != nil {
				log.Errorf("error fetching file: %v", fileResult.err)
				if fileResult.err != gorm.ErrRecordNotFound {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching file"})
					return
				}
				continue
			}

			file := fileResult.file

			// Filter files shared by the user
			if file.FileShareState.ID != 0 && fileUser.Permission == entity.OwnerPermission {
				if _, exists := sharedByUserMap[file.ID]; exists {
					continue
				}
				sharedByUser = append(sharedByUser, *file)
				sharedByUserMap[file.ID] = struct{}{}
			}

			// Check if the file CID has already been processed
			if _, exists := sharedWithUserMap[file.CID]; exists {
				continue // Skip processing the file if this CID has already been processed
			}

			// Concurrently fetch users with the file CID
			usersWithFile, err := query.FindUsersByFileCID(file.CID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching users with file"})
				return
			}

			// Filter out usersID that belong to others than the current user
			usersWithFileFiltered := make([]uint, 0, len(usersWithFile))
			for _, usrID := range usersWithFile {
				if usrID != user.ID {
					usersWithFileFiltered = append(usersWithFileFiltered, usrID)
				}
			}

			if file.ID != 0 && fileUser.Permission == entity.SharedPermission && len(usersWithFileFiltered) > 0 {
				if file.Root == "/" {
					sharestatefound, err := query.GetFileShareStateByFileUIDAndUserID(file.UID, authPayload.UserID)
					if err == nil {
						file.FileShareState = query.ConvertToDomainEntities(sharestatefound)
					}

					sharedWithUser = append(sharedWithUser, *file)
					// Mark this CID as processed to avoid duplicate entries
					sharedWithUserMap[file.CID] = struct{}{}
				}
			}
		}

		// Close the channel after processing
		close(fileChan)

		// at this point we have all the shared files, now we need to get the folders

		// get user folders from the table "folder_user"
		foldersUser, err := query.GetFoldersUserFromUser(user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching user Folders"})
			return
		}

		// start variables for shared folders
		var FoldersharedwithUser entity.Folders
		var FoldersharedByUser entity.Folders

		// iterate over folders
		for _, folderUser := range foldersUser {
			// try to get folder by its id
			folder, err := query.FindFolderByID(folderUser.FolderID)
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					// if it's not a "not found" error it means it's probably an internal error, so stop here
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching file"})
					return
				}
				// if it's a "not found" error, continue with the next folder
				continue
			}

			// get users with folder CID
			usersWithFolder, err := query.FindUsersByFolderCID(folder.CID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching users with file"})
				return
			}

			// filter out usersID that belong to others than the current user}
			usersWithFolderFiltered := []uint{}
			for _, usrID := range usersWithFolder {
				if usrID != user.ID {
					usersWithFolderFiltered = append(usersWithFolderFiltered, usrID)
				}
			}

			// if the folder is shared to current user and its more than one user with the folder
			// then it's a shared (with the user) folder
			if folder.ID != 0 {
				if folderUser.Permission == entity.SharedPermission && len(usersWithFolderFiltered) > 0 {
					if folder.Root == "/" /*|| !query.IsInSharedFolder(folder.Root, authPayload.UserID)*/ {
						// TODO: check if the file/forder is in a shared folder or not
						// (because if we only show the files in Root, the elements in a non-shared folder will not be shown)
						FoldersharedwithUser = append(FoldersharedwithUser, *folder)
					}
				} else if folderUser.Permission == entity.OwnerPermission && len(usersWithFolderFiltered) > 0 {
					if folder.Root == "/" /*|| !query.IsInSharedFolder(folder.Root, authPayload.UserID)*/ {
						FoldersharedByUser = append(FoldersharedByUser, *folder)
					}
				}
			}
		}

		response := SharedListUser{
			SharedWithMe: SharedNode{
				Files:   sharedWithUser,
				Folders: FoldersharedwithUser,
			},
			SharedByMe: SharedNode{
				Files:   sharedByUser,
				Folders: FoldersharedByUser,
			},
		}

		ctx.JSON(http.StatusOK, response)
	})

}
