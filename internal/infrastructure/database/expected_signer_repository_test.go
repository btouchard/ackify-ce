//go:build integration

// SPDX-License-Identifier: AGPL-3.0-or-later
package database

import (
	"context"
	"testing"
)

func TestExpectedSignerRepository_AddExpected(t *testing.T) {
	testDB := SetupTestDB(t)
	setupExpectedSignersTable(t, testDB)
	repo := NewExpectedSignerRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name      string
		docID     string
		emails    []string
		addedBy   string
		wantError bool
	}{
		{
			name:      "add single expected signer",
			docID:     "doc-001",
			emails:    []string{"user1@example.com"},
			addedBy:   "admin@example.com",
			wantError: false,
		},
		{
			name:      "add multiple expected signers",
			docID:     "doc-002",
			emails:    []string{"user1@example.com", "user2@example.com", "user3@example.com"},
			addedBy:   "admin@example.com",
			wantError: false,
		},
		{
			name:      "add duplicate emails (should not error)",
			docID:     "doc-003",
			emails:    []string{"duplicate@example.com", "duplicate@example.com"},
			addedBy:   "admin@example.com",
			wantError: false,
		},
		{
			name:      "add empty list",
			docID:     "doc-004",
			emails:    []string{},
			addedBy:   "admin@example.com",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearExpectedSignersTable(t, testDB)

			err := repo.AddExpected(ctx, tt.docID, tt.emails, tt.addedBy)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify records were added
			if !tt.wantError && len(tt.emails) > 0 {
				signers, err := repo.ListByDocID(ctx, tt.docID)
				if err != nil {
					t.Fatalf("failed to list signers: %v", err)
				}

				expectedCount := len(uniqueStrings(tt.emails))
				if len(signers) != expectedCount {
					t.Errorf("expected %d signers, got %d", expectedCount, len(signers))
				}
			}
		})
	}
}

func TestExpectedSignerRepository_ListWithStatusByDocID(t *testing.T) {
	testDB := SetupTestDB(t)
	setupExpectedSignersTable(t, testDB)
	sigRepo := NewSignatureRepository(testDB.DB)
	expectedRepo := NewExpectedSignerRepository(testDB.DB)
	factory := NewSignatureFactory()
	ctx := context.Background()

	// Setup test data
	clearExpectedSignersTable(t, testDB)
	testDB.ClearTable(t)

	docID := "doc-status-test"
	emails := []string{"signed@example.com", "pending@example.com"}

	// Add expected signers
	err := expectedRepo.AddExpected(ctx, docID, emails, "admin@example.com")
	if err != nil {
		t.Fatalf("failed to add expected signers: %v", err)
	}

	// Add a signature for one of them
	sig := factory.CreateSignatureWithDocAndUser(docID, "user-signed", "signed@example.com")
	err = sigRepo.Create(ctx, sig)
	if err != nil {
		t.Fatalf("failed to create signature: %v", err)
	}

	// Test ListWithStatusByDocID
	signers, err := expectedRepo.ListWithStatusByDocID(ctx, docID)
	if err != nil {
		t.Fatalf("failed to list signers with status: %v", err)
	}

	if len(signers) != 2 {
		t.Fatalf("expected 2 signers, got %d", len(signers))
	}

	// Check that one has signed and one hasn't
	signedCount := 0
	pendingCount := 0
	for _, s := range signers {
		if s.HasSigned {
			signedCount++
			if s.SignedAt == nil {
				t.Error("signed signer should have signed_at timestamp")
			}
		} else {
			pendingCount++
			if s.SignedAt != nil {
				t.Error("pending signer should not have signed_at timestamp")
			}
		}
	}

	if signedCount != 1 {
		t.Errorf("expected 1 signed, got %d", signedCount)
	}
	if pendingCount != 1 {
		t.Errorf("expected 1 pending, got %d", pendingCount)
	}
}

