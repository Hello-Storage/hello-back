package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/gin-gonic/gin"
)

type WeeklyUsersStats struct {
	Week       string `json:"week"`
	TotalUsers int    `json:"total_users"`
}

type SharedUsersData struct {
	WeeklyStatistics []WeeklyUsersStats
	timer            *time.Timer
}

var (
	usersInstance *SharedUsersData
	usersOnce     sync.Once
	usersMutex    sync.Mutex
)

func GetUsersInstance() *SharedUsersData {
	usersOnce.Do(func() {
		usersInstance = &SharedUsersData{}
		usersInstance.startUsersBackgroundJob()
	})
	return usersInstance
}

func (s *SharedUsersData) calculateWeeklyStats() ([]WeeklyUsersStats, error) {
	startDate, endDate := getStartAndEndUserDates()

	var weeklyStats []WeeklyUsersStats

	for weekStartDate := startDate; weekStartDate.Before(endDate); weekStartDate = weekStartDate.AddDate(0, 0, 7) {
		weekEndDate := weekStartDate.AddDate(0, 0, 7)

		if weekEndDate.After(endDate) {
			weekEndDate = endDate
		}

		totalUsers, err := query.CountTotalUsers(weekStartDate.Format("2006-01-02"))
		if err != nil {
			return nil, fmt.Errorf("cannot get total users: %s", err)
		}

		weeklyStats = append(weeklyStats, WeeklyUsersStats{
			Week:       weekStartDate.Format("2006-01-02"),
			TotalUsers: totalUsers,
		})
	}

	usersMutex.Lock()
	s.WeeklyStatistics = weeklyStats
	usersMutex.Unlock()

	return weeklyStats, nil
}

func (s *SharedUsersData) startUsersBackgroundJob() {
	// Stop any existing timer
	if s.timer != nil {
		s.timer.Stop()
	}

	// Start or restart the timer
	s.timer = time.AfterFunc(1*time.Minute, func() {
		usersMutex.Lock()
		defer usersMutex.Unlock()
		s.stopUsersBackgroundJob()
	})

	go func() {
		for {
			select {
			case <-s.timer.C:
				return // Exit the goroutine when timer expires
			default:
				newStats, err := s.calculateWeeklyStats()
				if err != nil {
					log.Errorf("cannot calculate weekly stats: %s", err)
					return
				}
				usersMutex.Lock()
				s.WeeklyStatistics = newStats
				usersMutex.Unlock()
				time.Sleep(1 * time.Second) // Example delay
			}
		}
	}()
}

func (s *SharedUsersData) stopUsersBackgroundJob() {
	// Perform any necessary cleanup here
	// Reset the instance to allow for a fresh start on the next request
	usersInstance = nil
	usersOnce = sync.Once{}
}

func getStartAndEndUserDates() (time.Time, time.Time) {
	// You need to implement the logic for calculating the start and end dates for the weekly intervals
	// This can involve querying your database for the earliest and latest creation dates of files
	// For now, let's assume we have a function that gives us these dates
	startDate, endDate, err := query.GetStartAndEndUserDatesPublic()
	if err != nil {
		log.Errorf("cannot get start and end dates for files: %s", err)
		return time.Time{}, time.Time{}
	}

	return startDate, endDate
}

func GetWeeklyUserStats(router *gin.RouterGroup) {
	router.GET("/statistics/users/weekly-stats", func(c *gin.Context) {
		sharedData := GetUsersInstance()
		usersMutex.Lock()
		statsCopy := make([]WeeklyUsersStats, len(sharedData.WeeklyStatistics))
		copy(statsCopy, sharedData.WeeklyStatistics)
		usersMutex.Unlock()

		c.JSON(http.StatusOK, sharedData.WeeklyStatistics)
	})
}
