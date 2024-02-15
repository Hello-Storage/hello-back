package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

// FileShareStates represents a file_share_state result set.
type FileShareStates []FileShareState

type FileShareState struct {
	ID         uint           `gorm:"primarykey"                          json:"id"`
	FileUID    string         `gorm:"type:varchar(42);uniqueIndex;references:UID;referencedTable:files"              json:"file_uid"`
	PublicFile PublicFile     `gorm:"foreignKey:FileUID;references:FileUID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"public_file"`
	CreatedAt  time.Time      `gorm:"index"                               json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index"                               json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                               json:"deleted_at"`
}

func (FileShareState) TableName() string {
	return "file_share_states"
}

func (m *FileShareState) Create() error {
	return db.Db().Create(m).Error
}

func (m *FileShareState) Save() error {
	return db.Db().Save(m).Error
}

// TxSave saves the file share state in a transaction
func (m *FileShareState) TxSave(tx *gorm.DB) error {
	return tx.Save(m).Error
}

func (m *FileShareState) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

// Delete also deletes the public file
func (m *FileShareState) Delete() error {
	//delete public file
	// Fin
	return db.Db().Delete(m).Error
}
