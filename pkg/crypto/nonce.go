package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateNonce generates a cryptographically secure random nonce
func GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(nonceBytes), nil
}