func TestExpectedSignerRepository_GetStats(t *testing.T) {
	testDB := SetupTestDB(t)
	setupExpectedSignersTable(t, testDB)
	sigRepo := NewSignatureRepository(testDB.DB)
	expectedRepo := NewExpectedSignerRepository(testDB.DB)
	factory := NewSignatureFactory()
	ctx := context.Background()

	// Setup test data
	clearExpectedSignersTable(t, testDB)
	testDB.ClearTable(t)

	docID := "doc-stats-test"
	emails := []string{
		"user1@example.com",
		"user2@example.com",
		"user3@example.com",
		"user4@example.com",
	}

	// Add expected signers
	err := expectedRepo.AddExpected(ctx, docID, emails, "admin@example.com")
	if err != nil {
		t.Fatalf("failed to add expected signers: %v", err)
	}

	// Add signatures for 2 out of 4
	sig1 := factory.CreateSignatureWithDocAndUser(docID, "sub1", "user1@example.com")
	sig2 := factory.CreateSignatureWithDocAndUser(docID, "sub2", "user2@example.com")

	if err := sigRepo.Create(ctx, sig1); err != nil {
		t.Fatalf("failed to create sig1: %v", err)
	}
	if err := sigRepo.Create(ctx, sig2); err != nil {
		t.Fatalf("failed to create sig2: %v", err)
	}

	// Get stats
	stats, err := expectedRepo.GetStats(ctx, docID)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	// Verify stats
	if stats.DocID != docID {
		t.Errorf("expected doc_id %s, got %s", docID, stats.DocID)
	}
	if stats.ExpectedCount != 4 {
		t.Errorf("expected ExpectedCount 4, got %d", stats.ExpectedCount)
	}
	if stats.SignedCount != 2 {
		t.Errorf("expected SignedCount 2, got %d", stats.SignedCount)
	}
	if stats.PendingCount != 2 {
		t.Errorf("expected PendingCount 2, got %d", stats.PendingCount)
	}
	expectedRate := 50.0
	if stats.CompletionRate != expectedRate {
		t.Errorf("expected CompletionRate %.2f, got %.2f", expectedRate, stats.CompletionRate)
	}
}

func TestExpectedSignerRepository_Remove(t *testing.T) {
	testDB := SetupTestDB(t)
	setupExpectedSignersTable(t, testDB)
	repo := NewExpectedSignerRepository(testDB.DB)
	ctx := context.Background()

	// Setup
	clearExpectedSignersTable(t, testDB)
	docID := "doc-remove-test"
	emails := []string{"user1@example.com", "user2@example.com"}
	err := repo.AddExpected(ctx, docID, emails, "admin@example.com")
	if err != nil {
		t.Fatalf("failed to add expected signers: %v", err)
	}

	// Remove one
	err = repo.Remove(ctx, docID, "user1@example.com")
	if err != nil {
		t.Fatalf("failed to remove signer: %v", err)
	}

	// Verify only one remains
	signers, err := repo.ListByDocID(ctx, docID)
	if err != nil {
		t.Fatalf("failed to list signers: %v", err)
	}

	if len(signers) != 1 {
		t.Errorf("expected 1 signer remaining, got %d", len(signers))
	}
	if signers[0].Email != "user2@example.com" {
		t.Errorf("expected user2@example.com to remain, got %s", signers[0].Email)
	}

	// Try removing non-existent should error
	err = repo.Remove(ctx, docID, "nonexistent@example.com")
	if err == nil {
		t.Error("expected error when removing non-existent signer")
	}
}

func TestExpectedSignerRepository_IsExpected(t *testing.T) {
	testDB := SetupTestDB(t)
	setupExpectedSignersTable(t, testDB)
	repo := NewExpectedSignerRepository(testDB.DB)
	ctx := context.Background()

	// Setup
	clearExpectedSignersTable(t, testDB)
	docID := "doc-check-test"
	emails := []string{"expected@example.com"}
	err := repo.AddExpected(ctx, docID, emails, "admin@example.com")
	if err != nil {
		t.Fatalf("failed to add expected signer: %v", err)
	}

	// Check expected email
	exists, err := repo.IsExpected(ctx, docID, "expected@example.com")
	if err != nil {
		t.Fatalf("failed to check expected: %v", err)
	}
	if !exists {
		t.Error("expected email should exist")
	}

	// Check non-expected email
	exists, err = repo.IsExpected(ctx, docID, "notexpected@example.com")
	if err != nil {
		t.Fatalf("failed to check expected: %v", err)
	}
	if exists {
		t.Error("non-expected email should not exist")
	}
}

// Helper functions

func setupExpectedSignersTable(t *testing.T, testDB *TestDB) {
	t.Helper()

	schema := `
		DROP TABLE IF EXISTS expected_signers;

		CREATE TABLE expected_signers (
			id BIGSERIAL PRIMARY KEY,
			doc_id TEXT NOT NULL,
			email TEXT NOT NULL,
			added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			added_by TEXT NOT NULL,
			notes TEXT,
			UNIQUE (doc_id, email)
		);

		CREATE INDEX idx_expected_signers_doc_id ON expected_signers(doc_id);
		CREATE INDEX idx_expected_signers_email ON expected_signers(email);
	`

	_, err := testDB.DB.Exec(schema)
	if err != nil {
		t.Fatalf("failed to setup expected_signers table: %v", err)
	}
}

func clearExpectedSignersTable(t *testing.T, testDB *TestDB) {
	t.Helper()
	_, err := testDB.DB.Exec("TRUNCATE TABLE expected_signers RESTART IDENTITY")
	if err != nil {
		t.Fatalf("failed to clear expected_signers table: %v", err)
	}
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
