package entity

import (
	"fmt"

	"github.com/Hello-Storage/hello-back/internal/db"
	"gorm.io/gorm"
)

type Referral struct {
	ID           uint `gorm:"primarykey"                                         json:"id"`
	ReferrerID   uint `gorm:"index;not null;references:ID;referencedTable:users" json:"referrer_id"`
	ReferredID   uint `gorm:"index;not null;references:ID;referencedTable:users" json:"referred_id"`
	UserDetailID uint `gorm:"index;not null;references:ID;referencedTable:user_details" json:"user_detail_id"`
}

// TableName returns the entity table name.
func (Referral) TableName() string {
	return "referrals"
}

func (m *Referral) Create() error {
	return db.Db().Create(m).Error
}

func (m *Referral) TxCreate(tx *gorm.DB) error {
	return tx.Create(m).Error
}

func (m *Referral) Save() error {
	return db.Db().Save(m).Error
}

// Update a face property in the database.
func (m *Referral) Update(attr string, value interface{}) error {
	if m.ID == 0 {
		return fmt.Errorf("empty id")
	}

	return db.Db().Model(m).Update(attr, value).Error
}
