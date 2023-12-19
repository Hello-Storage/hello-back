package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	TokenID   uuid.UUID `json:"token_id"`
	UserID    uint      `json:"id"`
	UserUID   string    `json:"uid"`
	UserName  string    `json:"name"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(user_id uint, user_uid, user_name string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	var expirationDate time.Time
	if duration == 0 {
		// If the duration is zero, set the expiration date to the year 9999,
		// we need to change this when we want to put expiration for the api
		expirationDate = time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	} else {
		expirationDate = time.Now().Add(duration)
	}

	payload := &Payload{
		TokenID:   tokenID,
		UserID:    user_id,
		UserUID:   user_uid,
		UserName:  user_name,
		IssuedAt:  time.Now(),
		ExpiredAt: expirationDate,
	}
	return payload, nil
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if payload.ExpiredAt.IsZero() {
		return nil
	}

	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
