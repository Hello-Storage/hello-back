package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/pkg/rnd"
)

// RegisteredUsers finds all registered users.
func RegisteredUsers() (result entity.Users) {
	if err := db.Db().Where("id > 0").Find(&result).Error; err != nil {
		log.Errorf("users: %s", err)
	}

	return result
}

func FindUser(find entity.User) *entity.User {
	m := &entity.User{}

	stmt := db.Db()

	if find.ID != 0 && find.Name != "" {
		stmt = stmt.Where("id = ? OR name = ?", find.ID, find.Name)
	} else if find.ID != 0 {
		stmt = stmt.Where("id = ?", find.ID)
	} else if rnd.IsUID(find.UID, entity.UserUID) {
		stmt = stmt.Where("uid = ?", find.UID)
	} else if find.Name != "" {
		stmt = stmt.Where("name = ?", find.Name)
	} else {
		return nil
	}

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m

}

func FindUserByName(name string) *entity.User {
	m := &entity.User{}

	stmt := db.Db()

	stmt = stmt.Where("name = ?", name).Preload("Email").Preload("Wallet").Preload("Github")

	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m
}

func FindUserByEmail(email string) *entity.User {
	u := &entity.User{}

	subquery := db.Db().Table("emails").Select("user_id").Where("email = ?", email)
	if err := db.Db().Model(u).Preload("Email").Where("id IN (?)", subquery).First(u).Error; err == nil {
		return u
	} else {
		return nil
	}
}

func FindUserByWalletAddress(walletAddress string) *entity.User {
	u := &entity.User{}

	subquery := db.Db().Table("wallets").Select("user_id").Where("address = ?", walletAddress)
	if err := db.Db().Model(u).Preload("Wallet").Where("id IN (?)", subquery).First(u).Error; err == nil {
		return u
	} else {
		return nil
	}
}

func FindUserWithWallet(userID uint) *entity.User {
	var user entity.User
	if result := db.Db().Preload("Wallet").First(&user, userID); result.Error != nil {
		log.Errorf("failed to find user: %s", result.Error)
		return nil
	}
	return &user
}

func FindUserByGithub(github_id uint) *entity.User {
	u := &entity.User{}

	subquery := db.Db().Table("githubs").Select("user_id").Where("github_id = ?", github_id)
	if err := db.Db().Model(u).Preload("Github").Where("id IN (?)", subquery).First(u).Error; err == nil {
		return u
	} else {
		return nil
	}
}
