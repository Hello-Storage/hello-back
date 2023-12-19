package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPasetoMaker creates a new PasetoMaker
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

// CreateToken creates a new token for a specific user and duration
func (maker *PasetoMaker) CreateToken(user_id uint, user_uid, user_name string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(user_id, user_uid, user_name, duration)
	if err != nil {
		return "", payload, err
	}

	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	return token, payload, err
}

// VerifyToken checks if the token is valid or not
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// CreateApiKey creates a new API key for a specific user
func (maker *PasetoMaker) CreateApiKey(user_id uint, user_uid, user_name string) (string, *Payload, error) {
	//duration of 0 for no expiration
	payload, err := NewPayload(user_id, user_uid, user_name, 0)
	if err != nil {
		return "", payload, err
	}

	apiKey, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	return apiKey, payload, err
}

// VerifyApiKey checks if the API key is valid or not
func (maker *PasetoMaker) VerifyApiKey(apiKey string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(apiKey, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
