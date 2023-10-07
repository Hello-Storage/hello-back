package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/pkg/rnd"
)

func FindFolder(find entity.Folder) *entity.Folder {
	m := &entity.Folder{}

	stmt := db.Db()

	if find.ID != 0 && find.Title != "" {
		stmt = stmt.Where("id = ? OR title = ?", find.ID, find.Title)
	} else if find.ID != 0 {
		stmt = stmt.Where("id = ?", find.ID)
	} else if rnd.IsUID(find.UID, entity.FolderUID) {
		stmt = stmt.Where("uid = ?", find.UID)
	} else if find.Title != "" {
		stmt = stmt.Where("title = ?", find.Title)
	} else {
		return nil
	}

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m

}

// FoldersByRoot returns folders in a given directory.
func FoldersByRoot(root string) (folders entity.Folders, err error) {
	if err := db.Db().Where("root = ?", root).Find(&folders).Error; err != nil {
		return folders, err
	}

	return folders, nil
}

func FindRootFoldersByUser(user_id uint) (folders entity.Folders, err error) {
	if err := db.Db().
		Table("folders").
		Joins("LEFT JOIN folders_users on folders_users.folder_id = folders.id").
		Where("folders.root = '/' AND folders_users.permission = 'owner' AND folders_users.user_id = ?", user_id).
		Find(&folders).Error; err != nil {
		return folders, err
	}

	return folders, nil
}

func FindFolderByTitleAndRoot(title, root string) *entity.Folder {
	m := &entity.Folder{}

	stmt := db.Db()
	stmt = stmt.Where("title = ? AND root = ?", title, root)

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m
}

func FindFolderPathByRoot(root string) entity.Folders {
	if root == "/" {
		return entity.Folders{}
	}

	m := FindFolder(entity.Folder{UID: root})

	return append(FindFolderPathByRoot(m.Root), *m)
}

// FindFolderByUID finds a folder by UID.
func FindFolderByUID(uid string) (*entity.Folder, error) {
	m := &entity.Folder{}

	if err := db.Db().Where("uid = ?", uid).First(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}

// DeleteFolderByUID deletes a folder by UID.
func DeleteFolderByUID(uid string) error {
	if err := db.Db().Where("uid = ?", uid).Delete(&entity.Folder{}).Error; err != nil {
		return err
	}
	return nil
}

// FindFolderUsers finds a user with a certain permission level for a folder
func FindFolderUser(folderID, userID uint) (*entity.FolderUser, error) {
	fu := &entity.FolderUser{}
	if err := db.Db().Where("folder_id = ? AND user_id = ?", folderID, userID).First(fu).Error; err != nil {
		return nil, err
	}
	return fu, nil
}

// GetChildFoldersByUID returns child folders of a given folder.
func GetChildFoldersByUID(uid string) (folders entity.Folders, err error) {
	if err := db.Db().Where("root = ?", uid).Find(&folders).Error; err != nil {
		return folders, err
	}
	return folders, nil
}

func GetFolderFilesByUID(folderUID string) (files entity.Files, err error) {
	if err := db.Db().Where("root = ?", folderUID).Find(&files).Error; err != nil {
		return files, err
	}
	return files, nil
}

// query for count all public folders
func CountPublicFolders() (publicfolders int64, err error) {
	if err := db.Db().Table("folders").Where("encryption_status = 'public'").Count(&publicfolders).Error; err != nil {
		return publicfolders, err
	}

	return publicfolders, nil
}

// query total sum user public folders folders , need id in table users for make inner join with table folders_users and folders
func CountTotalPublicFoldersUser(user_uid string) (publicfolders int64, err error) {
	if err := db.Db().Table("folders").Select("COUNT(*)").Joins("INNER JOIN folders_users ON folders_users.folder_id = folders.id").Joins("INNER JOIN users ON users.id = folders_users.user_id").Where("users.uid = ? AND folders.deleted_at IS NULL ", user_uid).Scan(&publicfolders).Error; err != nil {
		return publicfolders, err
	}

	return publicfolders, nil
}
