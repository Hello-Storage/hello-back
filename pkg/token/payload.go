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

	payload := &Payload{
		TokenID:   tokenID,
		UserID:    user_id,
		UserUID:   user_uid,
		UserName:  user_name,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
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
