package commands

import (
	"context"
	// "time"
	// "github.com/Hello-Storage/hello-back/internal/entity"

	"github.com/Hello-Storage/hello-back/internal/config"
	"github.com/Hello-Storage/hello-back/internal/event"
	"github.com/Hello-Storage/hello-back/internal/server"
)

var log = event.Log

func Start() {
	// init logger
	config.InitLogger()

	// load env
	err := config.LoadEnv()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// connect db and define enum types
	err = config.ConnectDB()
	if err != nil {
		log.Fatal("cannot connect to DB and create enums:", err)
	}

	config.InitDb()

	// connect redis
	// config.ConnectRedis()

	// Pass this context down the chain.
	cctx, cancel := context.WithCancel(context.Background())

	// This block is in case we want to delete files_groups records that are expired every 24 hours
	// // Schedule the timer to execute the function at regular intervals
	// ticker := time.NewTicker(24 * time.Hour) // Execute every 24 hours
	// defer ticker.Stop()

	// // Goroutine to execute the function in the background
	// go func() {
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			entity.DeleteExpiredShareGroups()
	// 		case <-cctx.Done():
	// 			return
	// 		}
	// 	}
	// }()

	server.Start(cctx)

	// Cancel the context when the server stops
	cancel()
}
