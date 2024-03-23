package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
)

type ArweaveTransaction struct {
	Id      string    `gorm:"primarykey" json:"transaction_id"`
	Owner   string    `gorm:"type:varchar(64)" json:"transaction_owner"`
	Message string    `gorm:"type:varchar(64)" json:"transaction_message"`
	Date    time.Time `gorm:"type:timestamp" json:"transaction_date"`
}

// TableName returns the entity name
func (ArweaveTransaction) TableName() string {
	return "arweave_transactions"
}

func (m *ArweaveTransaction) Create() error {
	return db.Db().Create(m).Error
}

func (m *ArweaveTransaction) Save() error {
	return db.Db().Save(m).Error
}