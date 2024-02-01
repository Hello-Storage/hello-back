package query

import (
	"errors"
	"fmt"
	"time"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/pkg/rnd"
	"gorm.io/gorm"
)

// RegisteredUsers finds all registered users.
func RegisteredUsers() (result entity.Users) {
	if err := db.Db().Where("id > 0").Find(&result).Error; err != nil {
		log.Errorf("users: %s", err)
	}

	return result
}

func FindUserByUID(id uint) (*entity.User, error) {

	var user entity.User
	err := db.Db().Where("id = ?", id).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found for ID: %d", id)
		}
		return nil, fmt.Errorf("failed to find file: %v", err)
	}

	return &user, nil
}

func FindUser(find entity.User) *entity.User {
	m := &entity.User{}

	stmt := db.Db().Preload("Wallet")

	//INFO[2023-10-17T13:01:30Z] user id: 4
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
		log.Error(err)
		return nil
	}

	//print m:

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

// Count total users in database
func CountUsers() (totalusers int64, err error) {
	if err := db.Db().Table("users").Count(&totalusers).Error; err != nil {
		return totalusers, err
	}

	return totalusers, nil
}

func FindUserByEmail(email string) *entity.User {
	u := &entity.User{}

	subquery := db.Db().Table("emails").Select("user_id").Where("email = ?", email)
	if err := db.Db().Model(u).Preload("Wallet").Preload("Email").Where("id IN (?)", subquery).First(u).Error; err == nil {
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
	if result := db.Db().Model(user).Preload("Wallet").Preload("Detail").Where("id = ?", userID).First(&user); result.Error != nil {
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

func CountTotalUsers(daystring string) (totalUsers int, err error) {
	query := db.Db().
		Table("users").
		Select("COALESCE(COUNT(*), 0)").
		Where("DATE(users.created_at) <= DATE(?)", daystring)
	if err := query.Scan(&totalUsers).Error; err != nil {
		return 0, err
	}

	return totalUsers, nil
}

func GetStartAndEndUserDatesPublic() (time.Time, time.Time, error) {
	var minDate, maxDate time.Time

	// Query for the earliest creation date
	err := db.Db().
		Table("users").
		Select("MIN(users.created_at)").
		Scan(&minDate).Error
	if err != nil {
		return minDate, maxDate, err
	}

	// Query for the latest creation date
	err = db.Db().
		Table("users").
		Select("MAX(users.created_at)").
		Scan(&maxDate).Error
	if err != nil {
		return minDate, maxDate, err
	}

	if minDate.IsZero() || maxDate.IsZero() { // Handle the case where there are no public files
		// This could be returning an error or default dates
		return time.Time{}, time.Time{}, errors.New("no users found")
	}

	return minDate, maxDate, nil
}

// Query get user files by user id
func GetFilesUserFromUser(user_id uint) ([]entity.FileUser, error) {
	var filesUsers []entity.FileUser

	if err := db.Db().Where("user_id = ? AND permission != ?", user_id, entity.DeletedPermission).Find(&filesUsers).Error; err != nil {
		return nil, err
	}

	return filesUsers, nil
}

// Query get user folders by user id
func GetFoldersUserFromUser(user_id uint) ([]entity.FolderUser, error) {
	var foldersUsers []entity.FolderUser

	if err := db.Db().Where("user_id = ?", user_id).Find(&foldersUsers).Error; err != nil {
		return nil, err
	}

	return foldersUsers, nil
}

// 2023-11-27 13:18:35 backend   | time="2023-11-27T18:18:35Z" level=info msg="Calculated initial weekly user stats"
// 2023-11-27 13:18:35 backend   | time="2023-11-27T18:18:35Z" level=info msg="Calculated initial weekly storage stats"
// 2023-11-27 13:18:35 backend   | time="2023-11-27T18:18:35Z" level=error msg="runtime error: integer divide by zero"
// backend exited with code 0
