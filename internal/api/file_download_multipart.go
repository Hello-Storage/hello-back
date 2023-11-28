package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/pkg/s3"
	"github.com/gin-gonic/gin"
)

const ChunkSize = 5 * 1024 * 1024 // 5MB

// DownloadFile downloads file from filebase using s3
//
// GET /api/file/download/:uid
// @param uid path string true "file uid"
// @return 200 {string} string "ok"
func DownloadMultipartFile(router *gin.RouterGroup) {
	router.GET("/download/multipart/:uid", func(ctx *gin.Context) {
		// TO-DO check user auth & add user uid
		//authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		file_uid := ctx.Param("uid")

		// Multipart form
		//keyPath := authPayload.UserUID + "/" + file_uid
		keyPath := "/multipart/" + file_uid

		// Open a stream to the S3 object
		s3Service := *s3.NewS3Service(
			config.Env().WasabiAccessKey,
			config.Env().WasabiSecretKey,
			config.Env().WasabiRegion,
			config.Env().WasabiEndpoint,
		)

		reader, contentLength, contentType, err := s3Service.OpenStream(config.Env().WasabiBucket, keyPath)
		if err != nil {
			handleError(ctx, err)
			return
		}
		defer reader.Close()

		// Set headers
		ctx.Header("Content-Type", contentType)
		ctx.Header("Content-Length", fmt.Sprintf("%d", contentLength))
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file_uid))

		// Stream the file in chunks
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
	})

}

func handleError(ctx *gin.Context, err error) {
	if strings.Contains(err.Error(), "NoSuchKey") {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}
