package api

import (
	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
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

func DeleteFile(router *gin.RouterGroup) {
	router.DELETE("/delete/:uid", func(ctx *gin.Context) {
		// TO-DO check user auth & add user uid
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		file_uid := ctx.Param("uid")

		f, err := query.FindFileByUID(file_uid)

		if err != nil {
			AbortEntityNotFound(ctx)
			log.Errorf("file not found: %v", err)
			return
		}

		// Check if other users have the file
		// If not, delete the file from s3
		usersWithFile, err := query.FindUsersByFileCID(f.CID)
		if err != nil {
			AbortInternalServerError(ctx)
			log.Errorf("error finding users by file CID: %v", err)
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
		keyPath := authPayload.UserUID + "/" + f.UID

		// If more than one user has the file, delete the file from the database

		if len(usersWithFile) > 1 {

		} else {
			// Try deleting using the userUID/fileUID keyPath
			err = DeleteFileFromS3(keyPath, s3Config)

			if err != nil {
				log.Print("error deleting file from s3, trying using the CID. ")
				log.Print(err)
				// If not found, try deleting using the CID as the keyPath
				keyPath = f.CID
				err = DeleteFileFromS3(keyPath, s3Config)
				if err != nil {
					log.Print("error deleting file from s3: ")
					log.Print(err)
				}
			}
			// remove user storage quantity
			user_detail := query.FindUserDetailByUserID(authPayload.UserID)

			if err := user_detail.Update("storage_used", user_detail.StorageUsed-uint(f.Size)); err != nil {
				log.Errorf("removing storage_used: %s", err)
				AbortInternalServerError(ctx)
				return
			}

		}

		// Delete file user
		if err := query.DeleteFileUser(f_u); err != nil {
			AbortInternalServerError(ctx)
			log.Errorf("delete file user error: %v", err)
			return
		}

		//Delete file
		if err := query.DeleteFileByUID(f.UID); err != nil {
			AbortInternalServerError(ctx)
			log.Errorf("delete file error: %v", err)
			return
		}

		ctx.JSON(200, gin.H{
			"message": "ok",
		})
	})
}

// internal delete one file
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
