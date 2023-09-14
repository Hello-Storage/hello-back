package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
)

func FindUserDetailByUserID(user_id uint) *entity.UserDetail {
	m := &entity.UserDetail{}
	stmt := db.Db()

	stmt = stmt.Where("user_id = ?", user_id)

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m
}

func FindUserDetailByUserUID(user_uid string) *entity.UserDetail {
	m := &entity.UserDetail{}
	stmt := db.Db()

	stmt = stmt.Joins("LEFT JOIN users on users.id = user_details.user_id")
	stmt = stmt.Where("users.uid = ?", user_uid)

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m
}
