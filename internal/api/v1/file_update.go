package v1

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

func FileUpdate(router *gin.RouterGroup) {
	router.PUT("/files/:uid", func(ctx *gin.Context) {

		//start the transactions to make changes
		tx := db.Db().Begin()
		//get the user data from de apikey
		authPayload := ctx.MustGet(constant.APIKeyHeaderKey).(*token.Payload)
		//params
		uid := ctx.Param("uid")
		var fileResponses []form.FileResponse

		// Multipart form
		formMultipart, err := ctx.MultipartForm()
		if err != nil {
			fmt.Printf("multipart form: %s", err)
			api.AbortBadRequest(ctx)
			return
		}

		//make sure the request body is ok
		root := formMultipart.Value["root"]
		files := formMultipart.File["file"]
		if len(files) < 1 {
			api.AbortBadRequest(ctx)
			return
		}
		file := files[0]
		// set the root
		var r string
		if len(root) > 0 {
			r = root[0]
		} else {
			r = "/"
		}

		//make sure the file exist
		f, err := query.FindFileByUID(uid)
		if err != nil {
			api.AbortEntityNotFound(ctx)
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
		actual_root, _, err := api.GetAndProcessFileRoot(file_path, r, authPayload.UserID, entity.Public)
		if err != nil {
			fmt.Printf("get and process file root: %s", err)
			api.AbortInternalServerError(ctx)
			return
		}

		f.Name = file.Filename
		f.Root = actual_root
		f.Mime = mime
		f.Size = file.Size
		f.UpdatedAt = time.Now()

		f.Save()

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

		tx.Commit()
		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
			"files":  fileResponses,
		})
	})
}
