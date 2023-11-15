package api

import (
	"net/http"
	"time"

	"github.com/Hello-Storage/hello-back/internal/query"
	"github.com/gin-gonic/gin"
)

type WeeklyUserStats struct {
	Week       string `json:"week"`
	TotalUsers int    `json:"total_users"`
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
		startDate, endDate := getStartAndEndUserDates()

		var weeklyStats []WeeklyUserStats
		currentDate := time.Now();

		for weekStartDate := startDate; weekStartDate.Before(endDate); weekStartDate = weekStartDate.AddDate(0, 0, 7) {
			weekEndDate := weekStartDate.AddDate(0, 0, 6)
			if weekEndDate.After(currentDate) {
				continue
			}

			if weekEndDate.After(endDate) {
				weekEndDate = endDate
			}

			totalUsers, err := query.CountTotalUsers(weekStartDate.Format("2006-01-02"))
			if err != nil {
				log.Errorf("cannot get total users: %s", err)
				AbortInternalServerError(c)
				return
			}

			weeklyStats = append(weeklyStats, WeeklyUserStats{
				Week:       weekStartDate.Format("2006-01-02"),
				TotalUsers: totalUsers,
			})
		}

		c.JSON(http.StatusOK, weeklyStats)
	})
}
