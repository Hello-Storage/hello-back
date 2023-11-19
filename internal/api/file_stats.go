package api

import (
	"fmt"
	"net/http"
	"sync"
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

type SharedStatistics struct {
	Statistics Statistics
	timer      *time.Timer
}

var (
	statisticsInstance *SharedStatistics
	statisticsOnce     sync.Once
	statisticsMutex    sync.Mutex
)

func GetStatisticsInstance() *SharedStatistics {
	statisticsOnce.Do(func() {
		statisticsInstance = &SharedStatistics{}
		statisticsInstance.startStatisticsBackgroundJob()
	})
	return statisticsInstance
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

func (s *SharedStatistics) startStatisticsBackgroundJob() {
	// Stop any existing timer
	if s.timer != nil {
		s.timer.Stop()
	}

	// Start or restart the timer
	s.timer = time.NewTimer(1 * time.Minute)

	go func() {
		for {
			<-s.timer.C
			newStats, err := s.CalculateStatistics()
			if err != nil {
				log.Errorf("cannot calculate weekly stats: %s", err)
				return
			}
			statisticsMutex.Lock()
			s.Statistics = newStats
			statisticsMutex.Unlock()
			s.timer.Reset(1 * time.Minute) // Reset the timer for the next interval
		}
	}()
}

func (s *SharedStatistics) stopStatisticsBackgroundJob() {
	// Perform any necessary cleanup here
	// Reset the instance to allow for a fresh start on the next request
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	statisticsOnce = sync.Once{}
}

func (s *SharedStatistics) CalculateStatistics() (Statistics, error) {
	totalusedstorage, err := query.CountTotalUsedStorage()
	if err != nil {
		return Statistics{}, fmt.Errorf("cannot get total used storage: %s", err)
	}

	totalusers, err := query.CountUsers()
	if err != nil {
		return Statistics{}, fmt.Errorf("cannot get total users: %s", err)
	}

	upfile, err := query.CountFiles()
	if err != nil {
		return Statistics{}, fmt.Errorf("cannot get total uploaded files: %s", err)
	}

	//medium size files
	msize := totalusedstorage / upfile

	encryptedfiles, err := query.CountEncryptedFiles()
	if err != nil {
		return Statistics{}, fmt.Errorf("cannot get total encrypted files: %s", err)
	}

	publicfiles, err := query.CountPublicFiles()
	if err != nil {
		return Statistics{}, fmt.Errorf("cannot get total public files: %s", err)
	}

	publicfolders, err := query.CountPublicFolders()
	if err != nil {
		return Statistics{}, fmt.Errorf("cannot get total public folders: %s", err)
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
		// CountDailyStorage:    daylystorage,
	}

	statisticsMutex.Lock()
	s.Statistics = stats
	statisticsMutex.Unlock()

	return stats, nil
}

func GetStatistics(router *gin.RouterGroup) {

	router.GET("/statistics", func(c *gin.Context) {
		sharedData := GetStatisticsInstance()
		statisticsMutex.Lock()
		defer statisticsMutex.Unlock()

		c.JSON(http.StatusOK, sharedData.Statistics)
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

type WeeklyStorageStats struct {
	Week        string `json:"week"`
	UsedStorage int64  `json:"usedStorage"`
}

type SharedStorageData struct {
	WeeklyStatistics []WeeklyStorageStats
	timer            *time.Timer
}

var (
	storageInstance *SharedStorageData
	storageOnce     sync.Once
	storageMutex    sync.Mutex
)

func GetStorageInstance() *SharedStorageData {
	storageOnce.Do(func() {
		storageInstance = &SharedStorageData{}
		storageInstance.startStorageBackgroundJob()
	})
	return storageInstance
}

func (s *SharedStorageData) CalculateWeeklyStorageStats() ([]WeeklyStorageStats, error) {

	// Logic to determine the start and end dates for the weekly intervals

	// CHANGES TO DO:
	// Same logic as before but without Gin context and AbortEntityNotFound
	// Return the result instead of modifying the shared data directly
	// Log errors instead of aborting the request
	startDate, endDate := getStartAndEndFileDates()

	var weeklyStats []WeeklyStorageStats

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
			return nil, fmt.Errorf("cannot get total public used storage: %s", err)
		}

		weeklyStats = append(weeklyStats, WeeklyStorageStats{
			Week:        weekStartDate.Format("2006-01-02"),
			UsedStorage: usedStorage,
		})

	}

	storageMutex.Lock()
	s.WeeklyStatistics = weeklyStats
	storageMutex.Unlock()

	return weeklyStats, nil
}



func (s *SharedStorageData) startStorageBackgroundJob() {
	// Stop any existing timer
	if s.timer != nil {
		s.timer.Stop()
	}

	s.timer = time.NewTimer(1 * time.Minute)

	go func() {
		for {
			<-s.timer.C
			newStats, err := s.CalculateWeeklyStorageStats()
			if err != nil {
				log.Errorf("cannot calculate weekly stats: %s", err)
				return
			}
			storageMutex.Lock()
			s.WeeklyStatistics = newStats
			storageMutex.Unlock()
			log.Print("getting file stats")
			s.timer.Reset(1 * time.Minute) // Reset the timer for the next interval
		}
	}()
}

func GetWeeklyPublicStats(router *gin.RouterGroup) {
	router.GET("/statistics/files/weekly-stats", func(c *gin.Context) {

		sharedData := GetStorageInstance()
		storageMutex.Lock()
		statsCopy := make([]WeeklyStorageStats, len(sharedData.WeeklyStatistics))
		copy(statsCopy, sharedData.WeeklyStatistics)
		storageMutex.Unlock()

		c.JSON(http.StatusOK, sharedData.WeeklyStatistics)
	})

}
