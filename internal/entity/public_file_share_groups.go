package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type PublicFileShareGroup struct {
	ID           uint   `gorm:"primarykey"                   json:"id"`
	ShareGroupID uint   `json:"share_group_id"`
	ShareHash    string `json:"share_hash"`
}

func (PublicFileShareGroup) TableName() string {
	return "public_file_share_group"
}

func (m *PublicFileShareGroup) Create() error {
	return db.Db().Create(m).Error
}

func (m *PublicFileShareGroup) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

func (m *PublicFileShareGroup) Save() error {
	return db.Db().Save(m).Error
}

func (m *PublicFileShareGroup) Delete() error {
	return db.Db().Unscoped().Delete(m).Error
}
