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
	// CountDailyStorage    int64 `json:"CountDailyStorage"`
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
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		totalusers, err := query.CountUsers()
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		upfile, err := query.CountFiles()
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		msize, err := query.CountMediumSizeFiles()
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		encryptedfiles, err := query.CountEncryptedFiles()
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		publicfiles, err := query.CountPublicFiles()
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		// temptime := time.Now().Format("2006-01-02 15:04:05")

		//To Do: implment .map or .foreach for recursive query
		// daylystorage, err := query.CountDailyStorage(temptime)
		// if err != nil {
		// 	AbortEntityNotFound(c)
		// 	return
		// }

		stats := Statistics{
			TotalUsedStorage:     totalusedstorage,
			UploadedFile:         upfile,
			TotalUsers:           totalusers,
			CountMediumSizeFiles: msize,
			EncryptedFiles:       encryptedfiles,
			PublicFiles:          publicfiles,
			// CountDailyStorage:    daylystorage,
		}

		c.JSON(http.StatusOK, stats)
	})
}
