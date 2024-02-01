package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type PublicFileUserShared struct {
	ID                   uint           `gorm:"primarykey"                   json:"id"`
	FileUID              string         `gorm:"type:varchar(42);uniqueIndex" json:"file_uid"`
	ShareHash            string         `gorm:"type:varchar(256)"            json:"share_hash"`
	Name                 string         `gorm:"type:varchar(1024);"          json:"name"`
	Mime                 string         `gorm:"type:varchar(256)"            json:"mime_type"`
	Size                 int64          `                                   json:"size"`
	CID                  string         `gorm:"type:varchar(64)"             json:"cid"`
	CIDOriginalDecrypted string         `gorm:"type:varchar(256)"            json:"cid_original_decrypted"`
}

func (PublicFileUserShared) TableName() string {
	return "public_files_user_shared"
}

func (m *PublicFileUserShared) Create() error {
	return db.Db().Create(m).Error
}

func (m *PublicFileUserShared) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

func (m *PublicFileUserShared) Save() error {
	return db.Db().Save(m).Error
}

func (m *PublicFileUserShared) Delete() error {
	return db.Db().Unscoped().Delete(m).Error
}
