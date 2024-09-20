package api

import (
	"net/http"
	"time"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/s3"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gin-gonic/gin"
)

// CreateFile creates file based on the given file meta.
//
// POST /api/file/create
// - JSON customFileMeta that is returned by the frontend
func CreateFile(router *gin.RouterGroup) {
	router.POST("/create", func(ctx *gin.Context) {
		tx := db.Db().Begin()
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		var customFileMeta form.CustomFileMeta
		if err := ctx.ShouldBind(&customFileMeta); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//log entire request (not only file meta)

		// Check if file already exists
		/*
			var file entity.File
			err := db.UnscopedDb().Where("cid = ?", customFileMeta.CID).First(&file).Error
			if err == nil {
				ctx.JSON(http.StatusOK, gin.H{"file": file})
				return
			}
		*/
		//process file root
		root := customFileMeta.Root

		var r string
		if len(root) > 0 {
			r = root
		} else {
			r = "/"
		}

		actual_root, firstCreatedRootUID, err := GetAndProcessFileRoot(customFileMeta.Path, r, authPayload.UserID, customFileMeta.EncryptionStatus)

		if err != nil {
			log.Errorf("failed to process file root: %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var isInPool bool
		if customFileMeta.IsInPool != nil {
			isInPool = *customFileMeta.IsInPool
		} else {
			// Check if files exist in s3
			s3Config := aws.Config{
				Credentials: credentials.NewStaticCredentials(
					config.Env().StorageAccessKey,
					config.Env().StorageSecretKey,
					"",
				),
				Endpoint:         aws.String(config.Env().StorageEndpoint),
				Region:           aws.String(config.Env().StorageRegion),
				S3ForcePathStyle: aws.Bool(true),
			}
			_, err := s3.HeadObject(s3Config, config.Env().StorageBucket, customFileMeta.CID)
			if err != nil {
				log.Printf("error: %s", err)
				//this means that the object doesn't exist at S3
				isInPool = false
			} else {
				usersWithFile, err := query.FindUsersByFileCID(customFileMeta.CID)
				if err != nil {
					AbortInternalServerError(ctx)
					log.Errorf("error finding users by file CID: %v", err)
					return
				}
				if len(usersWithFile) > 0 {
					isInPool = true
				} else {
					isInPool = false
				}
			}
		}
		// Create file
		file := entity.File{
			Name:                 customFileMeta.Name,
			Root:                 actual_root,
			CID:                  customFileMeta.CID,
			CIDOriginalEncrypted: &customFileMeta.CIDOriginalEncrypted,
			Mime:                 customFileMeta.MimeType,
			Size:                 customFileMeta.Size,
			EncryptionStatus:     customFileMeta.EncryptionStatus,
			IsInPool:             &isInPool,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := file.TxCreate(tx); err != nil {
			log.Errorf("failed to create file: %s", err)
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//if is in pool, permission is sharedpermission
		//otherwise, ownerpermission
		var f_u entity.FileUser
		if isInPool {
			//create file share state
			f_u = entity.FileUser{
				FileID:     file.ID,
				UserID:     authPayload.UserID,
				Permission: entity.SharedPermission,
			}
		} else {
			f_u = entity.FileUser{
				FileID:     file.ID,
				UserID:     authPayload.UserID,
				Permission: entity.OwnerPermission,
			}
		}

		if err := f_u.TxCreate(tx); err != nil {
			log.Errorf("failed to create file_user: %s", err)
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// add user storage quantity
		user_detail := query.FindUserDetailByUserID(authPayload.UserID)

		if err := user_detail.TxUpdate(tx, "storage_used", user_detail.StorageUsed+uint(file.Size)); err != nil {
			log.Errorf("adding storage_used: %s", err)
			AbortInternalServerError(ctx)
			return
		}
		tx.Commit()

		tx.Commit()

		fResponse := form.FileResponse{
			ID:        file.ID,
			UID:       file.UID,
			IsInPool:  file.IsInPool,
			CreatedAt: file.CreatedAt.String(),
			UpdatedAt: file.UpdatedAt.String(),
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":       "success",
			"firstRootUID": firstCreatedRootUID,
			"fileCreated":  fResponse,
		})
	})
}
