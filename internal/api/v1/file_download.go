package v1

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
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
		authPayload := ctx.MustGet(constant.APIKeyHeaderKey).(*token.Payload)

		uid := ctx.Param("uid")

		//make sure the file exist
		f, err := query.FindFileByUID(uid)
		if err != nil {
			api.AbortEntityNotFound(ctx)
			return
		}

		keyPath := f.CID + strconv.FormatUint(uint64(authPayload.UserID), 10)
		out, error := api.DownloadFileFromS3(keyPath)
		//if error contains "NoSuchKey" then set keyPath without the userUID
		if error != nil {
			if strings.Contains(error.Error(), "NoSuchKey") {
				keyPath = f.CID
				out, error = api.DownloadFileFromS3(keyPath)
				fmt.Printf("download file: %s", error)
				if error != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{
						"message": error.Error(),
					})
					return
				}

			} else {
				fmt.Printf("download file: %s", error)
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": error.Error(),
				})
				return
			}

		}
		// Set the correct content type and file name
		ctx.Header("Content-Type", *out.ContentType)
		fmt.Printf("Content-Type: %s", *out.ContentType)
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; cid=%s", f.CID))

		// Copy the file data to the response
		_, error = io.Copy(ctx.Writer, out.Body)
		if error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": error.Error(),
			})
		}

	})
}
