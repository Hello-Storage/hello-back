package entity

import (
	"fmt"
	"time"

	"github.com/Hello-Storage/hello-back/internal/migrate"
	"gorm.io/gorm"
)

type Tables map[string]interface{}

// Entities contains database entities and their table names.
var Entities = Tables{
	Error{}.TableName():                    &Error{},
	User{}.TableName():                     &User{},
	UserDetail{}.TableName():               &UserDetail{},
	UserLogin{}.TableName():                &UserLogin{},
	ReferredUser{}.TableName():             &ReferredUser{},
	Plan{}.TableName():                     &Plan{},
	Subscription{}.TableName():             &Subscription{},
	Email{}.TableName():                    &Email{},
	Wallet{}.TableName():                   &Wallet{},
	Github{}.TableName():                   &Github{},
	File{}.TableName():                     &File{},
	FileShareState{}.TableName():           &FileShareState{},
	PublicFile{}.TableName():               &PublicFile{},
	Folder{}.TableName():                   &Folder{},
	FileUser{}.TableName():                 &FileUser{},
	FolderUser{}.TableName():               &FolderUser{},
	Referral{}.TableName():                 &Referral{},
	PublicFileShareGroup{}.TableName():     &PublicFileShareGroup{},
	ShareGroup{}.TableName():               &ShareGroup{},
	ApiKey{}.TableName():                   &ApiKey{},
	ApiKeyFile{}.TableName():               &ApiKeyFile{},
	FileShareStatesUserShared{}.TableName(): &FileShareStatesUserShared{},
	PublicFileUserShared{}.TableName():     &PublicFileUserShared{},
	ArweaveTransaction{}.TableName():       &ArweaveTransaction{},
}

// WaitForMigration waits for the database migration to be successful.
func (list Tables) WaitForMigration(db *gorm.DB) {
	type RowCount struct {
		Count int
	}

	attempts := 100
	for name := range list {
		for i := 0; i <= attempts; i++ {
			count := RowCount{}
			if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) AS count FROM %s", name)).Scan(&count).Error; err == nil {
				log.Tracef("migrate: %s migrated", name)
				break
			} else {
				log.Tracef("migrate: waiting for %s migration (%s)", name, err.Error())
				time.Sleep(100 * time.Millisecond)
			}

			if i == attempts {
				panic("migration failed")
			}
		}
	}
}

// Truncate removes all data from tables without dropping them.
func (list Tables) Truncate(db *gorm.DB) {
	var name string

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (truncate)", r, name)
		}
	}()

	for name = range list {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1", name)).Error; err == nil {
			// log.Debugf("entity: removed all data from %s", name)
			break
		} else if err.Error() != "record not found" {
			log.Debugf("migrate: %s in %s", err, name)
		}
	}
}

// Migrate migrates all database tables of registered entities.
func (list Tables) Migrate(db *gorm.DB, opt migrate.Options) {
	var name string
	var entity interface{}

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (panic)", r, name)
		}
	}()

	log.Infof("migrate: running database migrations")

	// Run pre migrations, if any.
	if err := migrate.Run(db, opt.Pre()); err != nil {
		log.Error(err)
	}

	// Run ORM auto migrations.
	if opt.AutoMigrate {
		for name, entity = range list {
			if name == "users" || name == "wallets" || name == "githubs" {
				// if db.Migrator().HasTable(name) {
				continue
				// }
			}
			if err := db.AutoMigrate(entity); err != nil {
				log.Debugf("migrate: %s (waiting 1s)", err)

				time.Sleep(time.Second)

				if err = db.AutoMigrate(entity); err != nil {
					log.Errorf("migrate: failed migrating %s", name)
					panic(err)
				}
			}
		}
	}

	// Run main migrations, if any.
	if err := migrate.Run(db, opt); err != nil {
		log.Error(err)
	}
}

// Drop drops all database tables of registered entities.
func (list Tables) Drop(db *gorm.DB) {
	for _, entity := range list {
		if err := db.Migrator().DropTable(entity); err != nil {
			panic(err)
		}
	}
}
