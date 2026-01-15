// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSignature_JSONSerialization(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)
	createdAt := time.Date(2024, 1, 15, 10, 30, 46, 0, time.UTC)
	referer := "https://github.com/user/repo"
	prevHash := "abcd1234efgh5678"

	signature := &Signature{
		ID:          123,
		DocID:       "test-doc-123",
		UserSub:     "google-oauth2|123456789",
		UserEmail:   "test@example.com",
		UserName:    "Test User",
		SignedAtUTC: timestamp,
		PayloadHash: "SGVsbG8gV29ybGQ=",
		Signature:   "c2lnbmF0dXJlLWRhdGE=",
		Nonce:       "random-nonce-123",
		CreatedAt:   createdAt,
		Referer:     &referer,
		PrevHash:    &prevHash,
	}

	data, err := json.Marshal(signature)
	if err != nil {
		t.Fatalf("Failed to marshal signature: %v", err)
	}

	var unmarshaled Signature
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal signature: %v", err)
	}

	if unmarshaled.ID != signature.ID {
		t.Errorf("ID mismatch: got %v, expected %v", unmarshaled.ID, signature.ID)
	}
	if unmarshaled.DocID != signature.DocID {
		t.Errorf("DocID mismatch: got %v, expected %v", unmarshaled.DocID, signature.DocID)
	}
	if unmarshaled.UserSub != signature.UserSub {
		t.Errorf("UserSub mismatch: got %v, expected %v", unmarshaled.UserSub, signature.UserSub)
	}
	if unmarshaled.UserEmail != signature.UserEmail {
		t.Errorf("UserEmail mismatch: got %v, expected %v", unmarshaled.UserEmail, signature.UserEmail)
	}
	if unmarshaled.UserName != signature.UserName {
		t.Errorf("UserName mismatch: got %v, expected %v", unmarshaled.UserName, signature.UserName)
	}
	if !unmarshaled.SignedAtUTC.Equal(signature.SignedAtUTC) {
		t.Errorf("SignedAtUTC mismatch: got %v, expected %v", unmarshaled.SignedAtUTC, signature.SignedAtUTC)
	}
	if unmarshaled.PayloadHash != signature.PayloadHash {
		t.Errorf("PayloadHash mismatch: got %v, expected %v", unmarshaled.PayloadHash, signature.PayloadHash)
	}
	if unmarshaled.Signature != signature.Signature {
		t.Errorf("Signature mismatch: got %v, expected %v", unmarshaled.Signature, signature.Signature)
	}
	if unmarshaled.Nonce != signature.Nonce {
		t.Errorf("Nonce mismatch: got %v, expected %v", unmarshaled.Nonce, signature.Nonce)
	}
	if !unmarshaled.CreatedAt.Equal(signature.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v, expected %v", unmarshaled.CreatedAt, signature.CreatedAt)
	}
	if (unmarshaled.Referer == nil) != (signature.Referer == nil) {
		t.Errorf("Referer nil mismatch: got %v, expected %v", unmarshaled.Referer == nil, signature.Referer == nil)
	}
	if unmarshaled.Referer != nil && signature.Referer != nil && *unmarshaled.Referer != *signature.Referer {
		t.Errorf("Referer mismatch: got %v, expected %v", *unmarshaled.Referer, *signature.Referer)
	}
	if (unmarshaled.PrevHash == nil) != (signature.PrevHash == nil) {
		t.Errorf("PrevHash nil mismatch: got %v, expected %v", unmarshaled.PrevHash == nil, signature.PrevHash == nil)
	}
	if unmarshaled.PrevHash != nil && signature.PrevHash != nil && *unmarshaled.PrevHash != *signature.PrevHash {
		t.Errorf("PrevHash mismatch: got %v, expected %v", *unmarshaled.PrevHash, *signature.PrevHash)
	}
}

func TestSignature_JSONSerializationWithNilFields(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	createdAt := time.Date(2024, 1, 15, 10, 30, 46, 0, time.UTC)

	signature := &Signature{
		ID:          456,
		DocID:       "minimal-doc",
		UserSub:     "github|987654321",
		UserEmail:   "minimal@example.com",
		UserName:    "",
		SignedAtUTC: timestamp,
		PayloadHash: "bWluaW1hbA==",
		Signature:   "bWluaW1hbC1zaWc=",
		Nonce:       "minimal-nonce",
		CreatedAt:   createdAt,
		Referer:     nil,
		PrevHash:    nil,
	}

	data, err := json.Marshal(signature)
	if err != nil {
		t.Fatalf("Failed to marshal signature: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "user_name") {
		t.Error("user_name should be omitted when nil")
	}
	if strings.Contains(jsonStr, "referer") {
		t.Error("referer should be omitted when nil")
	}
	if strings.Contains(jsonStr, "prev_hash") {
		t.Error("prev_hash should be omitted when nil")
	}

	var unmarshaled Signature
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal signature: %v", err)
	}

	if unmarshaled.UserName != "" {
		t.Errorf("UserName should be empty string, got %v", unmarshaled.UserName)
	}
	if unmarshaled.Referer != nil {
		t.Errorf("Referer should be nil, got %v", unmarshaled.Referer)
	}
	if unmarshaled.PrevHash != nil {
		t.Errorf("PrevHash should be nil, got %v", unmarshaled.PrevHash)
	}
}

