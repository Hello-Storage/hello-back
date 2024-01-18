package form

import "time"

type CreateApiKeyRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}
type CreateApiKeyResponse struct {
	ApiKey          string    `json:"api_key"`
	ApiKeyExpiresAt time.Time `json:"api_key_expires_at"`
}
