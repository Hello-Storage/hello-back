package query

import (
	"errors"
	"fmt"
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"gorm.io/gorm"
)

// FindFileByUID returns file for the given UID.
func FindFileByUID(uid string) (*entity.File, error) {
	if uid == "" {
		return nil, fmt.Errorf("file uid required")
	}

	var file entity.File
	err := db.Db().Where("uid = ?", uid).First(&file).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("file not found for UID: %s", uid)
		}
		return nil, fmt.Errorf("failed to find file: %v", err)
	}

	return &file, nil
}

// FileByUID returns file for the given UID.
func FindFileByID(id uint) (*entity.File, error) {
	f := &entity.File{}

	err := db.Db().Where("id = ?", id).First(f).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return f, nil
}

// FindFileByCID returns file for the given CID.
func FindFileByCID(cid string) (*entity.File, error) {
	f := &entity.File{}

	err := db.Db().Where("cid = ?", cid).First(f).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return f, nil
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
	if err := db.Db().Table("files").Where("encryption_status = 'public'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// count total encrypted files in database
func CountEncryptedFiles() (encryptedfiles int64, err error) {
	if err := db.Db().Table("files").Where("encryption_status = 'encrypted'").Count(&encryptedfiles).Error; err != nil {
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

func FindShareStateByFileUID(file_uid string) (file_share_state entity.FileShareState, err error) {
	if err := db.Db().Preload("PublicFile").Where("file_uid = ?", file_uid).First(&file_share_state).Error; err != nil {

		return file_share_state, err
	}

	return file_share_state, nil
}

func CreateShareState(file *entity.File) (file_share_state entity.FileShareState, err error) {
	file_share_state = entity.FileShareState{
		FileUID: file.UID,
	}

	if err := db.Db().Create(&file_share_state).Error; err != nil {
		return file_share_state, err
	}

	return file_share_state, nil
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
		Where("files.root = '/' AND files_users.permission <> 'deleted' AND files_users.user_id = ? AND files.deleted_at IS NULL", user_id).
		Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

func FindUsersByFileCID(cid string) ([]uint, error) {
	var fileUsers []entity.FileUser
	var usersWF []uint

	// Join File and FileUser tables and find records by CID
	err := db.Db().
		Table("files_users").
		Select("files_users.user_id").
		Joins("JOIN files ON files.id = files_users.file_id").
		Where("files.c_id = ? AND files.deleted_at IS NULL", cid).
		Find(&fileUsers).Error

	if err != nil {
		return nil, err
	}

	// Extract user IDs from the result
	for _, fu := range fileUsers {
		usersWF = append(usersWF, fu.UserID)
	}

	return usersWF, nil
}

func FindFilesByUserAndFileCID(userID uint, cid string) ([]entity.File, error) {
	var files []entity.File

	// Join File and FileUser tables and find records by CID
	err := db.Db().Unscoped().
		Table("files_users").
		Select("files.*").
		Joins("JOIN files ON files.id = files_users.file_id").
		Where("files_users.user_id = ? AND files.c_id = ? AND files.deleted_at IS NULL", userID, cid).
		Find(&files).Error

	if err != nil {
		return nil, err
	}

	return files, nil
}

// DeleteFileByUID deletes a file by its UID.
func DeleteFileByUID(file_uid string) error {
	if file_uid == "" {
		return fmt.Errorf("file uid required")
	}

	// Get file_shared_state
	file_share_state, err := FindShareStateByFileUID(file_uid)

	if err == nil {
		// Delete file_shared_state
		if err := file_share_state.Delete(); err != nil {
			return err
		}
	}

	return db.Db().Where("uid = ?", file_uid).Delete(&entity.File{}).Error
}

// query for count all txt files
func CountTxtFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("encryption_status = 'public' AND mime = 'text/plain'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query for count all public png files
func CountPngFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("encryption_status = 'public' AND mime = 'image/png'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query for count all jpg files
func CountJpgFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("encryption_status = 'public' AND (mime = 'image/jpg' OR mime = 'image/jpeg')").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query for count all pdf files
func CountPdfFiles() (publicfiles int64, err error) {
	if err := db.Db().Table("files").Where("encryption_status = 'public' AND mime = 'application/pdf'").Count(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query total sum storaged_used of individual user, need id in table users for make inner join with table user_details
func CountTotalUsedStorageUser(user_uid string) (totalusedstorage int64, err error) {
	if err := db.Db().Table("user_details").Select("SUM(storage_used)").Joins("INNER JOIN users ON users.id = user_details.user_id").Where("users.uid = ? ", user_uid).Scan(&totalusedstorage).Error; err != nil {
		return totalusedstorage, err
	}

	return totalusedstorage, nil
}

// query total sum user encrypted files (encyption_status), need id in table users for make inner join with table files_users and files
func CountTotalEncryptedFilesUser(user_uid string) (encryptedfiles int64, err error) {
	if err := db.Db().Table("files").Select("COUNT(*)").Joins("INNER JOIN files_users ON files_users.file_id = files.id").Joins("INNER JOIN users ON users.id = files_users.user_id").Where("users.uid = ? AND files.encryption_status = 'encrypted'  AND files.deleted_at IS NULL", user_uid).Scan(&encryptedfiles).Error; err != nil {
		return encryptedfiles, err
	}

	return encryptedfiles, nil
}

// query total sum user publicfiles files (encyption_status), need id in table users for make inner join with table files_users and files
func CountTotalPublicFilesUser(user_uid string) (publicfiles int64, err error) {
	if err := db.Db().Table("files").Select("COUNT(*)").Joins("INNER JOIN files_users ON files_users.file_id = files.id").Joins("INNER JOIN users ON users.id = files_users.user_id").Where("users.uid = ? AND files.encryption_status = 'public' AND files.deleted_at IS NULL", user_uid).Scan(&publicfiles).Error; err != nil {
		return publicfiles, err
	}

	return publicfiles, nil
}

// query total sum user publicfiles and encryptedfiles files (encyption_status), need id in table users for make inner join with table files_users and files
func CountTotalFilesUser(user_uid string) (upfile int64, err error) {
	if err := db.Db().Table("files").Select("COUNT(*)").Joins("INNER JOIN files_users ON files_users.file_id = files.id").Joins("INNER JOIN users ON users.id = files_users.user_id").Where("users.uid = ?  AND files.deleted_at IS NULL", user_uid).Scan(&upfile).Error; err != nil {
		return upfile, err
	}

	return upfile, nil
}

// Query storage used by user up to a specific date
func CountStorageUsed(daystring string, user_uid string) (dailystorage int64, err error) {
	query := db.Db().
		Table("files").
		Select("GREATEST(COALESCE(SUM(CASE WHEN files_users.permission != 'shared' AND (files.deleted_at IS NULL OR DATE(files.deleted_at) > DATE(?)) THEN files.size WHEN files_users.permission != 'shared' AND (files.deleted_at IS NOT NULL AND DATE(files.deleted_at) < DATE(?)) THEN -files.size ELSE 0 END), 0), 0)", daystring, daystring).
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Where("users.uid = ? AND DATE(files.created_at) <= DATE(?)", user_uid, daystring)

	// Execute and scan the result
	if err := query.Scan(&dailystorage).Error; err != nil {
		return 0, err
	}

	return dailystorage, nil
}

// Query storage used up to a specific date
func CountPublicStorageUsed(daystring string) (dailystorage int64, err error) {
	query := db.Db().
		Table("files").
		Select("GREATEST(COALESCE(SUM(CASE WHEN files_users.permission != 'shared' AND (files.deleted_at IS NULL OR DATE(files.deleted_at) > DATE(?)) THEN files.size WHEN files_users.permission != 'shared' AND (files.deleted_at IS NOT NULL AND DATE(files.deleted_at) < DATE(?)) THEN -files.size ELSE 0 END), 0), 0)", daystring, daystring).
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Where("DATE(files.created_at) <= DATE(?)", daystring)

	// Execute and scan the result
	if err := query.Scan(&dailystorage).Error; err != nil {
		return 0, err
	}

	return dailystorage, nil
}

// Query total files used by user up to a specific date
func CountFilesUsed(daystring string, user_uid string) (dailyfiles int64, err error) {
	query := db.Db().
		Table("files").
		Select("GREATEST(COALESCE(COUNT(CASE WHEN files_users.permission != 'shared' AND (files.deleted_at IS NULL OR DATE(files.deleted_at) > DATE(?)) THEN 1 END), 0), 0)", daystring).
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Where("users.uid = ? AND DATE(files.created_at) <= DATE(?)", user_uid, daystring)

	// Execute and scan the result
	if err := query.Scan(&dailyfiles).Error; err != nil {
		return 0, err
	}

	return dailyfiles, nil
}

// Query public files used by user up to a specific date
func CountFilesUsedByStatus(daystring string, user_uid string, status string) (dailypublicfiles int64, err error) {
	query := db.Db().
		Table("files").
		Select("GREATEST(COALESCE(COUNT(CASE WHEN files_users.permission != 'shared' AND (files.deleted_at IS NULL OR DATE(files.deleted_at) > DATE(?)) THEN 1 END), 0), 0)", daystring).
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Where("users.uid = ? AND DATE(files.created_at) <= DATE(?) AND files.encryption_status = ?", user_uid, daystring, status)

	// Execute and scan the result
	if err := query.Scan(&dailypublicfiles).Error; err != nil {
		return 0, err
	}

	return dailypublicfiles, nil
}

// Query storage used by user up to a specific date & hour
func CountStorageUsedH(daystring string, user_uid string) (dailystorage int64, err error) {
	query := db.Db().
		Table("files").
		Select("GREATEST(COALESCE(SUM(CASE WHEN files_users.permission != 'shared' AND (files.deleted_at IS NULL OR (files.deleted_at) > (?)) THEN files.size WHEN files_users.permission != 'shared' AND (files.deleted_at IS NOT NULL AND DATE(files.deleted_at) < DATE(?)) THEN -files.size ELSE 0 END), 0), 0)", daystring, daystring).
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Where("users.uid = ? AND (files.created_at) <= (?)", user_uid, daystring)

	// Execute and scan the result
	if err := query.Scan(&dailystorage).Error; err != nil {
		return 0, err
	}

	return dailystorage, nil
}

// Query total files used by user up to a specific date & hour
func CountFilesUsedH(daystring string, user_uid string) (dailyfiles int64, err error) {
	query := db.Db().
		Table("files").
		Select("GREATEST(COALESCE(COUNT(CASE WHEN files_users.permission != 'shared' AND (files.deleted_at IS NULL OR (files.deleted_at) > (?)) THEN 1 END), 0), 0)", daystring).
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Where("users.uid = ? AND (files.created_at) <= (?)", user_uid, daystring)

	// Execute and scan the result
	if err := query.Scan(&dailyfiles).Error; err != nil {
		return 0, err
	}

	return dailyfiles, nil
}

// Query public files used by user up to a specific date & hour
func CountFilesUsedByStatusH(daystring string, user_uid string, status string) (dailypublicfiles int64, err error) {
	query := db.Db().
		Table("files").
		Select("GREATEST(COALESCE(COUNT(CASE WHEN files_users.permission != 'shared' AND (files.deleted_at IS NULL OR (files.deleted_at) > (?)) THEN 1 END), 0), 0)", daystring).
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Where("users.uid = ? AND (files.created_at) <= (?) AND files.encryption_status = ?", user_uid, daystring, status)

	// Execute and scan the result
	if err := query.Scan(&dailypublicfiles).Error; err != nil {
		return 0, err
	}

	return dailypublicfiles, nil
}

func GetStartAndEndFileDatesPublic() (time.Time, time.Time, error) {
	var minDate, maxDate time.Time

	// Query for the earliest creation date
	err := db.Db().
		Table("files").
		Select("MIN(files.created_at)").
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Scan(&minDate).Error
	if err != nil {
		return minDate, maxDate, err
	}

	// Query for the latest creation date
	err = db.Db().
		Table("files").
		Select("MAX(files.created_at)").
		Joins("INNER JOIN files_users ON files_users.file_id = files.id").
		Joins("INNER JOIN users ON users.id = files_users.user_id").
		Scan(&maxDate).Error
	if err != nil {
		return minDate, maxDate, err
	}

	if minDate.IsZero() || maxDate.IsZero() { // Handle the case where there are no public files
		// This could be returning an error or default dates
		return time.Time{}, time.Time{}, errors.New("no public files found")
	}

	return minDate, maxDate, nil
}

func CountTotalFiles(encryptionType string, daystring string) (totalFiles int64, err error) {
	query := db.Db().
		Table("files").
		Select("COALESCE(COUNT(*), 0)").
		Where("files.encryption_status = ? AND DATE(files.created_at) <= DATE(?)", encryptionType, daystring)

	if err := query.Scan(&totalFiles).Error; err != nil {
		return totalFiles, err
	}

	return totalFiles, nil
}

func QueryShareGroupByHash(shareGroupHash string) ([]string, error) {
	var publicFileShareGroups []entity.PublicFileShareGroup

	// Query the database to fetch records of PublicFileShareGroup associated with the shareGroupHash
	if err := db.Db().Where("share_group_hash = ?", shareGroupHash).Find(&publicFileShareGroups).Error; err != nil {
		return nil, err
	}

	// Extract share hashes from PublicFileShareGroup records
	var shareHashes []string
	for _, publicFileShareGroup := range publicFileShareGroups {
		shareHashes = append(shareHashes, publicFileShareGroup.ShareHash)
	}

	return shareHashes, nil
}

// DeletePublicFileShareGroupByShareHash deletes the record in publicFileShareGroups
// based on the sharing hash.
func DeletePublicFileShareGroupByShareHash(shareHash string) error {
	var publicFileShareGroup entity.PublicFileShareGroup

	// Find the record based on the sharing hash
	result := db.Db().Where("share_hash = ?", shareHash).First(&publicFileShareGroup)

	// Check for errors during the query
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Record not found, return without error
			return nil
		}
		// Other error occurred
		return err
	}

	// Delete the record
	if err := db.Db().Delete(&publicFileShareGroup).Error; err != nil {
		return err
	}

	return nil
}

func DeleteFileShareStatesByFileUID(fileUID string) error {
	return db.Db().Where("file_uid = ?", fileUID).Delete(&entity.FileShareState{}).Error
}

func DeleteEmptyShareGroup(shareGroupHash string) error {
	var count int64
	if err := db.Db().Model(&entity.PublicFileShareGroup{}).Where("share_group_hash = ?", shareGroupHash).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		return db.Db().Unscoped().Where("hash = ?", shareGroupHash).Delete(&entity.ShareGroup{}).Error
	}

	return nil
}

func GetApiFiles(user_id uint) (files entity.Files, err error) {
	var apiFiles []entity.ApiKeyFile
	if err := db.Db().Where("user_id = ?", user_id).Find(&apiFiles).Error; err != nil {
		return nil, err
	}

	var fileIDs []uint
	for _, apiFile := range apiFiles {
		fileIDs = append(fileIDs, apiFile.FileID)
	}

	if err := db.Db().Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		return nil, err
	}

	return files, nil
}
