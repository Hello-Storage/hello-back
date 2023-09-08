package entity

import (
	"fmt"

	"github.com/Hello-Storage/hello-back/internal/db"
)

type Referrals struct {
	ID          uint `gorm:"primarykey" json:"id"`
	ReferrerID  uint `gorm:"index;column:referrer_id" json:"referrer_id"`
	RefferredID uint `gorm:"index;column:referred_id" json:"referred_id"`
	UserDetailID uint
}

// TableName returns the entity table name.
func (Referrals) TableName() string {
	return "referrals"
}

func (m *Referrals) Create() error {
	return db.Db().Create(m).Error
}

func (m *Referrals) Save() error {
	return db.Db().Save(m).Error
}

// Update a face property in the database.
func (m *Referrals) Update(attr string, value interface{}) error {
	if m.ID == 0 {
		return fmt.Errorf("empty id")
	}

	return db.Db().Model(m).Update(attr, value).Error
}