func TestSignature_GetServiceInfo(t *testing.T) {
	tests := []struct {
		name            string
		referer         *string
		expectedService *string
		expectedIcon    *string
		expectedType    *string
	}{
		{
			name:            "GitHub referer param",
			referer:         stringPtr("github"),
			expectedService: stringPtr("GitHub"),
			expectedIcon:    stringPtr("https://cdn.simpleicons.org/github"),
			expectedType:    stringPtr("code"),
		},
		{
			name:            "GitLab referer param",
			referer:         stringPtr("gitlab"),
			expectedService: stringPtr("GitLab"),
			expectedIcon:    stringPtr("https://cdn.simpleicons.org/gitlab"),
			expectedType:    stringPtr("code"),
		},
		{
			name:            "Google Docs referer param",
			referer:         stringPtr("google-docs"),
			expectedService: stringPtr("Google Docs"),
			expectedIcon:    stringPtr("https://cdn.simpleicons.org/googledocs"),
			expectedType:    stringPtr("docs"),
		},
		{
			name:            "Google Sheets referer param",
			referer:         stringPtr("google-sheets"),
			expectedService: stringPtr("Google Sheets"),
			expectedIcon:    stringPtr("https://cdn.simpleicons.org/googlesheets"),
			expectedType:    stringPtr("sheets"),
		},
		{
			name:            "Notion referer param",
			referer:         stringPtr("notion"),
			expectedService: stringPtr("Notion"),
			expectedIcon:    stringPtr("https://cdn.simpleicons.org/notion"),
			expectedType:    stringPtr("notes"),
		},
		{
			name:            "nil referer",
			referer:         nil,
			expectedService: nil,
			expectedIcon:    nil,
			expectedType:    nil,
		},
		{
			name:            "empty referer",
			referer:         stringPtr(""),
			expectedService: nil,
			expectedIcon:    nil,
			expectedType:    nil,
		},
		{
			name:            "custom referer param",
			referer:         stringPtr("custom-service"),
			expectedService: stringPtr("custom-service"),
			expectedIcon:    stringPtr("https://cdn.simpleicons.org/link"),
			expectedType:    stringPtr("custom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := &Signature{
				Referer: tt.referer,
			}

			serviceInfo := signature.GetServiceInfo()

			if tt.expectedService == nil {
				if serviceInfo != nil {
					t.Errorf("Expected nil service info, got %+v", serviceInfo)
				}
				return
			}

			if serviceInfo == nil {
				t.Errorf("Expected service info, got nil")
				return
			}

			if serviceInfo.Name != *tt.expectedService {
				t.Errorf("Service name mismatch: got %v, expected %v", serviceInfo.Name, *tt.expectedService)
			}
			if serviceInfo.Icon != *tt.expectedIcon {
				t.Errorf("Service icon mismatch: got %v, expected %v", serviceInfo.Icon, *tt.expectedIcon)
			}
			if serviceInfo.Type != *tt.expectedType {
				t.Errorf("Service type mismatch: got %v, expected %v", serviceInfo.Type, *tt.expectedType)
			}
		})
	}
}

func TestSignature_ComputeRecordHash(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)
	createdAt := time.Date(2024, 1, 15, 10, 30, 46, 0, time.UTC)
	referer := "https://github.com/user/repo"

	signature := &Signature{
		ID:          123,
		DocID:       "test-doc-123",
		UserSub:     "google-oauth2|123456789",
		UserEmail:   "test@example.com",
		UserName:    "Test User",
		SignedAtUTC: timestamp,
		PayloadHash: "SGVsbG8gV29ybGQ=",
		Signature:   "c2lnbmF0dXJlLWRhdGE=",
		Nonce:       "random-nonce-123",
		CreatedAt:   createdAt,
		Referer:     &referer,
	}

	hash1 := signature.ComputeRecordHash()
	hash2 := signature.ComputeRecordHash()

	if hash1 != hash2 {
		t.Errorf("Hash computation is not deterministic: %v != %v", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("Hash should not be empty")
	}

	if !isValidBase64(hash1) {
		t.Errorf("Hash is not valid base64: %v", hash1)
	}

	originalID := signature.ID
	signature.ID = 456
	hashChanged := signature.ComputeRecordHash()
	if hashChanged == hash1 {
		t.Error("Hash should change when ID changes")
	}
	signature.ID = originalID

	signature.UserName = ""
	hashWithEmptyName := signature.ComputeRecordHash()
	if hashWithEmptyName == hash1 {
		t.Error("Hash should change when UserName becomes empty")
	}

	signature.UserName = "Test User"
	signature.Referer = nil
	hashWithNilReferer := signature.ComputeRecordHash()
	if hashWithNilReferer == hash1 {
		t.Error("Hash should change when Referer becomes nil")
	}
}

