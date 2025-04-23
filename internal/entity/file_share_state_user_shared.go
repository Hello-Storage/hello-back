package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type FileShareStatesUserShared struct {
	ID                   uint                 `gorm:"primarykey"                          json:"id"`
	FileUID              string               `gorm:"type:varchar(42);uniqueIndex;references:UID;referencedTable:files" json:"file_uid"`
	UserID               uint                 `gorm:"type:int" json:"user_id"`
	PublicFileUserShared PublicFileUserShared `gorm:"foreignKey:FileUID;references:FileUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"public_files_user_shared"`
}

func (FileShareStatesUserShared) TableName() string {
	return "file_share_states_user_shared"
}

func (m *FileShareStatesUserShared) Create() error {
	return db.Db().Create(m).Error
}

func (m *FileShareStatesUserShared) Save() error {
	return db.Db().Save(m).Error
}

func (m *FileShareStatesUserShared) TxSave(tx *gorm.DB) error {
	return tx.Save(m).Error
}

func (m *FileShareStatesUserShared) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

func (m *FileShareStatesUserShared) Delete() error {
	return db.Db().Delete(m).Error
}

func (m *FileShareStatesUserShared) TxDelete(tx *gorm.DB) error {
	return tx.Delete(m).Error
}
