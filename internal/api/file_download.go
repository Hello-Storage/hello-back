package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/s3"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

// DownloadFile downloads file from filebase using s3
//
// GET /api/file/download/:uid
// @param uid path string true "file uid"
// @return 200 {string} string "ok"
func DownloadFile(router *gin.RouterGroup) {
	router.GET("/download/:uid", func(ctx *gin.Context) {
		// TO-DO check user auth & add user uid
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		file_uid := ctx.Param("uid")

		// Multipart form
		keyPath := authPayload.UserUID + "/" + file_uid
		out, error := DownloadFileFromS3(keyPath)
		//if error contains "NoSuchKey" then set keyPath without the userUID
		if error != nil {
			if strings.Contains(error.Error(), "NoSuchKey") {
				//get file by uid
				f, _ := query.FindFileByUID(file_uid)
				keyPath = f.CID
				out, error = DownloadFileFromS3(keyPath)

				//if error contains "NoSuchKey" then set keyPath without the userUID
				if error != nil {
					if strings.Contains(error.Error(), "NoSuchKey") {
						keyPath = f.CID + strconv.FormatUint(uint64(authPayload.UserID), 10)
						out, error = DownloadFileFromS3(keyPath)
						log.Errorf("download file: %s", error)
						if error != nil {
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": error.Error(),
							})
							return
						}

					} else {
						log.Errorf("download file specific error: %v", error)
						ctx.JSON(http.StatusBadRequest, gin.H{
							"message": error.Error(),
						})
						return
					}

				}

			} else {
				log.Errorf("download file general error: %s", error.Error())
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": error.Error(),
				})
				return
			}

		}
		// Set the correct content type and file name
		ctx.Header("Content-Type", *out.ContentType)
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file_uid))

		// Copy the file data to the response
		_, error = io.Copy(ctx.Writer, out.Body)
		if error != nil {
			log.Errorf("download file: %s", error)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": error.Error(),
			})
		}

	})
}

// internal upload one file
func DownloadFileFromS3(key string) (*awsS3.GetObjectOutput, error) {
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

	out, err := s3.DownloadObject(s3Config, config.Env().WasabiBucket, key)

	return out, err
}
