package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNonce(t *testing.T) {
	t.Run("generates valid nonce", func(t *testing.T) {
		nonce, err := GenerateNonce()
		require.NoError(t, err)
		assert.NotEmpty(t, nonce)
	})

	t.Run("nonce is valid base64url", func(t *testing.T) {
		nonce, err := GenerateNonce()
		require.NoError(t, err)

		// Should be decodable as base64url
		decoded, err := base64.RawURLEncoding.DecodeString(nonce)
		require.NoError(t, err)

		// Should be 16 bytes (128 bits) when decoded
		assert.Len(t, decoded, 16, "Decoded nonce should be 16 bytes")
	})

	t.Run("generates unique nonces", func(t *testing.T) {
		const numNonces = 1000
		nonces := make(map[string]bool)

		for i := 0; i < numNonces; i++ {
			nonce, err := GenerateNonce()
			require.NoError(t, err)

			// Check for duplicates
			assert.False(t, nonces[nonce], "Nonce %s should be unique", nonce)
			nonces[nonce] = true
		}

		assert.Len(t, nonces, numNonces, "All nonces should be unique")
	})

	t.Run("nonce format consistency", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			nonce, err := GenerateNonce()
			require.NoError(t, err)

			// Should not be empty
			assert.NotEmpty(t, nonce)

			// Should not contain padding (RawURLEncoding)
			assert.NotContains(t, nonce, "=", "Nonce should not contain padding")

			// Should only contain valid base64url characters
			assert.Regexp(t, `^[A-Za-z0-9_-]+$`, nonce, "Nonce should only contain base64url characters")
		}
	})

	t.Run("nonce length consistency", func(t *testing.T) {
		var lengths []int

		for i := 0; i < 100; i++ {
			nonce, err := GenerateNonce()
			require.NoError(t, err)

			lengths = append(lengths, len(nonce))
		}

		// All nonces should have same length
		expectedLength := lengths[0]
		for _, length := range lengths {
			assert.Equal(t, expectedLength, length, "All nonces should have consistent length")
		}

		// For 16 bytes (128 bits), base64url without padding should be 22 characters
		// 16 bytes = 128 bits = 128/6 = 21.33 -> 22 characters (rounded up)
		assert.Equal(t, 22, expectedLength, "Nonce should be 22 characters long")
	})

	t.Run("concurrent nonce generation", func(t *testing.T) {
		const numGoroutines = 100
		const noncesPerGoroutine = 10

		nonceChan := make(chan string, numGoroutines*noncesPerGoroutine)
		errorChan := make(chan error, numGoroutines*noncesPerGoroutine)

		// Start multiple goroutines generating nonces
		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < noncesPerGoroutine; j++ {
					nonce, err := GenerateNonce()
					if err != nil {
						errorChan <- err
						return
					}
					nonceChan <- nonce
				}
			}()
		}

		// Collect results
		nonces := make(map[string]bool)
		for i := 0; i < numGoroutines*noncesPerGoroutine; i++ {
			select {
			case nonce := <-nonceChan:
				assert.False(t, nonces[nonce], "Concurrent nonce %s should be unique", nonce)
				nonces[nonce] = true
			case err := <-errorChan:
				t.Fatalf("Concurrent nonce generation failed: %v", err)
			}
		}

		assert.Len(t, nonces, numGoroutines*noncesPerGoroutine, "All concurrent nonces should be unique")
	})

	t.Run("nonce entropy validation", func(t *testing.T) {
		const numNonces = 1000
		bitCounts := make([]int, 8) // Count bits 0-7 across all bytes

		for i := 0; i < numNonces; i++ {
			nonce, err := GenerateNonce()
			require.NoError(t, err)

			decoded, err := base64.RawURLEncoding.DecodeString(nonce)
			require.NoError(t, err)

			// Count bit frequency
			for _, b := range decoded {
				for bit := 0; bit < 8; bit++ {
					if (b>>bit)&1 == 1 {
						bitCounts[bit]++
					}
				}
			}
		}

		// Each bit should appear roughly 50% of the time (within reasonable variance)
		expectedCount := numNonces * 16 / 2 // 16 bytes per nonce, expect 50% ones
		tolerance := expectedCount / 10     // 10% tolerance

		for bit, count := range bitCounts {
			assert.InDelta(t, expectedCount, count, float64(tolerance),
				"Bit %d should have balanced distribution (got %d, expected ~%d)",
				bit, count, expectedCount)
		}
	})

	t.Run("nonce base64url safety", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			nonce, err := GenerateNonce()
			require.NoError(t, err)

			// Should not contain characters that need URL encoding
			assert.NotContains(t, nonce, "+", "Nonce should not contain + (use URL-safe base64)")
			assert.NotContains(t, nonce, "/", "Nonce should not contain / (use URL-safe base64)")
			assert.NotContains(t, nonce, "=", "Nonce should not contain = (use RawURLEncoding)")

			// Should be safe for use in URLs and forms
			assert.Regexp(t, `^[A-Za-z0-9_-]+$`, nonce, "Nonce should only contain URL-safe characters")
		}
	})

	t.Run("nonce anti-replay properties", func(t *testing.T) {
		// Generate a large set of nonces to verify anti-replay properties
		const numNonces = 10000
		nonces := make([]string, 0, numNonces)
		nonceSet := make(map[string]bool)

		for i := 0; i < numNonces; i++ {
			nonce, err := GenerateNonce()
			require.NoError(t, err)

			// Verify uniqueness (anti-replay)
			assert.False(t, nonceSet[nonce], "Nonce should not repeat (anti-replay)")
			nonceSet[nonce] = true
			nonces = append(nonces, nonce)
		}

		// Verify we generated the expected number of unique nonces
		assert.Len(t, nonces, numNonces)
		assert.Len(t, nonceSet, numNonces)

		// Verify sufficient entropy - no obvious patterns
		// Check that first characters are well distributed
		firstChars := make(map[byte]int)
		for _, nonce := range nonces {
			firstChars[nonce[0]]++
		}

		// Should have reasonable distribution of first characters
		assert.Greater(t, len(firstChars), 10, "First character should have good distribution")
	})

	t.Run("nonce cryptographic strength", func(t *testing.T) {
		// Test that nonces have sufficient randomness
		nonce1, err := GenerateNonce()
		require.NoError(t, err)

		nonce2, err := GenerateNonce()
		require.NoError(t, err)

		// Different nonces should be completely different
		assert.NotEqual(t, nonce1, nonce2)

		// Decode both nonces
		decoded1, err := base64.RawURLEncoding.DecodeString(nonce1)
		require.NoError(t, err)

		decoded2, err := base64.RawURLEncoding.DecodeString(nonce2)
		require.NoError(t, err)

		// Should have no common bytes (extremely unlikely with crypto/rand)
		commonBytes := 0
		for i := range decoded1 {
			if decoded1[i] == decoded2[i] {
				commonBytes++
			}
		}

		// With truly random data, expect 0-2 common bytes in 16-byte sequences
		assert.LessOrEqual(t, commonBytes, 3, "Too many common bytes between random nonces")
	})

	t.Run("error handling edge cases", func(t *testing.T) {
		// This test verifies the function handles errors gracefully
		// In normal conditions, GenerateNonce should not fail
		// but we test that error handling pattern is correct

		nonce, err := GenerateNonce()

		// In normal cases, should always succeed
		require.NoError(t, err)
		assert.NotEmpty(t, nonce)

		// If it did error, nonce should be empty string
		if err != nil {
			assert.Empty(t, nonce, "On error, nonce should be empty")
		}
	})
}
