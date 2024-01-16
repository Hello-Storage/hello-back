package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
)

func CheckFolderPermByUser(folder_uid string, user_id uint) bool {

	m := &entity.FolderUser{}

	subquery := db.Db().Table("folders").Select("id").Where("uid = ?", folder_uid)

	if err := db.Db().Model(m).Where("folder_id = (?) AND user_id = ? AND (permission = 'owner' OR permission = 'shared')", subquery, user_id).First(m).Error; err == nil {
		return true
	}

	return false

}
