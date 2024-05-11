package entity

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type InvestAccount struct {
	ID         uint           `gorm:"primarykey"                          json:"id"`
	IP         string         `gorm:"type:varchar(16);unique"                    json:"ip"`
	CreatedAt  time.Time      `                                           json:"created_at"`
	UpdatedAt  time.Time      `                                           json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                               json:"deleted_at"`
	Code       string         `gorm:"type:varchar(64);index"              json:"code"`
	InvestCode InvestCode     `gorm:"foreignKey:Code"     json:"-"`
}

// TableName returns the entity table name.
func (InvestAccount) TableName() string {
	return "invest_accounts"
}

func (m *InvestAccount) Create() error {
	return db.Db().Create(m).Error
}

func (m *InvestAccount) Save() error {
	return db.Db().Save(m).Error
}
