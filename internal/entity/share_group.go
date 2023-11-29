package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
)

type ShareGroup struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Hash string `json:"hash" gorm:"unique;not null"`
}

func (ShareGroup) TableName() string {
	return "share_group"
}

func (m *ShareGroup) Create() error {
	hash, err := generateShareGroupHash()
	if err != nil {
		return err
	}

	m.Hash = hash

	return db.Db().Create(m).Error
}

func generateShareGroupHash() (string, error) {
	uniqueValue := fmt.Sprintf("%d", time.Now().UnixNano())

	hasher := sha256.New()
	hasher.Write([]byte(uniqueValue))
	hashBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashBytes), nil
}

func (m *ShareGroup) Save() error {
	return db.Db().Save(m).Error
}

func (m *ShareGroup) Delete() error {
	return db.Db().Unscoped().Delete(m).Error
}
