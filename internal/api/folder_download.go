package api

import (
	"encoding/base64"
	"io"
	"net/http"
	"path"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/gin-gonic/gin"
)

func getAllFiles(folderUID string, allFiles *[]entity.File, currentPath string) error {
	// Find the folder by UID
	var folder entity.Folder
	if err := db.Db().Where("uid = ?", folderUID).Find(&folder).Error; err != nil {
		return err
	}

	// Concatenate current path with the folder title
	currentPath = path.Join(currentPath, folder.Title)

	// Find all files inside the folder
	var files []entity.File
	if err := db.Db().Where("root = ?", folderUID).Find(&files).Error; err != nil {
		return err
	}

	// Concatenate current path with the folder title
	for i := range files {
		files[i].Path = path.Join(currentPath, files[i].Name)
	}

	*allFiles = append(*allFiles, files...)

	// Find all child folders
	var childFolders []entity.Folder
	if err := db.Db().Where("root = ?", folderUID).Find(&childFolders).Error; err != nil {
		return err
	}

	// Recursively get files from child folders
	for _, childFolder := range childFolders {
		if err := getAllFiles(childFolder.UID, allFiles, currentPath); err != nil {
			return err
		}
	}

	return nil
}

// GetFolderFiles returns all files of a folder
//
// GET /api/folder/files/:uid
func GetFolderFiles(router *gin.RouterGroup) {
	router.GET("/folder/files/:uid", func(ctx *gin.Context) {
		folderUID := ctx.Param("uid")
		log.Printf("folderUID: %s", folderUID)

		// Find all files inside the folder
		var allFiles []entity.File
		if err := getAllFiles(folderUID, &allFiles, ""); err != nil {
			log.Errorf("error getting all files: %s", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Unable to retrieve files",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "success",
			"files":   allFiles,
		})
	})

}
// DownloadFolder downloads all files of a folder as a ZIP
//
// GET /api/folder/download/:uid
func DownloadFolder(router *gin.RouterGroup) {
	router.GET("/folder/download/:uid", func(ctx *gin.Context) {
		// authPayload no needed for now
		// authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		folderUID := ctx.Param("uid")

		// Find the folder by UID
		var folder entity.Folder
		if err := db.Db().Where("uid = ?", folderUID).First(&folder).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Folder not found",
			})
			return
		}

		// Find all files inside the folder
		var allFiles []entity.File
		if err := getAllFiles(folderUID, &allFiles, ""); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Unable to retrieve files",
			})
			return
		}

		fileData := make([]map[string]interface{}, len(allFiles))

		for i, file := range allFiles {
			keyPath := file.CID // keypath is the CID of the file
			out, err := DownloadFileFromS3(keyPath)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}

			// Read the body into bytes
			bodyBytes, err := io.ReadAll(out.Body)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}

			// Encode the bytes as base64
			base64Data := base64.StdEncoding.EncodeToString(bodyBytes)

			fileData[i] = map[string]interface{}{
				"name":       file.Name,
				"data":       base64Data,
				"mime":       file.Mime,
				"size":       file.Size,
				"uid":        file.UID,
				"root":       file.Root,
				"date":       file.CreatedAt,
				"media_type": file.MediaType,
				"mime_type":  file.Mime,
				"updated_at": file.UpdatedAt,
				"path": file.Path,
				"encryption_status": file.EncryptionStatus,
				"cid": file.CID,
				"cid_original_encrypted": file.CIDOriginalEncrypted,
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "success",
			"files":   fileData,
		})

	})

}
