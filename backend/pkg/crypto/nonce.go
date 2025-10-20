// SPDX-License-Identifier: AGPL-3.0-or-later
package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateNonce creates a 16-byte cryptographically secure random nonce for replay attack prevention
func GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(nonceBytes), nil
}
