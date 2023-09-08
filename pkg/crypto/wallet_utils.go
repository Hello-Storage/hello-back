package crypto

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)


func IsValidEthereumAddress(address string) bool {
	// Check if address is valid
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		return false
	}
	return common.IsHexAddress(address)
}


func IsValidEthereumPrivateKey(privateKey string) bool {
	// Remove the 0x prefix
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}
	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return false
	}
	// Additional check: ensure the private key corresponds to a non-null public key
	return !crypto.PubkeyToAddress(key.PublicKey).Big().IsInt64()
}