func TestSignature_ComputeRecordHashDeterministic(t *testing.T) {
	// Test that the same signature data produces the same hash
	timestamp := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)
	createdAt := time.Date(2024, 1, 15, 10, 30, 46, 0, time.UTC)
	referer := "https://github.com/user/repo"

	sig1 := &Signature{
		ID:          123,
		DocID:       "test-doc-123",
		UserSub:     "google-oauth2|123456789",
		UserEmail:   "test@example.com",
		UserName:    "Test User",
		SignedAtUTC: timestamp,
		PayloadHash: "SGVsbG8gV29ybGQ=",
		Signature:   "c2lnbmF0dXJlLWRhdGE=",
		Nonce:       "random-nonce-123",
		CreatedAt:   createdAt,
		Referer:     &referer,
	}

	sig2 := &Signature{
		ID:          123,
		DocID:       "test-doc-123",
		UserSub:     "google-oauth2|123456789",
		UserEmail:   "test@example.com",
		UserName:    "Test User",
		SignedAtUTC: timestamp,
		PayloadHash: "SGVsbG8gV29ybGQ=",
		Signature:   "c2lnbmF0dXJlLWRhdGE=",
		Nonce:       "random-nonce-123",
		CreatedAt:   createdAt,
		Referer:     &referer,
	}

	hash1 := sig1.ComputeRecordHash()
	hash2 := sig2.ComputeRecordHash()

	if hash1 != hash2 {
		t.Errorf("Identical signatures should produce identical hashes: %v != %v", hash1, hash2)
	}
}

func TestSignatureRequest_Validation(t *testing.T) {
	validUser := &User{
		Sub:   "google-oauth2|123456789",
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name    string
		request SignatureRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: SignatureRequest{
				DocID:   "valid-doc-123",
				User:    validUser,
				Referer: stringPtr("https://github.com/user/repo"),
			},
			valid: true,
		},
		{
			name: "valid request without referer",
			request: SignatureRequest{
				DocID: "valid-doc-123",
				User:  validUser,
			},
			valid: true,
		},
		{
			name: "invalid request - empty DocID",
			request: SignatureRequest{
				DocID: "",
				User:  validUser,
			},
			valid: false,
		},
		{
			name: "invalid request - nil user",
			request: SignatureRequest{
				DocID: "valid-doc-123",
				User:  nil,
			},
			valid: false,
		},
		{
			name: "invalid request - invalid user",
			request: SignatureRequest{
				DocID: "valid-doc-123",
				User: &User{
					Sub:   "",
					Email: "test@example.com",
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidDocID := tt.request.DocID != ""
			hasValidUser := tt.request.User != nil && tt.request.User.IsValid()
			isValid := hasValidDocID && hasValidUser

			if isValid != tt.valid {
				t.Errorf("Request validation mismatch: got %v, expected %v for %+v", isValid, tt.valid, tt.request)
			}
		})
	}
}

func TestSignature_TimestampValidation(t *testing.T) {
	tests := []struct {
		name        string
		signedAt    time.Time
		createdAt   time.Time
		expectValid bool
	}{
		{
			name:        "valid timestamps - signedAt before createdAt",
			signedAt:    time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			createdAt:   time.Date(2024, 1, 15, 10, 30, 46, 0, time.UTC),
			expectValid: true,
		},
		{
			name:        "valid timestamps - same time",
			signedAt:    time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			createdAt:   time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			expectValid: true,
		},
		{
			name:        "questionable timestamps - createdAt before signedAt",
			signedAt:    time.Date(2024, 1, 15, 10, 30, 46, 0, time.UTC),
			createdAt:   time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := &Signature{
				SignedAtUTC: tt.signedAt,
				CreatedAt:   tt.createdAt,
			}

			isValid := !signature.CreatedAt.Before(signature.SignedAtUTC)
			if isValid != tt.expectValid {
				t.Errorf("Timestamp validation mismatch: got %v, expected %v", isValid, tt.expectValid)
			}

			if signature.SignedAtUTC.Location() != time.UTC {
				t.Error("SignedAtUTC should be in UTC timezone")
			}
			if signature.CreatedAt.Location() != time.UTC {
				t.Error("CreatedAt should be in UTC timezone")
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func isValidBase64(s string) bool {
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	for _, char := range s {
		found := false
		for _, valid := range validChars {
			if char == valid {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return len(s) > 0
}
