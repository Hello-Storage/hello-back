package v1

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gin-gonic/gin"
)

func DeleteFile(router *gin.RouterGroup) {
	router.DELETE("/files/:uid", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.APIKeyHeaderKey).(*token.Payload)

		//increment requests counter
		apiKey, err := query.FindApiKeyByUserID(authPayload.UserID)
		if err == nil && apiKey != nil {
			apiKey.IncrementKeyRequests()
		}

		uid := ctx.Param("uid")
		f, err := query.FindFileByUID(uid)

		// AbortEntityNotFound if file not found.
		if err != nil {
			api.AbortEntityNotFound(ctx)
			fmt.Printf("file not found: %v", err)
			return
		}

		tx := db.Db().Begin()

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

		inicio := time.Now()

		keyPath := f.CID + strconv.FormatUint(uint64(authPayload.UserID), 10)

		fmt.Printf("Attempting to delete file from S3. CID: %s\n", f.CID)
		error := api.DeleteFileFromS3(keyPath, s3Config)
		fmt.Printf("Delete operation completed. Elapsed time: %s\n", time.Since(inicio))
		fmt.Println(err)
		// Delete the file from s3 if it exists
		if error != nil {
			fmt.Print("error deleting file from s3 using the CID. ")
			fmt.Print(error)
		}
		// get the user details
		user_detail := query.FindUserDetailByUserID(authPayload.UserID)

		// removes storage_used from the database.
		updatedStorageUsed := uint(0)
		if uint(f.Size) < (user_detail.StorageUsed) {
			updatedStorageUsed = user_detail.StorageUsed - uint(f.Size)
		}

		fmt.Printf("Updating storage_used: Old Value=%d, Size to Remove=%d, New Value=%d", user_detail.StorageUsed, f.Size, updatedStorageUsed)

		if err := user_detail.TxUpdate(tx, "storage_used", updatedStorageUsed); err != nil {
			fmt.Printf("Error updating storage_used: %s", err)
		}

		// update the "deleted_at column" for this user
		if err := query.DeleteFileUser(tx, f_u); err != nil {
			tx.Rollback()
			api.AbortInternalServerError(ctx)
			fmt.Printf("delete file user error: %v", err)
			return
		}

		// Delete the apikey file
		if err := query.DeleteApiKeyFileByFileID(tx, f_u.FileID); err != nil {
			tx.Rollback()
			api.AbortInternalServerError(ctx)
			fmt.Printf("delete api key file error: %v", err)
			return
		}

		// Delete file permission for the user
		if err := query.DeleteFilePermission(tx, f_u.UserID, f_u.FileID); err != nil {
			tx.Rollback()
			api.AbortInternalServerError(ctx)
			fmt.Printf("delete file permission error: %v", err)
			return
		}

		tx.Commit()

		ctx.JSON(200, gin.H{
			"message": "ok",
		})
	})
}
