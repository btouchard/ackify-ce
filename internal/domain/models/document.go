// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "time"

// Document represents document metadata for tracking and integrity verification
type Document struct {
	DocID             string    `json:"doc_id" db:"doc_id"`
	Title             string    `json:"title" db:"title"`
	URL               string    `json:"url" db:"url"`
	Checksum          string    `json:"checksum" db:"checksum"`
	ChecksumAlgorithm string    `json:"checksum_algorithm" db:"checksum_algorithm"`
	Description       string    `json:"description" db:"description"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy         string    `json:"created_by" db:"created_by"`
}

// DocumentInput represents the input for creating/updating document metadata
type DocumentInput struct {
	Title             string `json:"title"`
	URL               string `json:"url"`
	Checksum          string `json:"checksum"`
	ChecksumAlgorithm string `json:"checksum_algorithm"`
	Description       string `json:"description"`
}
