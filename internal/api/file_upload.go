package api

import (
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

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

// UploadFiles upload files to wasabi using s3
//
// POST /api/file/upload
// Form: MultipartForm
// - files
// - root
func PutUploadFiles(router *gin.RouterGroup) {
	router.POST("/upload", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		// Multipart form
		form, err := ctx.MultipartForm()

		if err != nil {
			log.Errorf("multipart form: %s", err)
			AbortBadRequest(ctx)
			return
		}

		root := form.Value["root"][0]
		files := form.File["files"]
		status := form.Value["status"][0]

		log.Infof("status: %s", status)

		// Handle regular files
		for _, file := range files {

			_, params, err := mime.ParseMediaType(file.Header.Get("Content-Disposition"))
			if err != nil {
				log.Errorf("parse media type: %s", err)
				AbortInternalServerError(ctx)
				return
			}
			mime := file.Header.Get("Content-Type")

			// create corresponding folders to locate this file at proper path
			file_path := params["filename"]
			log.Infof("file_path, %v", file_path)
			actual_root, err := GetAndProcessFileRoot(
				file_path,
				root,
				authPayload.UserID,
				entity.EncryptionStatus(status),
			)
			log.Infof("actual_root: %s", actual_root)

			// create file
			f := entity.File{
				Name:   file_path[strings.LastIndex(file_path, "/")+1:],
				Root:   actual_root,
				Mime:   mime,
				Size:   file.Size,
				Status: entity.EncryptionStatus(status),
			}
			if err := f.Create(); err != nil {
				log.Errorf("create file: %s", err)
				AbortInternalServerError(ctx)
			}

			// create file_user relation
			f_u := entity.FileUser{
				FileID:     f.ID,
				UserID:     authPayload.UserID,
				Permission: entity.OwnerPermission,
			}
			if err := f_u.Create(); err != nil {
				AbortInternalServerError(ctx)
				return
			}

			keyPath := authPayload.UserUID + "/" + f.UID
			// upload file
			if err := UploadFileToS3(file, keyPath); err != nil {
				log.Errorf("uploading file to s3: %s", err)
				AbortInternalServerError(ctx)
				return
			}

			// add user storage quantity
			user_detail := query.FindUserDetailByUserID(authPayload.UserID)

			if err := user_detail.Update("storage_used", user_detail.StorageUsed+uint(file.Size)); err != nil {
				log.Errorf("adding storage_used: %s", err)
				AbortInternalServerError(ctx)
				return
			}

		}

		ctx.JSON(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	})
}

// internal upload one file
func UploadFileToS3(file *multipart.FileHeader, key string) error {

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

	err := s3.UploadObject(s3Config, file, config.Env().WasabiBucket, key)

	return err
}

// internal
// here root => uid format
func GetAndProcessFileRoot(
	file_path, root string,
	user_id uint,
	status entity.EncryptionStatus,
) (string, error) {
	res := strings.Split(file_path, "/")
	if len(res) == 1 {
		return root, nil
	}

	sub_file_path := strings.Join(res[1:], "/")
	sub_title := res[0]

	f := query.FindFolderByTitleAndRoot(sub_title, root)

	log.Infof("folder find by title and root: %v", f)
	if f == nil {
		f = &entity.Folder{
			Title:  sub_title,
			Root:   root,
			Status: status,
		}

		if err := f.Create(); err != nil {
			return "", errors.New("can't create folder")
		}
		// create folder_user relation
		f_u := &entity.FolderUser{
			FolderID:   f.ID,
			UserID:     user_id,
			Permission: entity.OwnerPermission,
		}

		if err := f_u.Create(); err != nil {
			return "", errors.New("can't create folder_user relation")
		}
	}

	return GetAndProcessFileRoot(sub_file_path, f.UID, user_id, status)
}
