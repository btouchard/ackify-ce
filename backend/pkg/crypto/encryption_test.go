// SPDX-License-Identifier: AGPL-3.0-or-later
package crypto

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32) // AES-256 requires 32 bytes
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("encrypt and decrypt successfully", func(t *testing.T) {
		plaintext := "my-secret-refresh-token-12345"

		ciphertext, err := EncryptToken(plaintext, key)
		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext)

		// Ciphertext should be different from plaintext
		assert.NotEqual(t, plaintext, string(ciphertext))

		// Decrypt
		decrypted, err := DecryptToken(ciphertext, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("encrypt produces different ciphertext each time", func(t *testing.T) {
		plaintext := "same-plaintext"

		ciphertext1, err := EncryptToken(plaintext, key)
		require.NoError(t, err)

		ciphertext2, err := EncryptToken(plaintext, key)
		require.NoError(t, err)

		// Different nonces should produce different ciphertexts
		assert.NotEqual(t, ciphertext1, ciphertext2)

		// Both should decrypt to the same plaintext
		decrypted1, err := DecryptToken(ciphertext1, key)
		require.NoError(t, err)
		decrypted2, err := DecryptToken(ciphertext2, key)
		require.NoError(t, err)

		assert.Equal(t, plaintext, decrypted1)
		assert.Equal(t, plaintext, decrypted2)
	})

	t.Run("decrypt with wrong key fails", func(t *testing.T) {
		plaintext := "secret-token"
		wrongKey := make([]byte, 32)
		_, err := rand.Read(wrongKey)
		require.NoError(t, err)

		ciphertext, err := EncryptToken(plaintext, key)
		require.NoError(t, err)

		// Try to decrypt with wrong key
		_, err = DecryptToken(ciphertext, wrongKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt")
	})

	t.Run("tampered ciphertext fails authentication", func(t *testing.T) {
		plaintext := "secret-token"

		ciphertext, err := EncryptToken(plaintext, key)
		require.NoError(t, err)

		// Tamper with the ciphertext
		tampered := make([]byte, len(ciphertext))
		copy(tampered, ciphertext)
		tampered[len(tampered)-1] ^= 0xFF // Flip bits in the last byte

		// Decryption should fail due to authentication tag mismatch
		_, err = DecryptToken(tampered, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt")
	})

	t.Run("handles special characters", func(t *testing.T) {
		plaintext := "token-with-special-chars: !@#$%^&*()_+-={}[]|\\:\";<>?,./~`"

		ciphertext, err := EncryptToken(plaintext, key)
		require.NoError(t, err)

		decrypted, err := DecryptToken(ciphertext, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("handles long tokens", func(t *testing.T) {
		plaintext := strings.Repeat("a", 10000) // 10KB token

		ciphertext, err := EncryptToken(plaintext, key)
		require.NoError(t, err)

		decrypted, err := DecryptToken(ciphertext, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestEncryptToken_InvalidInputs(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("empty plaintext", func(t *testing.T) {
		_, err := EncryptToken("", key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot encrypt empty plaintext")
	})

	t.Run("invalid key length - too short", func(t *testing.T) {
		shortKey := make([]byte, 16) // Only 16 bytes (AES-128)
		_, err := EncryptToken("plaintext", shortKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encryption key must be 32 bytes")
	})

	t.Run("invalid key length - too long", func(t *testing.T) {
		longKey := make([]byte, 64) // 64 bytes (too long)
		_, err := EncryptToken("plaintext", longKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encryption key must be 32 bytes")
	})

	t.Run("nil key", func(t *testing.T) {
		_, err := EncryptToken("plaintext", nil)
		assert.Error(t, err)
	})
}

func TestDecryptToken_InvalidInputs(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("empty ciphertext", func(t *testing.T) {
		_, err := DecryptToken([]byte{}, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot decrypt empty ciphertext")
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		shortCiphertext := []byte{0x01, 0x02, 0x03} // Only 3 bytes
		_, err := DecryptToken(shortCiphertext, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ciphertext too short")
	})

	t.Run("invalid key length", func(t *testing.T) {
		ciphertext, err := EncryptToken("test", key)
		require.NoError(t, err)

		shortKey := make([]byte, 16)
		_, err = DecryptToken(ciphertext, shortKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decryption key must be 32 bytes")
	})

	t.Run("nil key", func(t *testing.T) {
		ciphertext, err := EncryptToken("test", key)
		require.NoError(t, err)

		_, err = DecryptToken(ciphertext, nil)
		assert.Error(t, err)
	})

	t.Run("corrupted data", func(t *testing.T) {
		corruptedData := make([]byte, 50)
		_, err := rand.Read(corruptedData)
		require.NoError(t, err)

		_, err = DecryptToken(corruptedData, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt")
	})
}

func TestEncryption_SecurityProperties(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("nonce uniqueness", func(t *testing.T) {
		plaintext := "test-token"
		nonces := make(map[string]bool)

		// Generate 1000 encryptions and check nonce uniqueness
		for i := 0; i < 1000; i++ {
			ciphertext, err := EncryptToken(plaintext, key)
			require.NoError(t, err)

			// Extract nonce (first 12 bytes for GCM)
			nonce := string(ciphertext[:12])
			assert.False(t, nonces[nonce], "duplicate nonce detected")
			nonces[nonce] = true
		}

		assert.Len(t, nonces, 1000)
	})
}
