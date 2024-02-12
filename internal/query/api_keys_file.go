package query

import (
	"errors"

	"github.com/Hello-Storage/hello-back/internal/entity"
	"gorm.io/gorm"
)

// DeleteApiKeyFile deletes the ApiKeyFile record based on the provided FileUser.
func DeleteApiKeyFileByFileID(tx *gorm.DB, fileID uint) error {
	var apiKeyFile entity.ApiKeyFile

	result := tx.Where("file_id = ?", fileID).First(&apiKeyFile)

	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if err := tx.Delete(&apiKeyFile).Error; err != nil {
		return err
	}

	return nil
}
