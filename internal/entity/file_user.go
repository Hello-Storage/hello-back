package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type permission string

const (
	OwnerPermission  permission = "owner"
	SharedPermission permission = "shared"
)

// FileUser represents a one-to-many relation between File and User.

type FileUser struct {
	ID         uint       `gorm:"primarykey"           json:"id"`
	FileID     uint       `gorm:"index;column:file_id" json:"file_id"`
	UserID     uint       `gorm:"index;column:user_id" json:"user_id"`
	Permission permission `gorm:"not null;"            json:"permission"`
}

// TableName returns the entity table name.
func (FileUser) TableName() string {
	return "files_users"
}

func (m *FileUser) Create() error {
	return db.Db().Create(m).Error
}

func (m *FileUser) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

//update
func (m *FileUser) Update() error {
	return db.Db().Model(m).Updates(m).Error
}