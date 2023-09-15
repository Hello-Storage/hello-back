package form

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type RegisterUserRequest struct {
	Name string `json:"name"`
}

type LoginUserRequest struct {
	Name          string `json:"name"`
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature"      binding:"required"`
	Referral      string `json:"referral"`
}

type LoginUserResponse struct {
	SessionID             uuid.UUID `json:"session_id"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	WalletAddress         string    `json:"wallet_address"`
	WalletPrivateKey      string    `json:"wallet_private_key"`
}
