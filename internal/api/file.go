package api

import (
	"net/http"

	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/gin-gonic/gin"
)

type Statistics struct {
	TotalUsedStorage     int64 `json:"TotalUsedStorage"`
	UploadedFile         int64 `json:"UploadedFile"`
	TotalUsers           int64 `json:"TotalUsers"`
	CountMediumSizeFiles int64 `json:"CountMediumSizeFiles"`
	EncryptedFiles       int64 `json:"EncryptedFiles"`
	PublicFiles          int64 `json:"PublicFiles"`
}

// GetFile returns file details as JSON.
//
// GET /api/file/info/:uid
// Params:
// - uid
func GetFile(router *gin.RouterGroup) {

	router.GET("/info/:uid", func(c *gin.Context) {
		// To Do check access grant
		uid := c.Param("uid")

		p, err := query.FindFileByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}

func GetStatistics(router *gin.RouterGroup) {
	router.GET("/statistics", func(c *gin.Context) {
		totalusedstorage, err := query.CountTotalUsedStorage()
		totalusers, err := query.CountUsers()
		upfile, err := query.CountFiles()
		msize, err := query.CountMediumSizeFiles()
		encryptedfiles, err := query.CountEncryptedFiles()
		publicfiles, err := query.CountPublicFiles()

		stats := Statistics{
			TotalUsedStorage:     totalusedstorage,
			UploadedFile:         upfile,
			TotalUsers:           totalusers,
			CountMediumSizeFiles: msize,
			EncryptedFiles:       encryptedfiles,
			PublicFiles:          publicfiles,
		}

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, stats)
	})
}
