package v1

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

func GetFile(router *gin.RouterGroup) {

	router.GET("/file/:uid", func(c *gin.Context) {

		uid := c.Param("uid")

		p, err := query.FindFileByUID(uid)

		if err != nil {
			api.AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})

	router.GET("/files", func(c *gin.Context) {

		authPayload := c.MustGet(constant.APIKeyHeaderKey).(*token.Payload)

		files, err := query.GetApiFiles(authPayload.UserID)

		if err != nil {
			api.AbortEntityNotFound(c)
			return
		}
		var fileResponses []form.FileResponse

		for _, f := range files {

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

		}

		c.JSON(http.StatusOK, gin.H{
			"files": fileResponses,
		})
	})
}
