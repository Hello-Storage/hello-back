package query

import (
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"gorm.io/gorm"
)

// DeleteFileUser
func DeleteFileUser(f_u entity.FileUser) error {
	return db.Db().Table("files").
		Where("id = ?", f_u.FileID).
		Update("deleted_at", time.Now()).
		Error
}

func DeleteFilePermission(user_uid uint, file_id uint) error {
	return db.Db().Table("files_users").
		Where("file_id = ? AND user_id = ?", file_id, user_uid).
		Update("permission", entity.DeletedPermission).
		Error
}

// GetNextOwner returns the ID of the next user who owns the file.
//
// @param userUID - The UID of the user who owns the file
// @param fileID - The ID of the file to look for
func GetNextOwner(userUID uint, fileID uint) (entity.FileUser, error) {
	var file entity.File
	if err := db.Db().Table("files").Where("id = ?", fileID).First(&file).Error; err != nil {
		return entity.FileUser{}, err
	}

	var files []entity.File
	if err := db.Db().Table("files").
		Where("c_id = ? AND deleted_at IS NULL", file.CID).
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		return entity.FileUser{}, err
	}

	var nextFileID uint
	if len(files) > 1 {
		nextFileID = files[1].ID
	} else {
		return entity.FileUser{}, nil
	}

	var fileUser entity.FileUser
	if err := db.Db().Table("files_users").
		Where("file_id = ?", nextFileID).
		First(&fileUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Handle the case where the file is not associated with any user
			return entity.FileUser{}, nil
		}
		return entity.FileUser{}, err
	}

	return fileUser, nil
}

func SetOwnerPermision(user_uid uint, file_id uint) error {
	return db.Db().Table("files_users").
		Where("file_id = ? AND user_id = ?", file_id, user_uid).
		Update("permission", entity.OwnerPermission).
		Error
}
