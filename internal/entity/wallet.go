package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/pkg/rnd"
	"gorm.io/gorm"
)

type AccountType string

const (
	Provider AccountType = "provider"
	Google   AccountType = "google"
	GitHub   AccountType = "github"
)

type Wallet struct {
	ID         uint   `gorm:"primarykey"                            json:"id"`
	Address    string `gorm:"type:varchar(50);not null;uniqueIndex" json:"address"`
	Type       string `gorm:"type:account_type;not null;default:'provider'" json:"type"`
	PrivateKey []byte `gorm:"type:bytea;" json:"private_key"`
	Nonce      string `gorm:"type:varchar(16);not null"             json:"nonce"`
	UserID     uint
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
