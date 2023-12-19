package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
)

type ApiKey struct {
	ID          uint      `gorm:"primarykey"       json:"id"`
	UserID      uint      `gorm:"index;column:user_id" json:"user_id"`
	ApiKey      string    ` json:"api_key"`
	KeyRequests int       `json:"key_requests"`
	CreatedAt   time.Time `gorm:"index"                               json:"created_at"`
}

// TableName returns the entity table name.
func (ApiKey) TableName() string {
	return "api_keys"
}

func (m *ApiKey) Create() error {
	return db.Db().Create(m).Error
}

func (m *ApiKey) Save() error {
	return db.Db().Save(m).Error
}

func (m *ApiKey) IncrementKeyRequests() error {
	m.KeyRequests++
	return m.Save()
}
