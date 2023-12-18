package v1

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func FileCreate(router *gin.RouterGroup) {
	router.POST("/files", func(ctx *gin.Context) {
		tx := db.Db().Begin()
		authPayload := ctx.MustGet(constant.APIKeyHeaderKey).(*token.Payload)

		var fileResponses []form.FileResponse

		// Multipart form
		formMultipart, err := ctx.MultipartForm()

		if err != nil {
			fmt.Printf("multipart form: %s", err)
			api.AbortBadRequest(ctx)
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

		var totalSize uint
		var firstRootUID string
		// Handle regular files
		for _, file := range files {

			totalSize += uint(file.Size)

			//cid of file
			cid, err := GetCid(file)
			if err != nil {
				fmt.Printf("getting cid: %s", err)
				api.AbortInternalServerError(ctx)
				return
			}

			_, params, err := mime.ParseMediaType(file.Header.Get("Content-Disposition"))
			if err != nil {
				fmt.Printf("parse media type: %s", err)
				api.AbortInternalServerError(ctx)
				return
			}
			mime := file.Header.Get("Content-Type")

			// create corresponding folders to locate this file at proper path
			file_path := params["filename"]
			actual_root, firstCreatedRoot, err := api.GetAndProcessFileRoot(file_path, r, authPayload.UserID, entity.Public)

			if err != nil {
				fmt.Printf("get and process file root: %s", err)
				api.AbortInternalServerError(ctx)
				return
			}

			if firstRootUID == "" {
				firstRootUID = firstCreatedRoot
			}

			// create file
			f := entity.File{
				Name:             file.Filename,
				Root:             actual_root,
				CID:              cid,
				Mime:             mime,
				Size:             file.Size,
				EncryptionStatus: entity.Public,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			if err := f.TxCreate(tx); err != nil {
				fmt.Printf("create file: %s", err)
				tx.Rollback()
				api.AbortInternalServerError(ctx)
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
				fmt.Printf("create file_user relation: %s", err)
				tx.Rollback()
				api.AbortInternalServerError(ctx)
				return
			}

			// create api_key_file relation
			a_p_f := entity.ApiKeyFile{
				FileID: f.ID,
				UserID: authPayload.UserID,
			}

			if err := a_p_f.TxCreate(tx); err != nil {
				fmt.Printf("create file_user relation: %s", err)
				tx.Rollback()
				api.AbortInternalServerError(ctx)
				return
			}

			keyPath := f.CID + strconv.FormatUint(uint64(authPayload.UserID), 10)

			// upload file
			go func(file *multipart.FileHeader, keyPath string) {
				if err := api.UploadFileToS3(file, keyPath); err != nil {
					fmt.Printf("uploading file to s3: %s", err)
					tx.Rollback()
					api.AbortInternalServerError(ctx)
					return
				}
			}(file, keyPath)

		}

		// add user storage quantity
		user_detail := query.FindUserDetailByUserID(authPayload.UserID)

		if err := user_detail.TxUpdate(tx, "storage_used", user_detail.StorageUsed+totalSize); err != nil {
			fmt.Printf("adding storage_used: %s", err)
			api.AbortInternalServerError(ctx)
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

func GetCid(file *multipart.FileHeader) (string, error) {
	fileContent, err := file.Open()
	if err != nil {
		return "opening file:", err
	}
	defer fileContent.Close()

	buffer, err := io.ReadAll(fileContent)
	if err != nil {
		return "reading file:", err
	}

	hash, err := multihash.Sum(buffer, multihash.SHA2_256, -1)
	if err != nil {
		return "getting cid:", err
	}

	c := cid.NewCidV1(cid.Raw, hash)
	return c.String(), nil
}
