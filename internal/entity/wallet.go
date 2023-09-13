package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/pkg/rnd"
	"gorm.io/gorm"
)

type Wallet struct {
	ID         uint   `gorm:"primarykey"                            json:"id"`
	Type       string `gorm:"type:varchar(10);not null;default:eth" json:"type"`
	Address    string `gorm:"type:varchar(42);not null;uniqueIndex" json:"address"`
	PrivateKey string `gorm:"type:varchar(64);"                     json:"private_key"`
	Nonce      string `gorm:"type:varchar(16);not null"             json:"nonce"`
	Signature  string `gorm:"type:varchar(132);"                    json:"signature"`
	UserID     uint   `gorm:"uniqueIndex"`
}

// TableName returns the entity table name.
func (Wallet) TableName() string {
	return "wallets"
}

func (m *Wallet) Create() error {
	return db.Db().Create(m).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Wallet) BeforeCreate(db *gorm.DB) error {
	m.Nonce = rnd.GenerateRandomString(16)
	db.Statement.SetColumn("nonce", m.Nonce)
	return nil
}

func (m *Wallet) Save() error {
	return db.Db().Save(m).Error
}
