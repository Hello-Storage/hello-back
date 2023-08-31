package entity

import "github.com/Hello-Storage/hello-back/internal/db"

// FolderUser represents a one-to-many relation between File and User.

type FolderUser struct {
	ID         uint       `gorm:"primarykey"`
	FolderID   uint       `gorm:"index;column:folder_id" json:"folder_id"`
	UserID     uint       `gorm:"index;column:user_id"   json:"user_id"`
	Permission permission `gorm:"not null;"              json:"permission"`
}

// TableName returns the entity table name.
func (FolderUser) TableName() string {
	return "folders_users"
}

func (m *FolderUser) Create() error {
	return db.Db().Create(m).Error
}

func (m *FolderUser) Save() error {
	return db.Db().Save(m).Error
}

// update
func (m *FolderUser) Update() error {
	log.Info(m)
	return db.Db().Model(m).Updates(m).Error // <- esto de aqui hay algo que da error
}
