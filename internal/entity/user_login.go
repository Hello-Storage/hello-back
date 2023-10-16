package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)


type UserLogins []UserLogin
// UserLogin represents the schema for the user login information.
type UserLogin struct {
	ID         uint           `gorm:"primarykey" json:"id"`               // Unique identifier
	LoginDate  time.Time      `gorm:"not null" json:"login_date"`         // Date and time of user login
	WalletAddr string         `gorm:"type:varchar(256);not null" json:"wallet_address"` // Wallet address of the user
	CreatedAt  time.Time      `json:"created_at"`                        // Timestamp of when the record was created
	UpdatedAt  time.Time      `json:"updated_at"`                        // Timestamp of when the record was last updated
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`            // Soft delete timestamp
}


// TableName returns the entity table name.
func (UserLogin) TableName() string {
	return "user_logins"
}

func (m *UserLogin) Create() error {
	return db.Db().Create(m).Error
}

func (m *UserLogin) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

func (m *UserLogin) Save() error {
	return db.Db().Save(m).Error
}
