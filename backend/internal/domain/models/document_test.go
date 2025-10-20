// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "testing"

func TestDocument_HasChecksum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		document *Document
		expected bool
	}{
		{
			name: "Document with checksum",
			document: &Document{
				Checksum: "abc123def456",
			},
			expected: true,
		},
		{
			name: "Document without checksum",
			document: &Document{
				Checksum: "",
			},
			expected: false,
		},
		{
			name: "Document with whitespace checksum",
			document: &Document{
				Checksum: "   ",
			},
			expected: true, // Non-empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.document.HasChecksum()
			if result != tt.expected {
				t.Errorf("HasChecksum() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDocument_GetExpectedChecksumLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		checksumAlgorithm string
		expectedLength    int
	}{
		{
			name:              "SHA-256 algorithm",
			checksumAlgorithm: "SHA-256",
			expectedLength:    64,
		},
		{
			name:              "SHA-512 algorithm",
			checksumAlgorithm: "SHA-512",
			expectedLength:    128,
		},
		{
			name:              "MD5 algorithm",
			checksumAlgorithm: "MD5",
			expectedLength:    32,
		},
		{
			name:              "Unknown algorithm",
			checksumAlgorithm: "UNKNOWN",
			expectedLength:    0,
		},
		{
			name:              "Empty algorithm",
			checksumAlgorithm: "",
			expectedLength:    0,
		},
		{
			name:              "Lowercase sha-256",
			checksumAlgorithm: "sha-256",
			expectedLength:    0, // Case sensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc := &Document{
				ChecksumAlgorithm: tt.checksumAlgorithm,
			}
			result := doc.GetExpectedChecksumLength()
			if result != tt.expectedLength {
				t.Errorf("GetExpectedChecksumLength() = %v, want %v", result, tt.expectedLength)
			}
		})
	}
}
