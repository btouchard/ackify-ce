// SPDX-License-Identifier: AGPL-3.0-or-later
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
)

const (
	// PKCE code verifier length (RFC 7636 recommends 43-128 characters)
	codeVerifierLength = 43

	// Valid characters for code verifier: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
	codeVerifierCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
)

var (
	// Regex to validate code verifier format (RFC 7636)
	codeVerifierRegex = regexp.MustCompile(`^[A-Za-z0-9\-\._~]{43,128}$`)
)

// GenerateCodeVerifier generates a cryptographically secure PKCE code verifier
// The verifier is a random string of 43-128 characters using the unreserved character set.
// Returns a base64 URL-safe encoded string suitable for OAuth2 PKCE flow.
func GenerateCodeVerifier() (string, error) {
	// Generate random bytes (32 bytes = 43 characters in base64)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64 URL-safe (no padding)
	verifier := base64.RawURLEncoding.EncodeToString(randomBytes)

	// Validate the generated verifier
	if !ValidateCodeVerifier(verifier) {
		return "", fmt.Errorf("generated verifier failed validation")
	}

	return verifier, nil
}

// GenerateCodeChallenge generates a PKCE code challenge from a code verifier
// Uses the S256 method: BASE64URL(SHA256(ASCII(code_verifier)))
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// ValidateCodeVerifier validates that a code verifier meets RFC 7636 requirements
// - Length: 43-128 characters
// - Characters: [A-Za-z0-9-._~]
func ValidateCodeVerifier(verifier string) bool {
	return codeVerifierRegex.MatchString(verifier)
}
