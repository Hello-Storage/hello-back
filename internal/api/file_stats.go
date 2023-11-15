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

type WeeklyStats struct {
	Week        string `json:"week"`
	UsedStorage int64  `json:"usedStorage"`
	Total       int64  `json:"total"`
	Public      int64  `json:"public"`
	Encrypted   int64  `json:"encrypted"`
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

func getStartAndEndFileDates() (time.Time, time.Time) {
	// You need to implement the logic for calculating the start and end dates for the weekly intervals
	// This can involve querying your database for the earliest and latest creation dates of files
	// For now, let's assume we have a function that gives us these dates
	startDate, endDate, err := query.GetStartAndEndFileDatesPublic()
	if err != nil {
		log.Errorf("cannot get start and end dates for files: %s", err)
		return time.Time{}, time.Time{}
	}

	return startDate, endDate
}

func GetWeeklyPublicStats(router *gin.RouterGroup) {
	router.GET("/statistics/files/weekly-stats", func(c *gin.Context) {

		// Logic to determine the start and end dates for the weekly intervals
		// This depends on how you store and can retrieve the creation dates of the files
		// For now, let's assume we have a function that gives us these dates
		startDate, endDate := getStartAndEndFileDates()

		var weeklyStats []WeeklyStats

		for weekStartDate := startDate; weekStartDate.Before(endDate); weekStartDate = weekStartDate.AddDate(0, 0, 7) {
			weekEndDate := weekStartDate.AddDate(0, 0, 7)
			if weekEndDate.After(endDate) {
				weekEndDate = endDate
			}

			// You need to implement the logic for calculating statistics for the week
			// This can involce aggregating data from your database queries
			usedStorage, err := query.CountPublicStorageUsed(weekStartDate.Format("2006-01-02"))
			if err != nil {
				log.Errorf("cannot get total public used storage: %s", err)
				AbortEntityNotFound(c)
				return
			}

			publicFiles, err := query.CountTotalFiles("public", weekStartDate.Format("2006-01-02"))
			if err != nil {
				log.Errorf("cannot get total public files: %s", err)
				AbortEntityNotFound(c)
				return
			}

			encryptedFiles, err := query.CountTotalFiles("encrypted", weekStartDate.Format("2006-01-02"))
			if err != nil {
				log.Errorf("cannot get total encrypted files: %s", err)
				AbortEntityNotFound(c)
				return
			}

			totalFiles := publicFiles + encryptedFiles

			weeklyStats = append(weeklyStats, WeeklyStats{
				Week:        weekStartDate.Format("2006-01-02"),
				UsedStorage: usedStorage,
				Total:       totalFiles,
				Public:      publicFiles,
				Encrypted:   encryptedFiles,
			})

		}

		c.JSON(http.StatusOK, weeklyStats)
	})

}
