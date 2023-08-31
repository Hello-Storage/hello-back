package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/pkg/rnd"
	"gorm.io/gorm"
)

const (
	FolderUID = byte('d')
)

type Folders []Folder

type Folder struct {
	ID        uint           `gorm:"primarykey"                          json:"id"`
	UID       string         `gorm:"type:varchar(42);uniqueIndex;"       json:"uid"`
	Title     string         `gorm:"type:varchar(255);"                  json:"title"`
	Path      string         `gorm:"type:varchar(1024);default:'/';"     json:"path"` // folderA/folderB/***
	Root      string         `gorm:"type:varchar(42);index;default:'/';" json:"root"` // parent folder uid
	CreatedAt time.Time      `                                           json:"created_at"`
	UpdatedAt time.Time      `                                           json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                               json:"deleted_at"`
}

// TableName returns the entity table name.
func (Folder) TableName() string {
	return "folders"
}

func (m *Folder) Create() error {
	return db.Db().Create(m).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Folder) BeforeCreate(db *gorm.DB) error {
	if rnd.IsUnique(m.UID, 'd') {
		return nil
	}
	m.UID = rnd.GenerateUID(FolderUID)
	db.Statement.SetColumn("UID", m.UID)

	return nil
}

func (m *Folder) FirstOrCreateFolderByTitleAndRoot() *Folder {
	result := Folder{}

	if err := db.Db().Where("title = ? AND root = ?", m.Title, m.Root).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("Folder: %s", err)
		return nil
	}

	return m
}

// update
func (m *Folder) UpdateRootOnly() error {
	return db.Db().Model(m).Where("UID = ?", m.UID).Update("Root", m.Root).Error

}

// IsFolderOwner checks if a user is the owner of a folder
func IsFolderOwner(folderID uint, userID uint) (bool, error) {
	var count int64
	err := db.Db().Table("folders_users").
		Where("folder_id = ? AND user_id = ? AND permission = ?", folderID, userID, OwnerPermission).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
