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

func FindReferredUsers(referralCode string) ([]string, error) {
	//get the user id from the referral code
	var wallet entity.Wallet
	var userDetail entity.UserDetail

	if err := db.Db().Where("address = ?", referralCode).First(&wallet).Error; err != nil {
		return nil, err
	}

	//get the user details from the user id
	if err := db.Db().Where("user_id = ?", wallet.UserID).First(&userDetail).Error; err != nil {
		return nil, err
	}

	//get the referred users from the user details

	// Find matching record.
	var userDetails []entity.UserDetail
	if err := db.Db().Where("referred_by = ?", userDetail.ID).Find(&userDetails).Error; err != nil {
		return nil, err
	}

	//get the users from the users details' user_id
	var addresses []string
	for _, detail := range userDetails {
		var wallet entity.Wallet

		if err := db.Db().Where("user_id = ?", detail.UserID).First(&wallet).Error; err != nil {
			return nil, err
		}
		addresses = append(addresses, wallet.Address)

	}

	return addresses, nil
}

func FindReferrerIdFromReferredId(referred_id uint) uint {
	m := &entity.Referral{}
	stmt := db.Db()

	stmt = stmt.Where("referred_id = ?", referred_id)

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return 0
	}
	return m.ReferrerID
}

func UpdateReferralStorage(user_id uint) error {
	detail := &entity.UserDetail{}

	if err := db.Db().Preload("Referrals").Where("user_id = ?", user_id).First(&detail).Error; err != nil {
		return err
	}

	detail.ReferralStorage = uint(len(detail.Referrals) * 10 * 1024 * 1024 * 1024)

	if err := detail.Save(); err != nil {
		return err
	}

	return nil
}
