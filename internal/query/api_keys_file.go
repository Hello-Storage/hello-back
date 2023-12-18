package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
)

// DeleteApiKeyFile deletes the ApiKeyFile record based on the provided FileUser.
func DeleteApiKeyFile(f_u entity.FileUser) error {
	return db.Db().Table("api_key_files").
		Delete(nil).
		Where("file_id = ?", f_u.FileID).
		Error
}
