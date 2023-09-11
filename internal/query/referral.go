package query

import (
	"fmt"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
)

func CheckReferralCode(referral_code string) (uint, error) {
	//get the user id from the referral code
	m := &entity.Wallet{}
	stmt := db.Db()

	stmt = stmt.Where("address = ?", referral_code)

	// Find matching record.
	fmt.Println("referral_code: ", referral_code)
	if err := stmt.First(m).Error; err != nil {
		return 0, err
	}

	return m.UserID, nil
}

func FindReferredUsers(referral_code string) ([]entity.User, error) {
	//get the user id from the referral code
	m := &entity.Wallet{}
	stmt := db.Db()

	stmt = stmt.Where("address = ?", referral_code)

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil, err
	}

	//get the user details from the user id
	n := &entity.UserDetail{}
	stmt = db.Db()

	stmt = stmt.Where("user_id = ?", m.UserID)

	// Find matching record.
	if err := stmt.First(n).Error; err != nil {
		return nil, err
	}

	//get the referred users from the user details
	o := &[]entity.User{}
	stmt = db.Db()

	stmt = stmt.Where("referred_by = ?", n.ID)

	// Find matching record.
	if err := stmt.Find(o).Error; err != nil {
		return nil, err
	}

	return *o, nil
}