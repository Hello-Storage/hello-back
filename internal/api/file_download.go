package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/pkg/s3"
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
		//authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		file_uid := ctx.Param("uid")

		// Multipart form
		//keyPath := authPayload.UserUID + "/" + file_uid
		keyPath := "/multipart/" + file_uid
		out, error := DownloadFileFromS3(keyPath)
		//if error contains "NoSuchKey" then set keyPath without the userUID
		if error != nil {
			if strings.Contains(error.Error(), "NoSuchKey") {
				//get file by uid
				/*
					f, err := query.FindFileByUID(file_uid)
					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{
							"message": err.Error(),
						})
						return
					}
				*/
				keyPath = "/multipart/" + file_uid
				out, error = DownloadFileFromS3(keyPath)
				log.Errorf("download file: %s", error)
				if error != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{
						"message": "Key: " + keyPath + "Error: " + error.Error(),
					})
					return
				}

			} else {
				log.Errorf("download file: %s", error)
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": error.Error(),
				})
				return
			}

		}
		// Set the correct content type and file name
		ctx.Header("Content-Type", *out.ContentType)
		log.Printf("Content-Type: %s", *out.ContentType)
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file_uid))

		// Copy the file data to the response
		_, error = io.Copy(ctx.Writer, out.Body)
		if error != nil {
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
