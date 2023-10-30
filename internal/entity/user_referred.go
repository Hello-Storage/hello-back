package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

// custom referred user from specific marketing agencies
type ReferredUser struct {
	ID         uint `gorm:"primarykey" json:"id"`
	ReferredID uint `json:"referred_id"`
	Referrer   string `json:"referrer"`
	CreatedAt time.Time      `                                    json:"created_at"`
	UpdatedAt time.Time      `                                    json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                        json:"deleted_at"`
}

// TableName returns the entity table name.
func (ReferredUser) TableName() string {
	return "referred_users"
}

func (m *ReferredUser) Create() error {
	return db.Db().Create(m).Error
}

func (m *ReferredUser) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}
