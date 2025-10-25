// SPDX-License-Identifier: AGPL-3.0-or-later
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptToken encrypts a plaintext token using AES-256-GCM
// The key must be 32 bytes for AES-256
// Returns: nonce + ciphertext + auth tag (combined)
func EncryptToken(plaintext string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes for AES-256, got %d bytes", len(key))
	}

	if plaintext == "" {
		return nil, fmt.Errorf("cannot encrypt empty plaintext")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode (provides both confidentiality and authenticity)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce (must be unique for each encryption with the same key)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate the plaintext
	// Seal appends the ciphertext and tag to the nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return ciphertext, nil
}

// DecryptToken decrypts a ciphertext using AES-256-GCM
// The key must be 32 bytes for AES-256
// Expects input format: nonce + ciphertext + auth tag (as created by EncryptToken)
func DecryptToken(ciphertext []byte, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("decryption key must be 32 bytes for AES-256, got %d bytes", len(key))
	}

	if len(ciphertext) == 0 {
		return "", fmt.Errorf("cannot decrypt empty ciphertext")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length (nonce + at least 1 byte of data + tag)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short: expected at least %d bytes, got %d", nonceSize, len(ciphertext))
	}

	// Extract nonce and ciphertext+tag
	nonce := ciphertext[:nonceSize]
	encryptedData := ciphertext[nonceSize:]

	// Decrypt and verify authenticity
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
