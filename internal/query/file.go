package query

import (
	"fmt"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
)

// FileByUID returns file for the given UID.
func FindFileByUID(uid string) (*entity.File, error) {
	f := entity.File{}

	if uid == "" {
		return &f, fmt.Errorf("file uid required")
	}

	err := db.Db().Where("uid = ?", uid).First(&f).Error

	return &f, err
}

// FilesByRoot return files in a given folder root.
func FindFilesByRoot(root string) (files entity.Files, err error) {
	if err := db.Db().Where("root = ?", root).Find(&files).Error; err != nil {
		return files, err
	}

	return files, err
}

// Count all files overall
func CountFiles() (upfile int64, err error) {
	if err := db.Db().Table("files").Count(&upfile).Error; err != nil {
		return upfile, err
	}

	return upfile, nil
}

// Count total public files in database
func CountPublicFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("status = 'public'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// count total encrypted files in database
func CountEncryptedFiles() (encryptedfiles int64, err error) {
	if err := db.Db().Table("files").Where("status = 'encrypted'").Count(&encryptedfiles).Error; err != nil {
		return encryptedfiles, err
	}

	return encryptedfiles, nil
}

func FindRootFilesByUser(user_id uint) (files entity.Files, err error) {
	if err := db.Db().
		Table("files").
		Joins("LEFT JOIN files_users on files_users.file_id = files.id").
		Where("files.root = '/' AND files_users.permission = 'owner' AND files_users.user_id = ?", user_id).
		Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// DeleteFileByUID deletes a file by its UID.
func DeleteFileByUID(file_uid string) error {
	if file_uid == "" {
		return fmt.Errorf("file uid required")
	}

	return db.Db().Where("uid = ?", file_uid).Delete(&entity.File{}).Error
}

// query of the average size of all the files among the users
func CountMediumSizeFiles() (msize int64, err error) {
	if err := db.Db().Table("files").Select("ROUND(AVG(size))").Scan(&msize).Error; err != nil {
		return msize, err
	}

	return msize, nil

}

// query total sum storaged_used of all users
func CountTotalUsedStorage() (totalusedstorage int64, err error) {
	if err := db.Db().Table("user_details").Select("SUM(storage_used)").Scan(&totalusedstorage).Error; err != nil {
		return totalusedstorage, err
	}

	return totalusedstorage, nil
}
