package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/migrate"
)

func InitDb(opt migrate.Options) {
	if !db.HasDbProvider() {
		log.Error("migrate: no database provider")
		return
	}

	start := time.Now()

	Entities.Migrate(db.Db(), opt)

	log.Debugf("migrate: completed in %s", time.Since(start))
}

// TO-DO create InitTestDb
