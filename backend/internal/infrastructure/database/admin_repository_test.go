//go:build integration

package database

import (
	"context"
	"testing"
)

func TestAdminRepository_ListDocumentsWithCounts_Integration(t *testing.T) {
	testDB := SetupTestDB(t)
	// Tables are created by migrations in SetupTestDB

	ctx := context.Background()
	repo := NewAdminRepository(testDB.DB)

	// Test listing documents - should succeed even if empty
	docs, err := repo.ListDocumentsWithCounts(ctx)
	if err != nil {
		t.Fatalf("ListDocumentsWithCounts failed: %v", err)
	}

	// docs can be nil or empty slice if no documents exist - both are valid
	_ = docs
}

func TestAdminRepository_ListSignaturesByDoc_Integration(t *testing.T) {
	testDB := SetupTestDB(t)

	_, err := testDB.DB.Exec(`
		CREATE TABLE IF NOT EXISTS signatures (
			id BIGSERIAL PRIMARY KEY,
			doc_id TEXT NOT NULL,
			user_sub TEXT NOT NULL,
			user_email TEXT NOT NULL,
			user_name TEXT,
			signed_at TIMESTAMPTZ NOT NULL,
			payload_hash TEXT NOT NULL,
			signature TEXT NOT NULL,
			nonce TEXT NOT NULL,
			referer TEXT,
			prev_hash TEXT,
			doc_checksum TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE (doc_id, user_sub)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create signatures table: %v", err)
	}

	ctx := context.Background()
	repo := NewAdminRepository(testDB.DB)

	// Test listing signatures for a doc
	sigs, err := repo.ListSignaturesByDoc(ctx, "test-doc")
	if err != nil {
		t.Fatalf("ListSignaturesByDoc failed: %v", err)
	}

	// sigs can be nil or empty if no signatures exist
	_ = sigs
}
