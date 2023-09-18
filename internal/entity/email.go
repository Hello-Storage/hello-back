package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
)

type Email struct {
	ID     uint   `gorm:"primarykey"       json:"id"`
	Email  string `gorm:"uniqueIndex;"     json:"email"`
	Secret string `gorm:"type:varchar(64)" json:"secret"`
	UserID uint
}

// TableName returns the entity table name.
func (Email) TableName() string {
	return "emails"
}

func (m *Email) Create() error {
	return db.Db().Create(m).Error
}

func (m *Email) Save() error {
	return db.Db().Save(m).Error
}
