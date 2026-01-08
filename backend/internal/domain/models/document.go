// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"time"

	"github.com/google/uuid"
)

// Document represents document metadata for tracking and integrity verification
type Document struct {
	DocID             string     `json:"doc_id" db:"doc_id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Title             string     `json:"title" db:"title"`
	URL               string     `json:"url" db:"url"`
	Checksum          string     `json:"checksum" db:"checksum"`
	ChecksumAlgorithm string     `json:"checksum_algorithm" db:"checksum_algorithm"`
	Description       string     `json:"description" db:"description"`
	ReadMode          string     `json:"read_mode" db:"read_mode"`
	AllowDownload     bool       `json:"allow_download" db:"allow_download"`
	RequireFullRead   bool       `json:"require_full_read" db:"require_full_read"`
	VerifyChecksum    bool       `json:"verify_checksum" db:"verify_checksum"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy         string     `json:"created_by" db:"created_by"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Storage fields for uploaded files
	StorageKey       string `json:"storage_key,omitempty" db:"storage_key"`
	StorageProvider  string `json:"storage_provider,omitempty" db:"storage_provider"`
	FileSize         int64  `json:"file_size,omitempty" db:"file_size"`
	MimeType         string `json:"mime_type,omitempty" db:"mime_type"`
	OriginalFilename string `json:"original_filename,omitempty" db:"original_filename"`
}

// DocumentInput represents the input for creating/updating document metadata
type DocumentInput struct {
	Title             string `json:"title"`
	URL               string `json:"url"`
	Checksum          string `json:"checksum"`
	ChecksumAlgorithm string `json:"checksum_algorithm"`
	Description       string `json:"description"`
	ReadMode          string `json:"read_mode"`
	AllowDownload     *bool  `json:"allow_download"`
	RequireFullRead   *bool  `json:"require_full_read"`
	VerifyChecksum    *bool  `json:"verify_checksum"`

	// Storage fields for uploaded files
	StorageKey       string `json:"storage_key,omitempty"`
	StorageProvider  string `json:"storage_provider,omitempty"`
	FileSize         int64  `json:"file_size,omitempty"`
	MimeType         string `json:"mime_type,omitempty"`
	OriginalFilename string `json:"original_filename,omitempty"`
}

// IsStored returns true if the document has an uploaded file
func (d *Document) IsStored() bool {
	return d.StorageKey != "" && d.StorageProvider != ""
}

// HasChecksum returns true if the document has a checksum configured
func (d *Document) HasChecksum() bool {
	return d.Checksum != ""
}

// GetExpectedChecksumLength returns the expected length for the configured algorithm
func (d *Document) GetExpectedChecksumLength() int {
	switch d.ChecksumAlgorithm {
	case "SHA-256":
		return 64
	case "SHA-512":
		return 128
	case "MD5":
		return 32
	default:
		return 0
	}
}

// GetDocID returns the document ID
func (d *Document) GetDocID() string {
	return d.DocID
}

// GetTitle returns the document title
func (d *Document) GetTitle() string {
	return d.Title
}

// GetURL returns the document URL
func (d *Document) GetURL() string {
	return d.URL
}
