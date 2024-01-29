package api

import (
	"fmt"
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gin-gonic/gin"
)

// DeleteFolder deletes the folder eand its associated files.
//
// DELETE /api/folder/delete/:uid
//
// @param uid path string true "folder uid"
// @return 200 {string} string "ok"

func DeleteFolder(router *gin.RouterGroup) {
	router.DELETE("/folder/:uid", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		folderUID := ctx.Param("uid")

		// Find folder by UID
		folder, err := query.FindFolderByUID(folderUID)
		if err != nil {
			AbortEntityNotFound(ctx)
			return
		}

		// Check user permission (ownership in this case)
		folderUser, err := query.FindFolderUser(folder.ID, authPayload.UserID)
		if err != nil || (folderUser.Permission != entity.SharedPermission && folderUser.Permission != entity.OwnerPermission) {
			fmt.Printf("folder find user: %s", err)
			ctx.JSON(http.StatusForbidden, gin.H{
				"message": "Permission denied",
			})
			return
		}

		// Delete folder and its contents recursively
		if err := DeleteFolderAndContentsRecursive(folderUID, authPayload.UserUID, authPayload.UserID); err != nil {
			fmt.Printf("folder delete contents recursive: %s", err)
			AbortInternalServerError(ctx)
			return
		}

		// Update user storage metrics
		// Implement as per your logic

		ctx.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})
}

func DeleteFolderAndContentsRecursive(folderUID, userUID string, userID uint) error {
	// Step 1: Delete all files in the folder
	if err := DeleteAllFilesInFolder(folderUID, userUID, userID); err != nil {
		fmt.Println("Error deleting files in folder: ", err)
		return err
	}

	// Step 2: Get all child folders
	childFolders, err := query.GetChildFoldersByUID(folderUID)
	if err != nil {
		return err
	}

	// Step 3: Recursively delete all child folders
	for _, childFolder := range childFolders {
		if err := DeleteFolderAndContentsRecursive(childFolder.UID, userUID, userID); err != nil {
			return err
		}
	}

	// Step 4: Delete the folder itself
	if err := query.DeleteFolderByUID(folderUID); err != nil {
		return err
	}

	return nil
}

// DeleteAllFilesInFolder deletes all files in a folder and its child folders (if any).
func DeleteAllFilesInFolder(folderUID, userUID string, userID uint) error {
	//Logic to delete all files in a folder and its child folders (if any)
	// This involves:
	// 1. Query all files in the folder
	var files []entity.File
	files, err := query.GetFolderFilesByUID(folderUID)
	if err != nil {
		return err
	}

	s3Config := aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.Env().WasabiAccessKey,
			config.Env().WasabiSecretKey,
			"",
		),
		Endpoint:         aws.String(config.Env().WasabiEndpoint),
		Region:           aws.String(config.Env().WasabiRegion),
		S3ForcePathStyle: aws.Bool(true),
	}

	// 2. Delete each file from S3
	for _, file := range files {

		f_u := entity.FileUser{
			FileID: file.ID,
			UserID: userID,
		}

		// delete the file share state in case it exists
		query.DeleteFileShareState(file.UID)
		// delete the file share state user shared in case it exists
		query.DeleteFileShareStatesUserShared(file.UID, userID)

		// Check if other users have the file
		usersWithFile, err := query.FindUsersByFileCID(file.CID)
		if err != nil {
			return err
		}

		keyPath := file.CID

		filesWithUser, err := query.FindFilesByUserAndFileCID(userID, file.CID)

		if err != nil {
			return err
		}

		// update the "deleted_at column" for this user
		if err := query.DeleteFileUser(f_u); err != nil {
			return err
		}

		if len(usersWithFile) > 1 || len(filesWithUser) > 1 {
			// If more than one user has the file, delete the file from the user and give the owner to the next user shared
			// Returns true if the entity is a file owner.

			isOwner, err := entity.IsFileOwner(file.ID, userID)
			fmt.Println("isOwner: ", isOwner)
			if err != nil {
				return err
			}

			if isOwner {
				// get the user details
				user_detail := query.FindUserDetailByUserID(userID)

				// removes storage_used from the database.
				newStorageUsed := user_detail.StorageUsed
				if uint(file.Size) <= user_detail.StorageUsed {
					newStorageUsed -= uint(file.Size)
				} else {
					newStorageUsed = 0
				}
				if err := user_detail.Update("storage_used", newStorageUsed); err != nil {
					log.Errorf("removing storage_used: %s", err)
				}

				// Delete file permission for the user
				if err := query.DeleteFilePermission(userID, f_u.FileID); err != nil {
					return err
				}

				//get id of the next owner
				nextOwner, err := query.GetNextOwner(userID, f_u.FileID)
				if err != nil {
					log.Errorf("get next owner error: %v", err)
					log.Printf("length of files with user: %v", len(filesWithUser))
					if len(filesWithUser) > 1 {
						nextFileUser, err := query.GetNextFileUser(userID, file.CID)
						if err != nil {
							return err
						} else {
							//give the owner
							log.Printf("next file user: %v", nextFileUser)
							query.SetOwnerPermision(nextFileUser.UserID, nextFileUser.FileID)
							query.SetNextFileInPool(nextFileUser.UserID, nextFileUser.FileID)
						}
					}
				} else {
					//give the owner
					query.SetOwnerPermision(nextOwner.UserID, nextOwner.FileID)
					// get the user details
					user_detail_next := query.FindUserDetailByUserID(nextOwner.UserID)
					// sums storage_used from the database.
					if err := user_detail_next.Update("storage_used", user_detail_next.StorageUsed+uint(file.Size)); err != nil {
						log.Errorf("adding storage_used: %s", err)
					}
				}
			} else {
				// Delete file permission for the user
				if err := query.DeleteFilePermission(f_u.UserID, f_u.FileID); err != nil {
					return err
				}
			}

		} else {
			// If not, delete the file from s3
			err = DeleteFileFromS3(keyPath, s3Config)

			// Delete the file from s3 if it exists
			if err != nil {
				keyPath = userUID + "/" + file.UID
				err = DeleteFileFromS3(keyPath, s3Config)
				if err != nil {
					log.Print("error deleting file from s3: ")
					log.Print(err)
				}
			}
			// get the user details
			user_detail := query.FindUserDetailByUserID(userID)

			// removes storage_used from the database.
			newStorageUsed := user_detail.StorageUsed
			if uint(file.Size) <= user_detail.StorageUsed {
				newStorageUsed -= uint(file.Size)
			} else {
				newStorageUsed = 0
			}
			if err := user_detail.Update("storage_used", newStorageUsed); err != nil {
				log.Errorf("removing storage_used: %s", err)
			}

			// update the "deleted_at column" for this user
			if err != nil {
				return err
			}

			// Delete file permission for the user
			if err := query.DeleteFilePermission(userID, file.ID); err != nil {
				return err
			}

		}
	}
	return nil
}
