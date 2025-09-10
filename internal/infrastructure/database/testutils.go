//go:build integration

package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"ackify/internal/domain/models"
	_ "github.com/lib/pq"
)

// TestDB holds test database configuration
type TestDB struct {
	DB     *sql.DB
	DSN    string
	dbName string
}

// SetupTestDB creates a test database connection and runs migrations
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Skip if not in integrations test mode
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integrations test (INTEGRATION_TESTS not set)")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:testpassword@localhost:5432/ackify_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	testDB := &TestDB{
		DB:     db,
		DSN:    dsn,
		dbName: fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), os.Getpid()),
	}

	// Create test schema
	if err := testDB.createSchema(); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Clean up on test completion
	t.Cleanup(func() {
		testDB.Cleanup()
	})

	return testDB
}

// createSchema creates the signatures table for testing
func (tdb *TestDB) createSchema() error {
	schema := `
		-- Drop table if exists (for cleanup)
		DROP TABLE IF EXISTS signatures;

		-- Create signatures table
		CREATE TABLE signatures (
			id BIGSERIAL PRIMARY KEY,
			doc_id TEXT NOT NULL,
			user_sub TEXT NOT NULL,
			user_email TEXT NOT NULL,
			user_name TEXT,
			signed_at_utc TIMESTAMPTZ NOT NULL,
			payload_hash_b64 TEXT NOT NULL,
			signature_b64 TEXT NOT NULL,
			nonce TEXT NOT NULL,
			referer TEXT,
			prev_hash_b64 TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			
			-- Constraints
			UNIQUE (doc_id, user_sub)
		);

		-- Create indexes for performance
		CREATE INDEX idx_signatures_doc_id ON signatures(doc_id);
		CREATE INDEX idx_signatures_user_sub ON signatures(user_sub);
		CREATE INDEX idx_signatures_user_email ON signatures(user_email);
		CREATE INDEX idx_signatures_created_at ON signatures(created_at);
		CREATE INDEX idx_signatures_id_asc ON signatures(id ASC);
	`

	_, err := tdb.DB.Exec(schema)
	return err
}

// Cleanup closes the database connection and cleans up
func (tdb *TestDB) Cleanup() {
	if tdb.DB != nil {
		// Drop all tables for cleanup
		_, _ = tdb.DB.Exec("DROP TABLE IF EXISTS signatures")
		_ = tdb.DB.Close()
	}
}

// ClearTable removes all data from the signatures table
func (tdb *TestDB) ClearTable(t *testing.T) {
	t.Helper()
	_, err := tdb.DB.Exec("TRUNCATE TABLE signatures RESTART IDENTITY")
	if err != nil {
		t.Fatalf("Failed to clear signatures table: %v", err)
	}
}

// GetTableCount returns the number of rows in signatures table
func (tdb *TestDB) GetTableCount(t *testing.T) int {
	t.Helper()
	var count int
	err := tdb.DB.QueryRow("SELECT COUNT(*) FROM signatures").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to get table count: %v", err)
	}
	return count
}

// SignatureFactory creates test signature objects
type SignatureFactory struct{}

// CreateValidSignature creates a valid signature for testing
func (f *SignatureFactory) CreateValidSignature() *models.Signature {
	now := time.Now().UTC()
	userName := "Test User"
	referer := "https://example.com/doc"

	return &models.Signature{
		DocID:          "test-doc-123",
		UserSub:        "user-123",
		UserEmail:      "test@example.com",
		UserName:       &userName,
		SignedAtUTC:    now,
		PayloadHashB64: "dGVzdC1wYXlsb2FkLWhhc2g=", // base64("test-payload-hash")
		SignatureB64:   "dGVzdC1zaWduYXR1cmU=",     // base64("test-signature")
		Nonce:          "test-nonce-123",
		Referer:        &referer,
		PrevHashB64:    nil, // Will be set for chained signatures
	}
}

// CreateSignatureWithDoc creates a signature for a specific document
func (f *SignatureFactory) CreateSignatureWithDoc(docID string) *models.Signature {
	sig := f.CreateValidSignature()
	sig.DocID = docID
	return sig
}

// CreateSignatureWithUser creates a signature for a specific user
func (f *SignatureFactory) CreateSignatureWithUser(userSub, userEmail string) *models.Signature {
	sig := f.CreateValidSignature()
	sig.UserSub = userSub
	sig.UserEmail = userEmail
	return sig
}

// CreateSignatureWithDocAndUser creates a signature for specific doc and user
func (f *SignatureFactory) CreateSignatureWithDocAndUser(docID, userSub, userEmail string) *models.Signature {
	sig := f.CreateValidSignature()
	sig.DocID = docID
	sig.UserSub = userSub
	sig.UserEmail = userEmail
	return sig
}

// CreateChainedSignature creates a signature with previous hash for chaining tests
func (f *SignatureFactory) CreateChainedSignature(prevHashB64 string) *models.Signature {
	sig := f.CreateValidSignature()
	sig.PrevHashB64 = &prevHashB64
	return sig
}

// CreateMinimalSignature creates signature with only required fields
func (f *SignatureFactory) CreateMinimalSignature() *models.Signature {
	now := time.Now().UTC()

	return &models.Signature{
		DocID:          "minimal-doc",
		UserSub:        "minimal-user",
		UserEmail:      "minimal@example.com",
		UserName:       nil, // NULL
		SignedAtUTC:    now,
		PayloadHashB64: "bWluaW1hbA==", // base64("minimal")
		SignatureB64:   "bWluaW1hbA==", // base64("minimal")
		Nonce:          "minimal-nonce",
		Referer:        nil, // NULL
		PrevHashB64:    nil, // NULL
	}
}

// AssertSignatureEqual compares two signatures for testing
func AssertSignatureEqual(t *testing.T, expected, actual *models.Signature) {
	t.Helper()

	if actual.DocID != expected.DocID {
		t.Errorf("DocID mismatch: got %s, want %s", actual.DocID, expected.DocID)
	}

	if actual.UserSub != expected.UserSub {
		t.Errorf("UserSub mismatch: got %s, want %s", actual.UserSub, expected.UserSub)
	}

	if actual.UserEmail != expected.UserEmail {
		t.Errorf("UserEmail mismatch: got %s, want %s", actual.UserEmail, expected.UserEmail)
	}

	if !isStringPtrEqual(actual.UserName, expected.UserName) {
		t.Errorf("UserName mismatch: got %v, want %v", actual.UserName, expected.UserName)
	}

	if actual.PayloadHashB64 != expected.PayloadHashB64 {
		t.Errorf("PayloadHashB64 mismatch: got %s, want %s", actual.PayloadHashB64, expected.PayloadHashB64)
	}

	if actual.SignatureB64 != expected.SignatureB64 {
		t.Errorf("SignatureB64 mismatch: got %s, want %s", actual.SignatureB64, expected.SignatureB64)
	}

	if actual.Nonce != expected.Nonce {
		t.Errorf("Nonce mismatch: got %s, want %s", actual.Nonce, expected.Nonce)
	}

	if !isStringPtrEqual(actual.Referer, expected.Referer) {
		t.Errorf("Referer mismatch: got %v, want %v", actual.Referer, expected.Referer)
	}

	if !isStringPtrEqual(actual.PrevHashB64, expected.PrevHashB64) {
		t.Errorf("PrevHashB64 mismatch: got %v, want %v", actual.PrevHashB64, expected.PrevHashB64)
	}
}

// isStringPtrEqual compares two string pointers
func isStringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// NewSignatureFactory creates a new signature factory
func NewSignatureFactory() *SignatureFactory {
	return &SignatureFactory{}
}
