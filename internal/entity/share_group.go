package entity

import (
	"github.com/Hello-Storage/hello-back/internal/db"
)

type ShareGroup struct {
	ID uint `gorm:"primarykey" json:"id"`
}

func (ShareGroup) TableName() string {
	return "share_group"
}

func (m *ShareGroup) Create() error {
	return db.Db().Create(m).Error
}

func (m *ShareGroup) Save() error {
	return db.Db().Save(m).Error
}

func (m *ShareGroup) Delete() error {
	return db.Db().Unscoped().Delete(m).Error
}
