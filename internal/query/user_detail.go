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

func FindReferrerFromAddress(address string) string {
	var wallet, referrerWallet entity.Wallet

	//get user based on address
	if err := db.Db().Where("address = ?", address).First(&wallet).Error; err != nil {
		return ""
	}

	//get referred_by from user_detail based on user_id
	referrerID := FindReferrerIdFromReferredId(wallet.UserID)
	if referrerID == 0 {
		return ""
	}

	//get wallet based on referrer_id
	if err := db.Db().Where("user_id = ?", referrerID).First(&referrerWallet).Error; err != nil {
		return ""
	}

	return referrerWallet.Address

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
