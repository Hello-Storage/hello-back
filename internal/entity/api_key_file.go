package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type ApiKeyFile struct {
	ID     uint `gorm:"primarykey"           json:"id"`
	FileID uint `gorm:"index;column:file_id" json:"file_id"`
	UserID uint `gorm:"index;column:user_id" json:"user_id"`
}

// TableName returns the entity table name.
func (ApiKeyFile) TableName() string {
	return "api_key_files"
}

func (m *ApiKeyFile) Create() error {
	return db.Db().Create(m).Error
}

func (m *ApiKeyFile) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

// update
func (m *ApiKeyFile) Update() error {
	return db.Db().Model(m).Updates(m).Error
}
