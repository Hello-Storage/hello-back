package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type FileShareStateUserShared struct {
	ID         uint                 `gorm:"primarykey"                          json:"id"`
	FileUID    string               `gorm:"type:varchar(42)"              json:"file_uid"`
	UserID     uint                 `gorm:"type:int" json:"user_id"`
	PublicFile PublicFileUserShared `gorm:"foreignKey:FileUID;references:FileUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"public_file"`
}

func (FileShareStateUserShared) TableName() string {
	return "file_share_states_user_shared"
}

func (m *FileShareStateUserShared) Create() error {
	return db.Db().Create(m).Error
}

func (m *FileShareStateUserShared) Save() error {
	return db.Db().Save(m).Error
}

func (m *FileShareStateUserShared) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

func (m *FileShareStateUserShared) Delete() error {
	return db.Db().Delete(m).Error
}
