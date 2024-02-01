package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Hello-Storage/hello-back/internal/api"
	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	start := time.Now()

	gin.SetMode(gin.DebugMode)

	// Create new HTTP router engine without standard middleware.
	router := gin.New()

	ProtectedRouter := gin.New()
	PublicRouter := gin.New()

	router.MaxMultipartMemory = 500 << 20

	//cors protection
	ApiKeyAPIv1CorsConfig := cors.New(cors.Config{
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowAllOrigins: true,
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Cross-Origin-Opener-Policy",
			"api_key",
		},
		MaxAge: 12 * time.Hour,
	})
	protectedCorsConfig := cors.New(cors.Config{
		AllowOrigins: []string{
			//development
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			//production
			"https://joinhello.app",
			"https://staging.joinhello.app",
			"https://www.staging.joinhello.app",
			"https://www.joinhello.app",
			"https://joinhello.vercel.app",
			"https://www.joinhello.vercel.app",
			"https://hello.storage",
			"https://www.hello.storage",
			"https://space.hello.app",
			"https://hello.app",
			"https://www.hello.app",
			"https://stats.hello.app",
			"https://www.stats.hello.app",
			"https://www.space.hello.app",
			"https://space.hello.storage",
			"https://www.space.hello.storage",
			"https://space.hello.ws",
			"https://www.space.hello.ws",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Cross-Origin-Opener-Policy",
			"Authorization",
		},
		AllowCredentials: false,
		AllowOriginFunc: func(origin string) bool {
			return strings.Contains(origin, "hello-storage.vercel.app")
		},
		MaxAge: 12 * time.Hour,
	})
	ProtectedRouter.Use(protectedCorsConfig)
	PublicRouter.Use(ApiKeyAPIv1CorsConfig)
	// Register common middleware.
	router.Use(gin.Recovery(), Logger(), middlewares.RateLimitMiddleware(100, 100))
	log.Info("server: common middleware registered")

	router.Any("/api/*action", gin.WrapH(ProtectedRouter))
	router.Any("/public-api/*action", gin.WrapH(PublicRouter))

	config.LoadEnv()

	// Register HTTP route handlers.
	registerRoutes(ProtectedRouter)
	RegisterApiRoutes(PublicRouter)

	log.Infof("port: %s", config.Env().AppPort)
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", "0.0.0.0", config.Env().AppPort),
		Handler:        router,
		ReadTimeout:    10 * time.Minute,
		WriteTimeout:   10 * time.Minute,
		MaxHeaderBytes: 5 << 30, // 5 GB
	}
	log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))
	go StartHttp(server)

	usersData := api.GetUsersInstance()
	storageData := api.GetStorageInstance()
	statisticsData := api.GetStatisticsInstance()

	initialUsersStats, err := usersData.CalculateWeeklyUsersStats()
	if err != nil {
		log.Errorf("cannot calculate initial weekly user stats: %s", err)
	} else {
		usersData.WeeklyStatistics = initialUsersStats
		log.Println("Calculated initial weekly user stats")
	}
	initialStorageStats, err := storageData.CalculateWeeklyStorageStats()
	if err != nil {
		log.Errorf("cannot calculate initial weekly storage stats: %s", err)
	} else {
		storageData.WeeklyStatistics = initialStorageStats
		log.Println("Calculated initial weekly storage stats")
	}
	initialStatistics, err := statisticsData.CalculateStatistics()
	if err != nil {
		log.Errorf("cannot calculate initial weekly stats: %s", err)
	} else {
		statisticsData.Statistics = initialStatistics
		log.Println("Calculated initial stats")
	}

	// Graceful HTTP server shutdown.
	<-ctx.Done()
	log.Info("server: shutting down")
	err = server.Close()
	if err != nil {
		log.Errorf("server: shutdown failed (%s)", err)
	}
}

// StartHttp starts the web server in http mode.
func StartHttp(s *http.Server) {
	if err := s.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Info("server: shutdown complete")
		} else {
			log.Errorf("server: %s", err)
		}
	}
}
