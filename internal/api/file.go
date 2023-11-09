package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
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
	CountDailyStorageUser        [12]int64 `json:"CountDailyStorageUser"`
	CountDailyFilesUser          [12]int64 `json:"CountDailyFilesUser"`
	CountDailyPublicFilesUser    [12]int64 `json:"CountDailyPublicFilesUser"`
	CountDailyEncryptedFilesUser [12]int64 `json:"CountDailyEncryptedFilesUser"`
}

type UserWeeklyStatistics struct {
	CountDailyStorageUser        [7]int64 `json:"CountDailyStorageUser"`
	CountDailyFilesUser          [7]int64 `json:"CountDailyFilesUser"`
	CountDailyPublicFilesUser    [7]int64 `json:"CountDailyPublicFilesUser"`
	CountDailyEncryptedFilesUser [7]int64 `json:"CountDailyEncryptedFilesUser"`
}

type UserMonthlyStatistics struct {
	CountDailyStorageUser        [30]int64 `json:"CountDailyStorageUser"`
	CountDailyFilesUser          [30]int64 `json:"CountDailyFilesUser"`
	CountDailyPublicFilesUser    [30]int64 `json:"CountDailyPublicFilesUser"`
	CountDailyEncryptedFilesUser [30]int64 `json:"CountDailyEncryptedFilesUser"`
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

// GetShareState returns file share state based on file id.
//
// GET /share/state
// Params:
// - file_uid
func GetShareState(router *gin.RouterGroup) {
	router.GET("/share/state", func(c *gin.Context) {
		//get file id from params
		file_uid := c.Query("file_uid")
		fileMutex := sync.Mutex{}

		fileMutex.Lock()
		defer fileMutex.Unlock()

		//check if file exists
		f, err := query.FindFileByUID(file_uid)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//get share state, if doesn't exist, create it
		share_state, err := query.FindShareStateByFileUID(file_uid)
		if err != nil {
			log.Errorf("Error finding share state: %s", err)
			share_state, err = query.CreateShareState(f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		c.JSON(http.StatusOK, share_state)
	})
}

// PublishFile publishes a file.
//
// POST /share/publish
// Params:
// - selectedShareFile: form.CustomFileMeta
func PublishFile(router *gin.RouterGroup) {
	router.POST("/share/publish", func(c *gin.Context) {
		//get file id from params
		var selectedShareFile form.CustomFileMeta
		err := c.BindJSON(&selectedShareFile)
		if err != nil {
			log.Errorf("cannot bind json: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//check if file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//get share state, if doesn't exist, create it
		share_state, err := query.FindShareStateByFileUID(selectedShareFile.UID)
		if err != nil {
			share_state, err = query.CreateShareState(f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		//update share state

		public_file, err := query.PublishFile(share_state, selectedShareFile)
		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		share_state.PublicFile = *public_file

		share_state.UpdatedAt = time.Now()

		err = share_state.Save()

		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, share_state)
	})
}

// UnpublishFile unpublishes a file.
//
// POST /share/unpublish
// Params:
// - selectedShareFile: form.CustomFileMeta
func UnpublishFile(router *gin.RouterGroup) {
	router.POST("/share/unpublish", func(c *gin.Context) {
		//get file id from params
		var selectedShareFile form.CustomFileMeta
		err := c.BindJSON(&selectedShareFile)
		if err != nil {
			log.Errorf("cannot bind json: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//check if file exists
		f, err := query.FindFileByUID(selectedShareFile.UID)
		if err != nil {
			log.Errorf("cannot get file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		//get share state, if doesn't exist, create it
		share_state, err := query.FindShareStateByFileUID(selectedShareFile.UID)
		if err != nil {
			share_state, err = query.CreateShareState(f)
			if err != nil {
				log.Errorf("cannot create share state: %s", err)
				AbortEntityNotFound(c)
				return
			}
		}

		//update share state
		//delete public file from db
		//update share state
		if share_state.PublicFile.ID != 0 { // Check if the PublicFile has a valid ID
			err = share_state.PublicFile.Delete()
			if err != nil {
				log.Errorf("cannot delete public file: %s", err)
				AbortEntityNotFound(c)
				return
			}
		} else {
			log.Info("No existing public file to delete.")
		}

		share_state.PublicFile = entity.PublicFile{}

		share_state.UpdatedAt = time.Now()

		err = share_state.Save()

		if err != nil {
			log.Errorf("cannot update share state: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, share_state)
	})
}

// GetPublishedFile publishes a file.
//
// GET /share/published/:hash
// Params:
// - hash: form.PublicFile.ShareHash
func GetPublishedFile(router *gin.RouterGroup) {
	router.GET("/share/published/:hash", func(c *gin.Context) {
		log.Print("GetPublishedFile")
		//get file id from params
		hash := c.Param("hash")

		log.Print(hash)
		//get public file
		public_file, err := query.FindPublicFileByHash(hash)
		if err != nil {
			log.Errorf("cannot get public file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, public_file)
	})
}

// GetPublishedFileName returns a published file name.
//
// GET /share/published/name/:hash
// Params:
// - hash: form.PublicFile.ShareHash
func GetPublishedFileName(router *gin.RouterGroup) {
	router.GET("/share/published/name/:hash", func(c *gin.Context) {
		//get file id from params
		hash := c.Param("hash")

		//get public file
		public_file, err := query.FindPublicFileByHash(hash)
		if err != nil {
			log.Errorf("cannot get public file: %s", err)
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, public_file.Name)
	})
}

func GetStatistics(router *gin.RouterGroup) {
	/*
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

	*/
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

		var countdailystorageuser [12]int64
		var countdailyfileuser [12]int64
		var countdailypublicfilesuser [12]int64
		var countdailyencryptedfilesuser [12]int64

		for i := 0; i < len(countdailystorageuser); i++ {
			err := error(nil)
			temptime := time.Now().Add(-time.Duration(i*2) * time.Hour).Format("2006-01-02 15:04:05")

			countdailystorageuser[i], err = query.CountStorageUsedH(temptime, uid)

			if err != nil {
				log.Errorf("cannot get total used storage for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailyfileuser[i], err = query.CountFilesUsedH(temptime, uid)

			if err != nil {
				log.Errorf("cannot get total used files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailypublicfilesuser[i], err = query.CountFilesUsedByStatusH(temptime, uid, "public")

			if err != nil {
				log.Errorf("cannot get total used public files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailyencryptedfilesuser[i], err = query.CountFilesUsedByStatusH(temptime, uid, "encrypted")

			if err != nil {
				log.Errorf("cannot get total used encrypted files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			log.Info(countdailystorageuser[i])
		}

		stats := UserDailyStatistics{
			CountDailyStorageUser:        countdailystorageuser,
			CountDailyFilesUser:          countdailyfileuser,
			CountDailyPublicFilesUser:    countdailypublicfilesuser,
			CountDailyEncryptedFilesUser: countdailyencryptedfilesuser,
		}

		c.JSON(http.StatusOK, stats)
	})

	router.GET("/statistics/:uid/week", func(c *gin.Context) {
		uid := c.Param("uid")

		var countdailystorageuser [7]int64
		var countdailyfileuser [7]int64
		var countdailypublicfilesuser [7]int64
		var countdailyencryptedfilesuser [7]int64

		for i := 0; i < len(countdailystorageuser); i++ {
			err := error(nil)
			temptime := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

			countdailystorageuser[i], err = query.CountStorageUsed(temptime, uid)

			if err != nil {
				log.Errorf("cannot get total used storage for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailyfileuser[i], err = query.CountFilesUsed(temptime, uid)

			if err != nil {
				log.Errorf("cannot get total used files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailypublicfilesuser[i], err = query.CountFilesUsedByStatus(temptime, uid, "public")

			if err != nil {
				log.Errorf("cannot get total used public files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailyencryptedfilesuser[i], err = query.CountFilesUsedByStatus(temptime, uid, "encrypted")

			if err != nil {
				log.Errorf("cannot get total used encrypted files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			log.Info(countdailystorageuser[i])
		}

		stats := UserWeeklyStatistics{
			CountDailyStorageUser:        countdailystorageuser,
			CountDailyFilesUser:          countdailyfileuser,
			CountDailyPublicFilesUser:    countdailypublicfilesuser,
			CountDailyEncryptedFilesUser: countdailyencryptedfilesuser,
		}

		c.JSON(http.StatusOK, stats)
	})

	router.GET("/statistics/:uid/month", func(c *gin.Context) {
		uid := c.Param("uid")

		var countdailystorageuser [30]int64
		var countdailyfileuser [30]int64
		var countdailypublicfilesuser [30]int64
		var countdailyencryptedfilesuser [30]int64

		for i := 0; i < len(countdailystorageuser); i++ {
			err := error(nil)
			temptime := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

			countdailystorageuser[i], err = query.CountStorageUsed(temptime, uid)

			if err != nil {
				log.Errorf("cannot get total used storage for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailyfileuser[i], err = query.CountFilesUsed(temptime, uid)

			if err != nil {
				log.Errorf("cannot get total used files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailypublicfilesuser[i], err = query.CountFilesUsedByStatus(temptime, uid, "public")

			if err != nil {
				log.Errorf("cannot get total used public files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			countdailyencryptedfilesuser[i], err = query.CountFilesUsedByStatus(temptime, uid, "encrypted")

			if err != nil {
				log.Errorf("cannot get total used encrypted files for user %s: %s", uid, err)
				AbortEntityNotFound(c)
				return
			}
			log.Info(countdailystorageuser[i])
		}

		stats := UserMonthlyStatistics{
			CountDailyStorageUser:        countdailystorageuser,
			CountDailyFilesUser:          countdailyfileuser,
			CountDailyPublicFilesUser:    countdailypublicfilesuser,
			CountDailyEncryptedFilesUser: countdailyencryptedfilesuser,
		}

		c.JSON(http.StatusOK, stats)
	})

}
