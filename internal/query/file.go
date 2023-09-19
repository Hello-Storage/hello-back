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

// Query daily storage used by all users in the last 24 hours
// func CountDailyStorage(daystring string) (dailystorage int64, err error) {
// 	log.Infof("daystring: %s", daystring)

// 	query := db.Db().Table("files").Select("SUM(size)")

// 	// Apply the date range filter
// 	query = query.Where("created_at >= DATE_TRUNC('DAY', TIMESTAMP ?) AND created_at < DATE_TRUNC('DAY', TIMESTAMP ?) + INTERVAL '1 DAY'", daystring, "2023-09-14 17:52:29")

// 	// Execute and scan the result
// 	if err := query.Scan(&dailystorage).Error; err != nil {
// 		return dailystorage, err
// 	}

// 	return dailystorage, nil
// }

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

// query for count all txt files
func CountTxtFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("media_type = 'text/plain' AND status = 'public'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query for count all public png files
func CountPngFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("media_type = 'image/png' AND status = 'public'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query for count all jpg files
func CountJpgFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("media_type = 'image/jpg' AND status = 'public'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query for count all pdf files
func CountPdfFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("media_type = 'application/pdf' AND status = 'public'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// Query daily storage used by all users in the last 24 hours
// func CountDailyStorage(daystring string) (dailystorage int64, err error) {
// 	log.Infof("daystring: %s", daystring)

// 	query := db.Db().Table("files").Select("SUM(size)")

// 	// Apply the date range filter
// 	query = query.Where("created_at >= DATE_TRUNC('DAY', TIMESTAMP ?) AND created_at < DATE_TRUNC('DAY', TIMESTAMP ?) + INTERVAL '1 DAY'", daystring, "2023-09-14 17:52:29")

// 	// Execute and scan the result
// 	if err := query.Scan(&dailystorage).Error; err != nil {
// 		return dailystorage, err
// 	}

// 	return dailystorage, nil
// }
