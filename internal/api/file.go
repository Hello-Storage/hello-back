package api

import (
	"net/http"
	"time"

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
	PublicFolders        int64 `json:"PublicFolders"`
	CountTxtFiles        int64 `json:"CountTxtFiles"`
	CountPngFiles        int64 `json:"CountPngFiles"`
	CountJpgFiles        int64 `json:"CountJpgFiles"`
	CountPdfFiles        int64 `json:"CountPdfFiles"`
	// CountDailyStorage    int64 `json:"CountDailyStorage"`
}

type UserStatistics struct {
	CountTotalUsedStorageUser    int64 `json:"CountTotalUsedStorageUser"`
	CountTotalEncryptedFilesUser int64 `json:"CountTotalEncryptedFilesUser"`
	CountTotalPublicFilesUser    int64 `json:"CountTotalPublicFilesUser"`
	CountTotalFilesUser          int64 `json:"CountTotalFilesUser"`
	CountTotalPublicFoldersUser  int64 `json:"CountTotalPublicFoldersUser"`
}

type UserDailyStatistics struct {
	CountDailyStorageUser1 int64 `json:"CountDailyStorageUser1"`
	CountDailyStorageUser2 int64 `json:"CountDailyStorageUser2"`
	CountDailyStorageUser3 int64 `json:"CountDailyStorageUser3"`
	CountDailyStorageUser4 int64 `json:"CountDailyStorageUser4"`
	CountDailyStorageUser5 int64 `json:"CountDailyStorageUser5"`
	CountDailyStorageUser6 int64 `json:"CountDailyStorageUser6"`
	CountDailyStorageUser7 int64 `json:"CountDailyStorageUser7"`
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
			log.Errorf("cannot get total used storage: %s", err)
			AbortEntityNotFound(c)
			return
		}

		totalusers, err := query.CountUsers()
		if err != nil {
			log.Errorf("cannot get total users: %s", err)
			AbortEntityNotFound(c)
			return
		}

		upfile, err := query.CountFiles()
		if err != nil {
			log.Errorf("cannot get total uploaded files: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//medium size files
		msize := totalusedstorage / upfile

		encryptedfiles, err := query.CountEncryptedFiles()
		if err != nil {
			log.Errorf("cannot get total encrypted files: %s", err)
			AbortEntityNotFound(c)
			return
		}

		publicfiles, err := query.CountPublicFiles()
		if err != nil {
			log.Errorf("cannot get total public files: %s", err)
			AbortEntityNotFound(c)
			return
		}

		publicfolders, err := query.CountPublicFolders()
		if err != nil {
			log.Errorf("cannot get total public folders: %s", err)
			AbortEntityNotFound(c)
			return
		}

		counttxtfiles, err := query.CountTxtFiles()
		if err != nil {
			log.Errorf("cannot get total txt files: %s", err)
			AbortEntityNotFound(c)
			return
		}

		countpngfiles, err := query.CountPngFiles()
		if err != nil {
			log.Errorf("cannot get total png fileas: %s", err)
			AbortEntityNotFound(c)
			return
		}

		countjpgfiles, err := query.CountJpgFiles()
		if err != nil {
			log.Errorf("cannot get total jpg files: %s", err)
			AbortEntityNotFound(c)
			return
		}

		countpdffiles, err := query.CountPdfFiles()
		if err != nil {
			log.Errorf("cannot get total pdf files: %s", err)
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
			PublicFolders:        publicfolders,
			CountTxtFiles:        counttxtfiles,
			CountPngFiles:        countpngfiles,
			CountJpgFiles:        countjpgfiles,
			CountPdfFiles:        countpdffiles,

			// CountDailyStorage:    daylystorage,
		}

		c.JSON(http.StatusOK, stats)
	})
	router.GET("/statistics/:uid", func(c *gin.Context) {

		uid := c.Param("uid")
		counttotalusedstorageuser, err := query.CountTotalUsedStorageUser(uid)
		if err != nil {
			log.Errorf("cannot get total used storage for user %s: %s", uid, err)
			AbortEntityNotFound(c)
			return
		}

		counttotalencryptedfilesuser, err := query.CountTotalEncryptedFilesUser(uid)
		if err != nil {
			log.Errorf("cannot get total encrypted files for user %s: %s", uid, err)
			AbortEntityNotFound(c)
			return
		}

		counttotalpublicfilesuser, err := query.CountTotalPublicFilesUser(uid)
		if err != nil {
			log.Errorf("cannot get total public files for user %s: %s", uid, err)
			AbortEntityNotFound(c)
			return
		}

		counttotalfilesuser, err := query.CountTotalFilesUser(uid)
		if err != nil {
			log.Errorf("cannot get total files for user %s: %s", uid, err)
			AbortEntityNotFound(c)
			return
		}

		counttotalpublicfoldersuser, err := query.CountTotalPublicFoldersUser(uid)
		if err != nil {
			log.Errorf("cannot get total public folders for user %s: %s", uid, err)
			AbortEntityNotFound(c)
			return
		}

		stats := UserStatistics{
			CountTotalUsedStorageUser:    counttotalusedstorageuser,
			CountTotalEncryptedFilesUser: counttotalencryptedfilesuser,
			CountTotalPublicFilesUser:    counttotalpublicfilesuser,
			CountTotalFilesUser:          counttotalfilesuser,
			CountTotalPublicFoldersUser:  counttotalpublicfoldersuser,
		}
		c.JSON(http.StatusOK, stats)
	})

	router.GET("/statistics/:uid/day", func(c *gin.Context) {
		uid := c.Param("uid")

		var countdailystorageuser [7]int64

		for i := 0; i < len(countdailystorageuser); i++ {
			err := error(nil)
			// Get today date in this format (year-month-day hour:minute:second)
			temptime1 := time.Now().AddDate(0, 0, -i).Format("2006-01-02 15:04:05")
			// sum 1 days to today date
			temptime2 := time.Now().AddDate(0, 0, -i+1).Format("2006-01-02 15:04:05")
			// set hours, seconds, minutes to 00:00:00
			temptime1 = temptime1[0:11] + "00:00:00"
			temptime2 = temptime2[0:11] + "00:00:00"

			countdailystorageuser[i], err = query.CountDailyStorageUser(temptime1, temptime2, uid)

			if err != nil {
				log.Errorf("cannot get total used storage for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			log.Info(countdailystorageuser[i])
		}

		// stats := UserDailyStatistics{
		// 	CountDailyStorageUser1: countdailystorageuser1,
		// 	CountDailyStorageUser2: countdailystorageuser2,
		// 	CountDailyStorageUser3: countdailystorageuser3,
		// 	CountDailyStorageUser4: countdailystorageuser4,
		// 	CountDailyStorageUser5: countdailystorageuser5,
		// 	CountDailyStorageUser6: countdailystorageuser6,
		// 	CountDailyStorageUser7: countdailystorageuser7,
		// }
		// c.JSON(http.StatusOK, stats)
		c.JSON(http.StatusOK, countdailystorageuser)

	})
}
