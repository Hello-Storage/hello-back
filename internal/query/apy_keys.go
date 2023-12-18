package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
)

// FindApiKeyByUserID finds an API key by the user ID.
func FindApiKeyByUserID(userID uint) (*entity.ApiKey, error) {
	var apiKey entity.ApiKey
	if err := db.Db().Where("user_id = ?", userID).First(&apiKey).Error; err != nil {
		return nil, err
	}
	return &apiKey, nil
}
