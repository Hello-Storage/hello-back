package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
)

type InvestCode struct {
	Code           string          `gorm:"primarykey"                      json:"code"`
	Email          string          `gorm:"type:varchar(64)"                json:"email"`
	SocialNetwork  string          `gorm:"type:varchar(64)"                json:"social_network"`
	InvestAccounts []InvestAccount `gorm:"foreignKey:Code;references:Code"`
}

// TableName returns the entity table name.
func (InvestCode) TableName() string {
	return "invest_codes"
}

func (m *InvestCode) Create() error {
	return db.Db().Create(m).Error
}

func (m *InvestCode) Save() error {
	return db.Db().Save(m).Error
}
