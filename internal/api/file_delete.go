package api

import (
	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/s3"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gin-gonic/gin"
)

// DownloadFile download file from filebase s3
//
// DELETE /api/file/delete/:uid
//
// @param uid path string true "file uid"
// @return 200 {string} string "ok"

// DeleteFile adds delete route to delete file from database and S3 path style is force. It is used to delete files and delete their metadata
//
// @param router - Gin router to add
func DeleteFile(router *gin.RouterGroup) {
	router.DELETE("/delete/:uid", func(ctx *gin.Context) {
		// TO-DO check user auth & add user uid
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		file_uid := ctx.Param("uid")
		f, err := query.FindFileByUID(file_uid)

		// AbortEntityNotFound if file not found.
		if err != nil {
			AbortEntityNotFound(ctx)
			log.Errorf("file not found: %v", err)
			return
		}

		// Start the transaction
		tx := db.Db().Begin()

		// delete the file share state in case it exists
		query.DeleteFileShareState(db.Db(), f.UID)
		// delete the file share state user shared in case it exists
		err = query.DeleteFileShareStatesUserShared(db.Db(), f.UID, authPayload.UserID)

		if err != nil {
			tx.Rollback()
			log.Errorf("error deleting file share state user shared: %v", err)
			AbortInternalServerError(ctx)
			return
		}

		f_u := entity.FileUser{
			FileID: f.ID,
			UserID: authPayload.UserID,
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
		keyPath := f.CID

		// Check if other users have the file
		usersWithFile, err := query.FindUsersByFileCID(f.CID)
		// AbortInternalServerError if there is an error finding users by file CID
		if err != nil {
			tx.Rollback()
			AbortInternalServerError(ctx)
			log.Errorf("error finding users by file CID: %v", err)
			return
		}

		filesWithUser, err := query.FindFilesByUserAndFileCID(authPayload.UserID, f.CID)
		// AbortInternalServerError if there is an error finding files by user and file CID
		if err != nil {
			tx.Rollback()
			AbortInternalServerError(ctx)
			log.Errorf("error finding files by user and file CID: %v", err)
			return
		}

		// Delete the file from s3 if there is more than one user
		if len(usersWithFile) > 1 || len(filesWithUser) > 1 {
			// If more than one user has the file, delete the file from the user and give the owner to the next user shared
			// Returns true if the entity is a file owner.
			isOwner, err := entity.IsFileOwner(f_u.FileID, f_u.UserID)
			if err != nil {
				tx.Rollback()
				AbortInternalServerError(ctx)
				log.Errorf("error finding file owner: %v", err)
				return
			}
			if isOwner {
				// get the user details
				user_detail := query.FindUserDetailByUserID(authPayload.UserID)

				// removes storage_used from the database.
				newStorageUsed := user_detail.StorageUsed
				if uint(f.Size) <= user_detail.StorageUsed {
					newStorageUsed -= uint(f.Size)
				} else {
					log.Warnf("File size (%d) is larger than current storage used (%d), setting storage_used to 0", f.Size, user_detail.StorageUsed)
					newStorageUsed = 0
				}
				if err := user_detail.TxUpdate(tx, "storage_used", newStorageUsed); err != nil {
					log.Errorf("removing storage_used: %s", err)
				}

				// update the "deleted_at column" for this user
				if err := query.DeleteFileUser(tx, f_u); err != nil {
					tx.Rollback()
					AbortInternalServerError(ctx)
					log.Errorf("delete file user error: %v", err)
					return
				}

				// Delete file permission for the user
				if err := query.DeleteFilePermission(tx, f_u.UserID, f_u.FileID); err != nil {
					tx.Rollback()
					AbortInternalServerError(ctx)
					log.Errorf("delete file permission error: %v", err)
					return
				}

				//get id of the next owner
				nextOwner, err := query.GetNextOwner(f_u.UserID, f_u.FileID)
				if err != nil {
					log.Errorf("get next owner error: %v", err)
					log.Printf("length of files with user: %v", len(filesWithUser))
					if len(filesWithUser) > 1 {
						log.Printf("users with file bigger than 1")
						nextFileUser, err := query.GetNextFileUser(f_u.UserID, f.CID)
						if err != nil {
							tx.Rollback()
							log.Errorf("get next file error: %v", err)
							AbortInternalServerError(ctx)
							return
						} else {
							//give the owner
							query.SetOwnerPermision(tx, nextFileUser.UserID, nextFileUser.FileID)
							query.SetNextFileInPool(tx, nextFileUser.UserID, nextFileUser.FileID)
						}
					}
				} else {
					//give the owner
					query.SetOwnerPermision(tx, nextOwner.UserID, nextOwner.FileID)
					// get the user details
					user_detail_next := query.FindUserDetailByUserID(nextOwner.UserID)

					// sums storage_used from the database.
					if err := user_detail_next.TxUpdate(tx, "storage_used", user_detail_next.StorageUsed+uint(f.Size)); err != nil {
						log.Errorf("adding storage_used: %s", err)
					}
				}

			} else {
				// update the "deleted_at column" for this user
				if err := query.DeleteFileUser(tx, f_u); err != nil {
					tx.Rollback()
					AbortInternalServerError(ctx)
					log.Errorf("delete file user error: %v", err)
					return
				}

				// Delete file permission for the user
				if err := query.DeleteFilePermission(tx, f_u.UserID, f_u.FileID); err != nil {
					tx.Rollback()
					AbortInternalServerError(ctx)
					log.Errorf("delete file permission error: %v", err)
					return
				}

			}
		} else {
			// If not, delete the file from s3
			err = DeleteFileFromS3(keyPath, s3Config)

			// Delete the file from s3 if it exists
			if err != nil {
				log.Print("error deleting file from s3, trying using the CID. ")
				log.Print(err)
				log.Printf("deleting again")
				// If not found, try deleting using the CID as the keyPath
				keyPath = authPayload.UserUID + "/" + f.UID
				err = DeleteFileFromS3(keyPath, s3Config)
				if err != nil {
					log.Print("error deleting file from s3: ")
					log.Print(err)
				}
			}
			// get the user details
			user_detail := query.FindUserDetailByUserID(authPayload.UserID)

			// removes storage_used from the database.
			newStorageUsed := user_detail.StorageUsed
			if uint(f.Size) <= user_detail.StorageUsed {
				newStorageUsed -= uint(f.Size)
			} else {
				log.Warnf("File size (%d) is larger than current storage used (%d), setting storage_used to 0", f.Size, user_detail.StorageUsed)
				newStorageUsed = 0
			}
			if err := user_detail.Update("storage_used", newStorageUsed); err != nil {
				log.Errorf("removing storage_used: %s", err)
			}

			// update the "deleted_at column" for this user
			if err := query.DeleteFileUser(tx, f_u); err != nil {
				tx.Rollback()
				AbortInternalServerError(ctx)
				log.Errorf("delete file user error: %v", err)
				return
			}

			// Delete file permission for the user
			if err := query.DeleteFilePermission(tx, f_u.UserID, f_u.FileID); err != nil {
				tx.Rollback()
				AbortInternalServerError(ctx)
				log.Errorf("delete file permission error: %v", err)
				return
			}

		}

		tx.Commit()

		ctx.JSON(200, gin.H{
			"message": "ok",
		})
	})
}

// internal delete one file
// DeleteFileFromS3 delete file from s3. key must be unique in wasabi bucket. This is needed for file deletion to work
//
// @param keyPath - path to file in s3
// @param s3Config - aws. Config to use for delete
//
// @return error if something went wrong nil otherwise TODO ( jayconrod ) : Remove this once we have a way to delete files
func DeleteFileFromS3(keyPath string, s3Config aws.Config) error {

	if keyPath == "" {
		log.Errorf("DeleteFileFromS3: file uid is empty")
		return nil
	}

	//delete file from s3

	if err := s3.DeleteObject(s3Config, config.Env().WasabiBucket, keyPath); err != nil {
		log.Errorf("DeleteFileFromS3: delete file from s3 error: %v", err)
		return err
	}

	return nil
}

func SetNextFileInPool(user_uid uint, file_id uint) error {
	return db.Db().Table("files").
		Where("id = ?", file_id).
		Update("is_in_pool", false).
		Error
}

// Delete Duplicated Files That Aren't In Pool
//
// GET /delete-duplicated-files
func DeleteDuplicatedFiles() {
	tx := db.Db().Begin()

	// Get all files that are not in pool
	filesNotInPool, err := query.FindFilesNotInPool()
	if err != nil {
		log.Errorf("cannot get files: %s", err)
		return
	}

	totalFiles := len(filesNotInPool)
	log.Printf("Total files not in pool to delete: %d", totalFiles)

	// Delete all files that are not in pool
	for i, file := range filesNotInPool {
		fileId := file.ID
		fileUid := file.UID
		var fileUser entity.FileUser
		fileUser.FileID = fileId

		if err := tx.Model(&fileUser).Where("file_id = ?", fileId).Delete(&entity.FileUser{}).Error; err != nil {
			log.Errorf("cannot delete file user: %s", err)
			tx.Rollback()
			return
		}

		if err := tx.Model(&file).Where("id = ?", fileId).Delete(&entity.File{}).Error; err != nil {
			log.Errorf("cannot delete file: %s", err)
			tx.Rollback()
			return
		}

		var fileShareState entity.FileShareState
		if err := tx.Model(&fileShareState).Where("file_uid = ?", fileUid).Delete(&entity.FileShareState{}).Error; err != nil {
			log.Errorf("cannot delete file permission: %s", err)
			tx.Rollback()
			return
		}

		var publicFile entity.PublicFile
		if err := tx.Model(&publicFile).Where("file_uid = ?", fileUid).Delete(&entity.PublicFile{}).Error; err != nil {
			log.Errorf("cannot delete public file: %s", err)
			tx.Rollback()
			return
		}

		if i%10 == 0 || i == totalFiles-1 {
			log.Printf("Progress: %d of %d files deleted", i+1, totalFiles)
		}

	}

	tx.Commit()

	log.Printf("Deletion process completed. %d files processed.", totalFiles)

	//return length of files not in pool

}
