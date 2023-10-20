package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

// FileShareStates represents a file_share_state result set.
type FileShareStates []FileShareState

type FileShareState struct {
	ID      uint   `gorm:"primarykey"                          json:"id"`
	FileUID string `gorm:"type:varchar(42);uniqueIndex"              json:"file_uid"`

	PublicFile PublicFile `gorm:"foreignKey:FileUID;references:FileUID"` // explicitly defining foreign key and references
	//PublicFileID uint           `gorm:"type:integer" json:"public_file_id"`
	CreatedAt time.Time `gorm:"index"                               json:"created_at"`
	UpdatedAt time.Time `gorm:"index"                               json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                               json:"deleted_at"`
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

func (m *FileShareState) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}
