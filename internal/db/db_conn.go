package db

import (
	"fmt"
	slog "log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbConn *DbConn

// DbConn is a gorm.DB connection provider.
type DbConn struct {
	Driver string
	Dsn    string

	once sync.Once
	db   *gorm.DB
}

// Open creates a new gorm db connection.
func (g *DbConn) Open() {
	dbLogger := logger.New(
		slog.New(os.Stdout, "\r\n", slog.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(g.Dsn), &gorm.Config{Logger: dbLogger, SkipDefaultTransaction: true, NowFunc: func() time.Time {
		return time.Now().UTC()
	}})

	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			fmt.Printf("gorm.Open(%s, %s) %d\n", g.Driver, g.Dsn, i)
			db, err := gorm.Open(postgres.Open(g.Dsn), &gorm.Config{})

			if db != nil && err == nil {
				break
			} else {
				time.Sleep(5 * time.Second)
			}
		}

		if err != nil || db == nil {
			fmt.Println(err)
			log.Fatal(err)
		}
	}

	if err == nil {
		log.Infof("database: %s connected!", g.Dsn)
	}

	g.db = db
}

// SetDbProvider sets the Gorm database connection provider.
func SetDbProvider(conn *DbConn) {
	dbConn = conn
}

// HasDbProvider returns true if a db provider exists.
func HasDbProvider() bool {
	return dbConn.db != nil
}
