package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"github.com/Hello-Storage/hello-back/internal/config"
)

var (
	ErrDecryptionFailed = errors.New("decryption failed")
)

func Encrypt(plainText string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(os.Getenv("ENCRYPTION_KEY")))
	if err != nil {
		fmt.Printf("failed to create new cipher: %v", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Printf("failed to create new gcm: %v", err)
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return ciphertext, nil
}

func Decrypt(ciphertext []byte) (string, error) {
	block, err := aes.NewCipher([]byte(config.Env().EncryptionKey))
	if err != nil {
		fmt.Printf("failed to create new cipher: %v", err)
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Printf("failed to create new gcm: %v", err)
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		fmt.Printf("failed to get nonce size: %v", err)
		return "", ErrDecryptionFailed
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Printf("failed to open gcm: %v", err)
		return "", err
	}

	return string(plaintext), nil
}