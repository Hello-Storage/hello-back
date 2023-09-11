package entity

import "github.com/Hello-Storage/hello-back/internal/db"

type Subscription struct {
	ID           uint `gorm:"primarykey"           json:"id"`
	PlanID       uint `gorm:"index" json:"plan_id"`
	UserID       uint `gorm:"index" json:"user_id"`
	UserDetailID uint
}

// TableName returns the entity table name.
func (Subscription) TableName() string {
	return "subscriptions"
}

func (m *Subscription) Create() error {
	return db.Db().Create(m).Error
}

func (m *Subscription) Save() error {
	return db.Db().Save(m).Error
}
