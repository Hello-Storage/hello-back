package api

import (
	"fmt"
	"net/http"

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
		if err != nil || folderUser.Permission != entity.OwnerPermission {
			fmt.Printf("folder find user: %s", err)
			ctx.JSON(http.StatusForbidden, gin.H{
				"message": "Permission denied",
			})
			return
		}

		// Delete folder and its contents recursively
		if err := DeleteFolderAndContentsRecursive(folderUID, authPayload.UserUID); err != nil {
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

func DeleteFolderAndContentsRecursive(folderUID, userUID string) error {
	// Step 1: Delete all files in the folder
	if err := DeleteAllFilesInFolder(folderUID, userUID); err != nil {
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
		if err := DeleteFolderAndContentsRecursive(childFolder.UID, userUID); err != nil {
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
func DeleteAllFilesInFolder(folderUID, userUID string) error {
	//Logic to delete all files in a folder and its child folders (if any)
	// This involves:
	// 1. Query all files in the folder
	var files []entity.File
	files, err := query.GetFolderFilesByUID(folderUID)
	if err != nil {
		return err
	}
	// 2. Delete each file from S3
	for _, file := range files {

		// Delete from S3
		keyPath := userUID + "/" + file.UID

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

		_, err = s3.HeadObject(s3Config, config.Env().WasabiBucket, keyPath)
		if err != nil {
			keyPath = file.CID
		}

		if !*file.IsInPool {
			if err := DeleteFileFromS3(keyPath, s3Config); err != nil {
				return err
			}
		}

		// Delete from DB
		if err := query.DeleteFileByUID(file.UID); err != nil {
			return err
		}

		// add user storage quantity
		user_detail := query.FindUserDetailByUserUID(userUID)

		if err := user_detail.Update("storage_used", user_detail.StorageUsed-uint(file.Size)); err != nil {
			log.Errorf("adding storage_used: %s", err)
		}
	}
	return nil
}
