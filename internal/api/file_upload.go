package api

import (
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/s3"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CheckFiles checks if files key (cid) exist in s3 "pool"
//
// POST /api/file/pool/check
// - JSON customFileMeta (array) that is passed from frontend
func CheckFilesExistInPool(router *gin.RouterGroup) {
	router.POST("/pool/check", func(ctx *gin.Context) {
		tx := db.Db().Begin()
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		var customFileMetas []form.CustomFileMeta
		if err := ctx.ShouldBindJSON(&customFileMetas); err != nil {
			log.Errorf("should bind json: %s", err)
			AbortBadRequest(ctx)
			return
		}

		// Check if files exist in s3
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

		// Create a map to track CIDs that have been processed
		cidProcessed := make(map[string]bool)

		//iterate over customFileMetas and check if the cid exists in s3
		var filesFoundResponses []form.FileResponse

		root := customFileMetas[0].Root

		var r string
		if len(root) > 0 {
			r = root
		} else {
			r = "/"
		}

		var firstRootUID string
		if len(customFileMetas) > 0 {
			for _, customFileMeta := range customFileMetas {
				log.Printf("Current file CID: %v", customFileMeta.CID)
				// Skip if CID has already been processed

				_, err := s3.HeadObject(s3Config, config.Env().WasabiBucket, customFileMeta.CID)
				if cidProcessed[customFileMeta.CID] {
					log.Infof("cidProcessed: %v", customFileMeta.CID)
					createFiles(&firstRootUID, customFileMeta, r, authPayload, tx, ctx, &filesFoundResponses)
					cidProcessed[customFileMeta.CID] = true
				} else if err != nil {
					//this means that the object doesn't exist at S3, so we can return CID to frontend for later upload of binary and metadata
					log.Info("headobject for following cid is nil:")
					log.Info(customFileMeta.CID)
					cidProcessed[customFileMeta.CID] = true
				} else {
					log.Infof("none of the above: %v", customFileMeta.CID)
					createFiles(&firstRootUID, customFileMeta, r, authPayload, tx, ctx, &filesFoundResponses)
					cidProcessed[customFileMeta.CID] = true
				}

			}
		} else {
			log.Print("customFileMetas is empty")
		}
		log.Printf("filesFoundResponses: %v", filesFoundResponses)

		tx.Commit()
		ctx.JSON(http.StatusOK, gin.H{
			"status":       "success",
			"filesFound":   filesFoundResponses,
			"firstRootUID": firstRootUID,
		})
	})
}

func createFiles(firstRootUID *string, customFileMeta form.CustomFileMeta, r string, authPayload *token.Payload, tx *gorm.DB, ctx *gin.Context, filesFoundResponses *[]form.FileResponse) {
	//this means that the object exists at S3, so we can create a file entry on database for the file

	//cid of file
	mime := customFileMeta.MimeType

	// create corresponding folders to locate this file at proper path
	file_path := customFileMeta.Path
	var f entity.File

	encryptionStatus := entity.Encrypted
	if customFileMeta.EncryptionStatus == entity.Public {
		encryptionStatus = entity.Public
	}

	actual_root, firstCreatedRoot, err := GetAndProcessFileRoot(file_path, r, authPayload.UserID, encryptionStatus)
	if err != nil {
		log.Errorf("get and process file root: %s", err)
		AbortInternalServerError(ctx)
		return
	}

	if *firstRootUID == "" {
		*firstRootUID = firstCreatedRoot
	}

	var isInPool bool = true

	// create file
	f = entity.File{
		Name:                 customFileMeta.Name,
		Root:                 actual_root,
		CID:                  customFileMeta.CID,
		CIDOriginalEncrypted: &customFileMeta.CIDOriginalEncrypted,
		Mime:                 mime,
		Size:                 customFileMeta.Size,
		EncryptionStatus:     encryptionStatus,
		CreatedAt:            time.Now(),
		IsInPool:             &isInPool,
		UpdatedAt:            time.Now(),
	}

	if err := f.TxCreate(tx); err != nil {
		log.Errorf("create file: %s", err)
		tx.Rollback()
		AbortInternalServerError(ctx)
	}

	fResponse := form.FileResponse{
		ID:                   f.ID,
		Name:                 f.Name,
		UID:                  f.UID,
		Root:                 f.Root,
		CID:                  f.CID,
		Mime:                 f.Mime,
		CIDOriginalEncrypted: f.CIDOriginalEncrypted,
		Size:                 f.Size,
		IsInPool:             &isInPool,
		EnryptionStatus:      f.EncryptionStatus,
		CreatedAt:            f.CreatedAt.String(),
		UpdatedAt:            f.UpdatedAt.String(),
	}

	*filesFoundResponses = append(*filesFoundResponses, fResponse)

	// create file_user relation with the shared permision
	f_u := entity.FileUser{
		FileID:     f.ID,
		UserID:     authPayload.UserID,
		Permission: entity.SharedPermission,
	}
	if err := f_u.TxCreate(tx); err != nil {
		log.Errorf("create file_user relation: %s", err)
		tx.Rollback()
		AbortInternalServerError(ctx)
		return
	}

}

// UploadFiles upload files to wasabi using s3
//
// POST /api/file/upload
// Form: MultipartForm
// - files
// - root
func PutUploadFiles(router *gin.RouterGroup) {

	router.POST("/upload", func(ctx *gin.Context) {
		tx := db.Db().Begin()
		authPayload := ctx.MustGet(constant.AuthorizationPayloadKey).(*token.Payload)
		var fileResponses []form.FileResponse

		// Multipart form
		formMultipart, err := ctx.MultipartForm()

		if err != nil {
			log.Errorf("multipart form: %s", err)
			AbortBadRequest(ctx)
			return
		}

		root := formMultipart.Value["root"]
		files := formMultipart.File["files"]

		var r string
		if len(root) > 0 {
			r = root[0]
		} else {
			r = "/"
		}

		encryptedFiles := formMultipart.File["encryptedFiles"]
		var totalSize uint
		var firstRootUID string
		// Handle regular files
		for index, file := range files {

			index := fmt.Sprintf("%d", index)
			totalSize += uint(file.Size)

			//cid of file
			cid, ok := formMultipart.Value["cid["+index+"]"]
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
			actual_root, firstCreatedRoot, err := GetAndProcessFileRoot(file_path, r, authPayload.UserID, entity.Public)

			if err != nil {
				log.Errorf("get and process file root: %s", err)
				AbortInternalServerError(ctx)
				return
			}

			if firstRootUID == "" {
				firstRootUID = firstCreatedRoot
			}

			// create file
			f := entity.File{
				Name:             file.Filename,
				Root:             actual_root,
				CID:              cid[0],
				Mime:             mime,
				Size:             file.Size,
				EncryptionStatus: entity.Public,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := f.TxCreate(tx); err != nil {
				log.Errorf("create file: %s", err)
				tx.Rollback()
				AbortInternalServerError(ctx)
			}

			fResponse := form.FileResponse{
				ID:              f.ID,
				Name:            f.Name,
				UID:             f.UID,
				Root:            f.Root,
				CID:             f.CID,
				Mime:            f.Mime,
				Size:            f.Size,
				EnryptionStatus: f.EncryptionStatus,
				CreatedAt:       f.CreatedAt.String(),
				UpdatedAt:       f.UpdatedAt.String(),
			}

			fileResponses = append(fileResponses, fResponse)

			// create file_user relation
			f_u := entity.FileUser{
				FileID:     f.ID,
				UserID:     authPayload.UserID,
				Permission: entity.OwnerPermission,
			}
			if err := f_u.TxCreate(tx); err != nil {
				log.Errorf("create file_user relation: %s", err)
				tx.Rollback()
				AbortInternalServerError(ctx)
				return
			}

			keyPath := f.CID

			// upload file

			go func(file *multipart.FileHeader, keyPath string) {
				if err := UploadFileToS3(file, keyPath); err != nil {
					log.Errorf("uploading file to s3: %s", err)
					tx.Rollback()
					AbortInternalServerError(ctx)
					return
				}
			}(file, keyPath)

		}

		for key, encryptedFile := range encryptedFiles {
			// Ensure the key exists and has values

			index := fmt.Sprintf("%d", key)

			totalSize += uint(encryptedFile.Size)

			//cid of encrypted buffer
			cid, ok := formMultipart.Value["cid["+index+"]"]
			if !ok || len(cid) == 0 {
				log.Warnf("Missing or empty cid for index %s", index)
				continue
			}

			cidOriginalEncrypted, ok := formMultipart.Value["cidOriginalEncrypted["+index+"]"]
			if !ok || len(cidOriginalEncrypted) == 0 {
				log.Warnf("Missing or empty cidOriginalEncrypted for index %s", index)
				continue
			}

			webkitRelativePath, ok := formMultipart.Value["webkitRelativePath["+index+"]"]
			if !ok || len(webkitRelativePath) == 0 {
				log.Warnf("Missing or empty webkitRelativePath for index %s", index)
				continue
			}

			mime := encryptedFile.Header.Get("Content-Type")

			// create corresponding folders to locate this file at proper path
			file_path := webkitRelativePath[0]
			actual_root, firstCreatedRoot, err := GetAndProcessFileRoot(file_path, r, authPayload.UserID, entity.Encrypted)
			if err != nil {
				log.Errorf("get and process file root: %s", err)
				AbortInternalServerError(ctx)
				return
			}
			if firstRootUID == "" {
				firstRootUID = firstCreatedRoot
			}

			// create file
			f := entity.File{
				Name:                 encryptedFile.Filename,
				Root:                 actual_root,
				CID:                  cid[0],
				CIDOriginalEncrypted: &cidOriginalEncrypted[0],
				Mime:                 mime,
				Size:                 encryptedFile.Size,
				EncryptionStatus:     entity.Encrypted,
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			}

			if err := f.TxCreate(tx); err != nil {
				log.Errorf("create encrypted file: %s", err)
				tx.Rollback()
				AbortInternalServerError(ctx)
				return
			}

			fResponse := form.FileResponse{
				Name:                 f.Name,
				UID:                  f.UID,
				Root:                 f.Root,
				CID:                  f.CID,
				CIDOriginalEncrypted: f.CIDOriginalEncrypted,
				ID:                   f.ID,
				Mime:                 f.Mime,
				Size:                 f.Size,
				EnryptionStatus:      f.EncryptionStatus,
				CreatedAt:            f.CreatedAt.String(),
				UpdatedAt:            f.UpdatedAt.String(),
			}

			fileResponses = append(fileResponses, fResponse)

			// create file_user relation
			f_u := entity.FileUser{
				FileID:     f.ID,
				UserID:     authPayload.UserID,
				Permission: entity.OwnerPermission,
			}
			if err := f_u.TxCreate(tx); err != nil {
				log.Errorf("create file_user relation: %s", err)
				tx.Rollback()
				AbortInternalServerError(ctx)
				return
			}

			keyPath := f.CID
			// upload file
			go func(file *multipart.FileHeader, keyPath string) {
				if err := UploadFileToS3(file, keyPath); err != nil {
					log.Errorf("uploading file to s3: %s", err)
					tx.Rollback()
					AbortInternalServerError(ctx)
					return
				}
			}(encryptedFile, keyPath)
		}

		// add user storage quantity
		user_detail := query.FindUserDetailByUserID(authPayload.UserID)

		if err := user_detail.TxUpdate(tx, "storage_used", user_detail.StorageUsed+totalSize); err != nil {
			log.Errorf("adding storage_used: %s", err)
			AbortInternalServerError(ctx)
			return
		}
		tx.Commit()
		ctx.JSON(http.StatusOK, gin.H{
			"status":       "success",
			"files":        fileResponses,
			"firstRootUID": firstRootUID,
		})

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

func GetAndProcessFileRoot(file_path, root string, user_id uint, encryption_status entity.EncryptionStatus) (currentRoot, firstRoot string, err error) {
	res := strings.Split(file_path, "/")
	if len(res) == 1 {
		return root, "", nil
	}

	sub_file_path := strings.Join(res[1:], "/")
	sub_title := res[0]

	f := query.FindFolderByTitleAndRoot(sub_title, root)

	log.Infof("folder find by title and root: %v", f)
	if f == nil {
		f = &entity.Folder{
			Title:            sub_title,
			Root:             root,
			EncryptionStatus: encryption_status,
		}

		if err := f.Create(); err != nil {
			return "", "", errors.New("can't create folder")
		}
		// create folder_user relation
		f_u := &entity.FolderUser{
			FolderID:   f.ID,
			UserID:     user_id,
			Permission: entity.OwnerPermission,
		}

		if err := f_u.Create(); err != nil {
			return "", "", errors.New("can't create folder_user relation")
		}
	}

	// Recursive call
	newRoot, firstCreatedRoot, err := GetAndProcessFileRoot(sub_file_path, f.UID, user_id, encryption_status)
	if err != nil {
		return "", "", err
	}

	// If this is the first folder created, keep its UID to return all the way up the recursive calls.
	if firstCreatedRoot == "" {
		firstCreatedRoot = f.UID
	}

	return newRoot, firstCreatedRoot, nil
}
