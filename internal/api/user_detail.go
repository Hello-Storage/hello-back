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

		// get user files from the table "file_user" (where permission != deleted)
		filesUser, err := query.GetFilesUserFromUser(user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching user files"})
			return
		}

		// start variables for shared files
		var sharedWithUser entity.Files
		var sharedByUser entity.Files

		// iterate over files
		for _, fileUser := range filesUser {
			// try to get file by its id
			file, err := query.FindFileByID(fileUser.FileID)
			if err != nil {
				log.Errorf("error fetching file: %v", err)
				if err != gorm.ErrRecordNotFound {
					// if it's not a "not found" error it means it's probably an internal error, so stop here
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching file"})
					return
				}
				// if it's a "not found" error, continue with the next file
				continue
			}

			// get users with file CID
			usersWithFile, err := query.FindUsersByFileCID(file.CID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching users with file"})
				return
			}

			// filter out usersID that belong to others than the current user
			usersWithFileFiltered := []uint{}
			for _, usrID := range usersWithFile {
				if usrID != user.ID {
					usersWithFileFiltered = append(usersWithFileFiltered, usrID)
				}
			}

			if file.ID != 0 {
				if fileUser.Permission == entity.SharedPermission && len(usersWithFileFiltered) > 0 {
					// if the file is shared to current user and its more than one user with the file
					// then it's a shared (with the user) file
					if file.Root == "/" /*|| !query.IsInSharedFolder(file.Root, authPayload.UserID) */{
						// if the file was shared by email or wallet, we need to get
						// the public file and its share state
						sharestatefound, err := query.GetFileShareStateByFileUIDAndUserID(file.UID, authPayload.UserID)
						if err == nil {
							file.FileShareState = query.ConvertToDomainEntities(sharestatefound)
						}
						sharedWithUser = append(sharedWithUser, *file)
					}
				} else if fileUser.Permission == entity.OwnerPermission && len(usersWithFileFiltered) > 0 {
					// if the file owned by current user and its more than one user with the file
					// then it's a shared (by the user) file
					if file.Root == "/" /*|| !query.IsInSharedFolder(file.Root, authPayload.UserID) */{
						// TODO: check if the file/forder is in a shared folder or not 
						// (because if we only show the files in Root, the elements in a non-shared folder will not be shown)
						sharedByUser = append(sharedByUser, *file)
					}
				}
			}
		}

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
