package query

import (
	"errors"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"gorm.io/gorm"
)

// DeleteApiKeyFile deletes the ApiKeyFile record based on the provided FileUser.
func DeleteApiKeyFileByFileID(fileID uint) error {
	var apiKeyFile entity.ApiKeyFile

	result := db.Db().Where("file_id = ?", fileID).First(&apiKeyFile)

	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if err := db.Db().Delete(&apiKeyFile).Error; err != nil {
		return err
	}

	return nil
}
