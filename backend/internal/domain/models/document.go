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
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy         string     `json:"created_by" db:"created_by"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// DocumentInput represents the input for creating/updating document metadata
type DocumentInput struct {
	Title             string `json:"title"`
	URL               string `json:"url"`
	Checksum          string `json:"checksum"`
	ChecksumAlgorithm string `json:"checksum_algorithm"`
	Description       string `json:"description"`
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
