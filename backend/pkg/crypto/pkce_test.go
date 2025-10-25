// SPDX-License-Identifier: AGPL-3.0-or-later
package crypto

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCodeVerifier(t *testing.T) {
	t.Run("generates valid verifier", func(t *testing.T) {
		verifier, err := GenerateCodeVerifier()
		require.NoError(t, err)
		assert.NotEmpty(t, verifier)

		// Check length (should be 43 characters for 32 bytes)
		assert.Len(t, verifier, 43)

		// Validate format
		assert.True(t, ValidateCodeVerifier(verifier))
	})

	t.Run("generates unique verifiers", func(t *testing.T) {
		verifiers := make(map[string]bool)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			verifier, err := GenerateCodeVerifier()
			require.NoError(t, err)

			// Check no collision
			assert.False(t, verifiers[verifier], "duplicate verifier generated")
			verifiers[verifier] = true
		}

		assert.Len(t, verifiers, iterations)
	})

	t.Run("contains only URL-safe characters", func(t *testing.T) {
		verifier, err := GenerateCodeVerifier()
		require.NoError(t, err)

		// Check it's valid base64 URL encoding
		_, err = base64.RawURLEncoding.DecodeString(verifier)
		assert.NoError(t, err)

		// Should not contain padding
		assert.False(t, strings.Contains(verifier, "="))
	})
}

func TestGenerateCodeChallenge(t *testing.T) {
	t.Run("generates correct SHA256 challenge", func(t *testing.T) {
		// RFC 7636 test vector
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		expectedChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

		challenge := GenerateCodeChallenge(verifier)
		assert.Equal(t, expectedChallenge, challenge)
	})

	t.Run("generates base64 URL-safe challenge", func(t *testing.T) {
		verifier, err := GenerateCodeVerifier()
		require.NoError(t, err)

		challenge := GenerateCodeChallenge(verifier)
		assert.NotEmpty(t, challenge)

		// Should be valid base64 URL encoding
		_, err = base64.RawURLEncoding.DecodeString(challenge)
		assert.NoError(t, err)

		// Should not contain padding
		assert.False(t, strings.Contains(challenge, "="))

		// Should be 43 characters (32 bytes SHA256 in base64)
		assert.Len(t, challenge, 43)
	})

	t.Run("different verifiers produce different challenges", func(t *testing.T) {
		verifier1, _ := GenerateCodeVerifier()
		verifier2, _ := GenerateCodeVerifier()

		challenge1 := GenerateCodeChallenge(verifier1)
		challenge2 := GenerateCodeChallenge(verifier2)

		assert.NotEqual(t, challenge1, challenge2)
	})

	t.Run("same verifier produces same challenge", func(t *testing.T) {
		verifier := "test_verifier_123456789012345678901234567"

		challenge1 := GenerateCodeChallenge(verifier)
		challenge2 := GenerateCodeChallenge(verifier)

		assert.Equal(t, challenge1, challenge2)
	})
}

func TestValidateCodeVerifier(t *testing.T) {
	tests := []struct {
		name     string
		verifier string
		valid    bool
	}{
		{
			name:     "valid 43 character verifier",
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			valid:    true,
		},
		{
			name:     "valid 128 character verifier",
			verifier: strings.Repeat("a", 128),
			valid:    true,
		},
		{
			name:     "valid with numeric and alphanumeric (50 chars)",
			verifier: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX",
			valid:    true,
		},
		{
			name:     "valid with special chars (44 chars)",
			verifier: "abc123-._~ABC123-._~abc123-._~abc123-._~abc",
			valid:    true,
		},
		{
			name:     "too short (42 chars)",
			verifier: strings.Repeat("a", 42),
			valid:    false,
		},
		{
			name:     "too long (129 chars)",
			verifier: strings.Repeat("a", 129),
			valid:    false,
		},
		{
			name:     "empty string",
			verifier: "",
			valid:    false,
		},
		{
			name:     "contains invalid character (space)",
			verifier: "dBjftJeZ4CVP mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			valid:    false,
		},
		{
			name:     "contains invalid character (+)",
			verifier: "dBjftJeZ4CVP+mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			valid:    false,
		},
		{
			name:     "contains invalid character (/)",
			verifier: "dBjftJeZ4CVP/mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			valid:    false,
		},
		{
			name:     "contains padding (=)",
			verifier: "dBjftJeZ4CVP=mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCodeVerifier(tt.verifier)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func BenchmarkGenerateCodeVerifier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateCodeVerifier()
	}
}

func BenchmarkGenerateCodeChallenge(b *testing.B) {
	verifier, _ := GenerateCodeVerifier()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GenerateCodeChallenge(verifier)
	}
}
