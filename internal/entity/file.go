package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/pkg/media"
	"github.com/Hello-Storage/hello-back/pkg/rnd"
	"gorm.io/gorm"
)

const (
	FileUID = byte('f')
)

// Files represents a file result set.
type Files []File

type File struct {
	ID        uint           `gorm:"primarykey"                          json:"id"`
	UID       string         `gorm:"type:varchar(42);uniqueIndex;"       json:"uid"`
	Name      string         `gorm:"type:varchar(1024);"                 json:"name"`
	Root      string         `gorm:"type:varchar(42);index;default:'/';" json:"root"` // parent folder uid
	Mime      string         `gorm:"type:varchar(64)"                    json:"mime_type"`
	Size      int64          `                                           json:"size"`
	MediaType string         `gorm:"type:varchar(16)"                    json:"media_type"`
	CreatedAt time.Time      `                                           json:"created_at"`
	UpdatedAt time.Time      `                                           json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                               json:"deleted_at"`
	Path      string         `gorm:"type:varchar(1024);"                 json:"path"` // full path
}

// TableName returns the entity table name.
func (File) TableName() string {
	return "files"
}

func (m *File) Create() error {
	return db.Db().Create(m).Error
}

func (m *File) Save() error {
	return db.Db().Save(m).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *File) BeforeCreate(db *gorm.DB) error {
	// Set MediaType based on FileName if empty.
	if m.MediaType == "" && m.Name != "" {
		m.MediaType = media.FromName(m.Name).String()
	}

	// Return if uid exists.
	if rnd.IsUnique(m.UID, FileUID) {
		return nil
	}

	m.UID = rnd.GenerateUID(FileUID)
	db.Statement.SetColumn("UID", m.UID)

	return nil
}

func (m *File) FirstOrCreateFile() *File {
	result := File{}

	if err := db.Db().Where("uid = ?", m.UID).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("file: %s", err)
		return nil
	}

	return m
}

// update
func (m *File) UpdateRootOnly() error {
	return db.Db().Model(m).Where("UID = ?", m.UID).Update("Root", m.Root).Error
}

// IsFolderOwner checks if a user is the owner of a folder
func IsFileOwner(folderID uint, userID uint) (bool, error) {
	var count int64
	err := db.Db().Table("files_users").
		Where("file_id = ? AND user_id = ? AND permission = ?", folderID, userID, OwnerPermission).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
