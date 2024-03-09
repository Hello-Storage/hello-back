package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/pkg/s3"
	"github.com/gin-gonic/gin"
)

// DownloadFolder downloads all files of a folder
//
// GET /api/folder/download/:uid
func DownloadMultipartFolder(router *gin.RouterGroup) {
	router.GET("/folder/download/multipart/:uid", func(ctx *gin.Context) {
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


		ctx.Writer.Header().Set("Content-Type", "multipart/mixed; boundary=boundary")

		for _, file := range allFiles {
			keyPath := file.CID
			
			// Stream each file
			streamFile(ctx, keyPath, file)

		}

		ctx.Writer.Write([]byte("\r\n--boundary--\r\n"))
	})

}

// streamFile streams a file from S3 as part of a multipart response
func streamFile(ctx *gin.Context, keyPath string, file entity.File) {
	// Open a stream to the S3 object
	s3Service := *s3.NewS3Service(
		config.Env().WasabiAccessKey,
		config.Env().WasabiSecretKey,
		config.Env().WasabiRegion,
		config.Env().WasabiEndpoint,
	)

	reader, _, contentType, err := s3Service.OpenStream(config.Env().WasabiBucket, keyPath)
	if err != nil {
		handleError(ctx, err)
		return
	}
	defer reader.Close()

	// Set headers
	ctx.Writer.Write([]byte("\r\n--boundary\r\n"))
	log.Printf("file name: %s", file.Name)
	ctx.Writer.Write([]byte(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", file.Name)))
	ctx.Writer.Write([]byte(fmt.Sprintf("Content-Type: %s\r\n\r\n", contentType)))

	// Stream the file in chunks
	//const ChunkSize = 5 * 1024 * 1024 // 5MB
	buffer := make([]byte, ChunkSize)
	for {
		bytesRead, readErr := reader.Read(buffer)
		if bytesRead > 0 {
			_, writeErr := ctx.Writer.Write(buffer[:bytesRead])
			if writeErr != nil {
				log.Errorf("cannot write to response: %s", writeErr)
				break
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			log.Errorf("cannot read from stream: %s", readErr)
			break
		}
	}
}