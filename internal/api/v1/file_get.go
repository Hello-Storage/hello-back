package v1

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/constant"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/Hello-Storage/hello-back/pkg/token"
	"github.com/gin-gonic/gin"
)

func GetFile(router *gin.RouterGroup) {

	router.GET("/files/:uid", func(c *gin.Context) {

		authPayload := c.MustGet(constant.APIKeyHeaderKey).(*token.Payload)

		//increment requests counter
		apiKey, err := query.FindApiKeyByUserID(authPayload.UserID)
		if err == nil && apiKey != nil {
			apiKey.IncrementKeyRequests()
		}

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

		//increment requests counter
		apiKey, err := query.FindApiKeyByUserID(authPayload.UserID)
		if err == nil && apiKey != nil {
			apiKey.IncrementKeyRequests()
		}

		pageNumber := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("pageSize", "10")

		pageNumberInt, err := strconv.Atoi(pageNumber)
		if err != nil || pageNumberInt < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil || pageSizeInt < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
			return
		}

		allFiles, err := query.GetApiFiles(authPayload.UserID)
		if err != nil {
			fmt.Print(err)
		}

		totalItems := len(allFiles)
		totalPages := int(math.Ceil(float64(totalItems) / float64(pageSizeInt)))

		startIndex := (pageNumberInt - 1) * pageSizeInt
		endIndex := pageNumberInt * pageSizeInt

		if startIndex >= totalItems {
			c.JSON(http.StatusOK, gin.H{"files": []form.FileResponse{}, "totalItems": totalItems, "totalPages": totalPages})
			return
		}

		if endIndex > totalItems {
			endIndex = totalItems
		}

		paginatedFiles := allFiles[startIndex:endIndex]

		var fileResponses []form.FileResponse

		for _, f := range paginatedFiles {
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
			"files":      fileResponses,
			"totalItems": totalItems,
			"totalPages": totalPages,
		})
	})
}

func Ping(router *gin.RouterGroup) {
	router.GET("/", func(c *gin.Context) {
		c.Header("X-Rate-Limit-Limit", "3")
		c.Header("X-Rate-Limit-Reset", "1")

		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the hello.app API!",
			"note":    "This route is for testing purposes and does not count towards the request limit. Be sure to enjoy the experience, 'hello' and goodbye.",
		})
	})
}
