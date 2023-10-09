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
	type FileResponse struct {
		Name string `json:"name"`
		UID string `json:"uid"`
		Root string `json:"root"`
		CID string `json:"cid"`
		Mime string `json:"mime"`
		Size int64 `json:"size"`
		EnryptionStatus entity.EncryptionStatus `json:"encryption_status"`
	}	

	var fileResponses []FileResponse

	router.POST("/upload", func(ctx *gin.Context) {
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)

		// Multipart form
		form, err := ctx.MultipartForm()

		if err != nil {
			log.Errorf("multipart form: %s", err)
			AbortBadRequest(ctx)
			return
		}

		root := form.Value["root"]
		files := form.File["files"]


		var r string
		if len(root) > 0 {
			r = root[0]
		} else {
			r = "/"
		}

		// Handle regular files
		for index, file := range files {

			index := fmt.Sprintf("%d", index)

			//cid of file
			cid, ok := form.Value["cid["+index+"]"]
			if !ok || len(cid) == 0 {
				log.Warnf("Missing or empty cid for index %s", index)
				continue
			}

			_, params, err := mime.ParseMediaType(file.Header.Get("Content-Disposition"))
			if err != nil {
				log.Errorf("parse media type: %s", err)
				AbortInternalServerError(ctx)
				return
			}
			mime := file.Header.Get("Content-Type")

			// create corresponding folders to locate this file at proper path
			file_path := params["filename"]
			actual_root, err := GetAndProcessFileRoot(file_path, r, authPayload.UserID, entity.Public)

			// create file
			f := entity.File{
				Name:   file.Filename,
				Root:   actual_root,
				CID:    cid[0],
				Mime:   mime,
				Size:   file.Size,
				EncryptionStatus: entity.Public,
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
				log.Errorf("create file_user relation: %s", err)
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

		encryptedFiles := form.File["encryptedFiles"]

		for key, encryptedFile := range encryptedFiles {
			// Ensure the key exists and has values

			index := fmt.Sprintf("%d", key)

			//cid of encrypted buffer
			cid, ok := form.Value["cid["+index+"]"]
			if !ok || len(cid) == 0 {
				log.Warnf("Missing or empty cid for index %s", index)
				continue
			}

			cidOriginalEncrypted, ok := form.Value["cidOriginalEncrypted["+index+"]"]
			if !ok || len(cidOriginalEncrypted) == 0 {
				log.Warnf("Missing or empty cidOriginalEncrypted for index %s", index)
				continue
			}

			webkitRelativePath, ok := form.Value["webkitRelativePath["+index+"]"]
			if !ok || len(webkitRelativePath) == 0 {
				log.Warnf("Missing or empty webkitRelativePath for index %s", index)
				continue
			}
			/*
				_, params, err := mime.ParseMediaType(encryptedFile.Header.Get("Content-Disposition"))
				if err != nil {
					log.Errorf("parse media type: %s", err)
					AbortInternalServerError(ctx)
					return
				}
			*/
			mime := encryptedFile.Header.Get("Content-Type")

			// create corresponding folders to locate this file at proper path
			file_path := webkitRelativePath[0]
			actual_root, err := GetAndProcessFileRoot(file_path, r, authPayload.UserID, entity.Encrypted)
			if err != nil {
				log.Errorf("get and process file root: %s", err)
				AbortInternalServerError(ctx)
				return
			}
			log.Infof("actual_root: %s", actual_root)
			//log.Infof("Length of CID: %d", len(cid[0]))

			// create file
			f := entity.File{
				Name:                 encryptedFile.Filename,
				Root:                 actual_root,
				CID:                  cid[0],
				CIDOriginalEncrypted: &cidOriginalEncrypted[0],
				Mime:                 mime,
				Size:                 encryptedFile.Size,
				EncryptionStatus:               entity.Encrypted,
			}

			if err := f.Create(); err != nil {
				log.Errorf("create encrypted file: %s", err)
				AbortInternalServerError(ctx)
				return
			}

			// create file_user relation
			f_u := entity.FileUser{
				FileID:     f.ID,
				UserID:     authPayload.UserID,
				Permission: entity.OwnerPermission,
			}
			if err := f_u.Create(); err != nil {
				log.Errorf("create file_user relation: %s", err)
				AbortInternalServerError(ctx)
				return
			}

		ctx.JSON(http.StatusOK, fmt.Sprintf("%d files and %d encrypted files uploaded!", len(files), len(encryptedFiles)))
			keyPath := authPayload.UserUID + "/" + f.UID
			// upload file
			if err := UploadFileToS3(encryptedFile, keyPath); err != nil {
				log.Errorf("uploading file to s3: %s", err)
				AbortInternalServerError(ctx)
				return
			}

			// add user storage quantity
			user_detail := query.FindUserDetailByUserID(authPayload.UserID)

			if err := user_detail.Update("storage_used", user_detail.StorageUsed+uint(encryptedFile.Size)); err != nil {
				log.Errorf("adding storage_used: %s", err)
				AbortInternalServerError(ctx)
				return
			}

		}

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
func GetAndProcessFileRoot(file_path, root string, user_id uint, encryption_status entity.EncryptionStatus) (string, error) {
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
			Title: sub_title,
			Root:  root,
			EncryptionStatus: encryption_status,
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

	return GetAndProcessFileRoot(sub_file_path, f.UID, user_id, encryption_status)
}