package token

import (
	"time"
)

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(user_id uint, user_uid, user_name string, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)

	// CreateApiKey creates a new apikey for a specific username
	CreateApiKey(user_id uint, user_uid, user_name string) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyApiKey(apikey string) (*Payload, error)
}
