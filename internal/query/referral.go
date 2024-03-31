package query

import (
	"errors"
	"fmt"

	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
)

// Create referal
func CreateReferral(referrer_id uint, userID uint, user_detailID uint) error {
	referral := entity.Referral{
		ReferrerID:   referrer_id,
		ReferredID:   userID,
		UserDetailID: user_detailID,
	}
	//check existance of the referal parts
	wallet := &entity.Wallet{}
	newUser := &entity.User{}
	newUserDetail := &entity.UserDetail{}
	err := db.Db().Where("address = ?", referrer_id).First(wallet)
	if err != nil {
		return errors.New("invalid referral code")
	}
	err = db.Db().Where("id = ?", userID).First(newUser)
	if err != nil {
		return errors.New("invalid referral code")
	}
	err = db.Db().Where("id = ?", user_detailID).First(newUserDetail)
	if err != nil {
		return errors.New("invalid referral code")
	}
	
	if err := db.Db().Create(&referral); err != nil {
		log.Errorf("failed to create referral: %v", err)
		return errors.New("failed to create referral")
	}

	if err := UpdateReferralStorage(referrer_id); err != nil {
		log.Errorf("failed to update referral storage: %v", err)
		return errors.New("failed to update referral storage")
	}

	return nil
}

func CheckReferralCode(referral_code string) (uint, error) {
	//get the user id from the referral code
	m := &entity.Wallet{}

	//must be a referral code
	if referral_code == "" || len(referral_code) < 5 {
		return 0, fmt.Errorf("invalid referral code: %s", referral_code)
	}

	// Find matching record.
	fmt.Println("referral_code: ", referral_code)
	if err := db.Db().Where("address = ?", referral_code).First(m).Error; err != nil {
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
	// 100 GB = 100 * 1024 * 1024 * 1024
	// if referral storage is 100 GB then return nil as storage limit reached already
	if detail.ReferralStorage == 100 * 1024 * 1024 * 1024  {
		fmt.Println("storage limit reached")
		return nil
	}

	detail.ReferralStorage = uint((len(detail.Referrals)+1) * 5 * 1024 * 1024 * 1024) // 5 GB per referral

	if err := detail.Save(); err != nil {
		return err
	}

	return nil
}