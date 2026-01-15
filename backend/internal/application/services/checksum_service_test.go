// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/btouchard/ackify-ce/backend/pkg/models"
)

type fakeVerificationRepository struct {
	verifications        []*models.ChecksumVerification
	shouldFailRecord     bool
	shouldFailGetHistory bool
	shouldFailGetLast    bool
}

func newFakeVerificationRepository() *fakeVerificationRepository {
	return &fakeVerificationRepository{
		verifications: make([]*models.ChecksumVerification, 0),
	}
}

func (f *fakeVerificationRepository) RecordVerification(_ context.Context, verification *models.ChecksumVerification) error {
	if f.shouldFailRecord {
		return errors.New("repository record failed")
	}

	verification.ID = int64(len(f.verifications) + 1)
	f.verifications = append(f.verifications, verification)
	return nil
}

func (f *fakeVerificationRepository) GetVerificationHistory(_ context.Context, docID string, limit int) ([]*models.ChecksumVerification, error) {
	if f.shouldFailGetHistory {
		return nil, errors.New("repository get history failed")
	}

	var result []*models.ChecksumVerification
	for _, v := range f.verifications {
		if v.DocID == docID {
			result = append(result, v)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

func (f *fakeVerificationRepository) GetLastVerification(_ context.Context, docID string) (*models.ChecksumVerification, error) {
	if f.shouldFailGetLast {
		return nil, errors.New("repository get last failed")
	}

	for i := len(f.verifications) - 1; i >= 0; i-- {
		if f.verifications[i].DocID == docID {
			return f.verifications[i], nil
		}
	}

	return nil, nil
}

type fakeDocumentRepository struct {
	documents     map[string]*models.Document
	shouldFailGet bool
}

func newFakeDocumentRepository() *fakeDocumentRepository {
	return &fakeDocumentRepository{
		documents: make(map[string]*models.Document),
	}
}

func (f *fakeDocumentRepository) GetByDocID(_ context.Context, docID string) (*models.Document, error) {
	if f.shouldFailGet {
		return nil, errors.New("repository get failed")
	}

	doc, exists := f.documents[docID]
	if !exists {
		return nil, nil
	}

	return doc, nil
}

func (f *fakeDocumentRepository) Create(_ context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	if f.shouldFailGet {
		return nil, errors.New("repository create failed")
	}

	doc := &models.Document{
		DocID:             docID,
		Title:             input.Title,
		URL:               input.URL,
		Checksum:          input.Checksum,
		ChecksumAlgorithm: input.ChecksumAlgorithm,
		Description:       input.Description,
		CreatedBy:         createdBy,
	}
	f.documents[docID] = doc
	return doc, nil
}

func (f *fakeDocumentRepository) FindByReference(_ context.Context, ref string, refType string) (*models.Document, error) {
	if f.shouldFailGet {
		return nil, errors.New("repository find failed")
	}

	for _, doc := range f.documents {
		if doc.URL == ref {
			return doc, nil
		}
	}

	return nil, nil
}

func (f *fakeDocumentRepository) List(_ context.Context, _, _ int) ([]*models.Document, error) {
	result := make([]*models.Document, 0, len(f.documents))
	for _, doc := range f.documents {
		result = append(result, doc)
	}
	return result, nil
}

func (f *fakeDocumentRepository) Search(_ context.Context, _ string, _, _ int) ([]*models.Document, error) {
	return []*models.Document{}, nil
}

func (f *fakeDocumentRepository) Count(_ context.Context, _ string) (int, error) {
	return len(f.documents), nil
}

func (f *fakeDocumentRepository) ListByCreatedBy(_ context.Context, _ string, _, _ int) ([]*models.Document, error) {
	return []*models.Document{}, nil
}

func (f *fakeDocumentRepository) SearchByCreatedBy(_ context.Context, _, _ string, _, _ int) ([]*models.Document, error) {
	return []*models.Document{}, nil
}

func (f *fakeDocumentRepository) CountByCreatedBy(_ context.Context, _, _ string) (int, error) {
	return 0, nil
}

func TestChecksumService_ValidateChecksumFormat(t *testing.T) {
	service := NewChecksumService(newFakeVerificationRepository(), newFakeDocumentRepository())

	tests := []struct {
		name      string
		checksum  string
		algorithm string
		wantError bool
	}{
		{
			name:      "valid SHA-256",
			checksum:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			algorithm: "SHA-256",
			wantError: false,
		},
		{
			name:      "valid SHA-512",
			checksum:  "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
			algorithm: "SHA-512",
			wantError: false,
		},
		{
			name:      "valid MD5",
			checksum:  "d41d8cd98f00b204e9800998ecf8427e",
			algorithm: "MD5",
			wantError: false,
		},
		{
			name:      "valid with uppercase",
			checksum:  "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
			algorithm: "SHA-256",
			wantError: false,
		},
		{
			name:      "valid with spaces",
			checksum:  "e3b0c442 98fc1c14 9afbf4c8 996fb924 27ae41e4 649b934c a495991b 7852b855",
			algorithm: "SHA-256",
			wantError: false,
		},
		{
			name:      "valid with hyphens",
			checksum:  "e3b0c442-98fc1c14-9afbf4c8-996fb924-27ae41e4-649b934c-a495991b-7852b855",
			algorithm: "SHA-256",
			wantError: false,
		},
		{
			name:      "invalid - too short for SHA-256",
			checksum:  "abc123",
			algorithm: "SHA-256",
			wantError: true,
		},
		{
			name:      "invalid - too long for MD5",
			checksum:  "d41d8cd98f00b204e9800998ecf8427eextra",
			algorithm: "MD5",
			wantError: true,
		},
		{
			name:      "invalid - non-hex characters",
			checksum:  "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			algorithm: "SHA-256",
			wantError: true,
		},
		{
			name:      "invalid - unsupported algorithm",
			checksum:  "abc123",
			algorithm: "SHA-1",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateChecksumFormat(tt.checksum, tt.algorithm)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestChecksumService_VerifyChecksum(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		docID              string
		document           *models.Document
		calculatedChecksum string
		verifiedBy         string
		wantValid          bool
		wantHasReference   bool
		wantError          bool
	}{
		{
			name:  "valid verification - checksums match",
			docID: "doc-001",
			document: &models.Document{
				DocID:             "doc-001",
				Checksum:          "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				ChecksumAlgorithm: "SHA-256",
			},
			calculatedChecksum: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			verifiedBy:         "user@example.com",
			wantValid:          true,
			wantHasReference:   true,
			wantError:          false,
		},
		{
			name:  "invalid verification - checksums differ",
			docID: "doc-002",
			document: &models.Document{
				DocID:             "doc-002",
				Checksum:          "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				ChecksumAlgorithm: "SHA-256",
			},
			calculatedChecksum: "0000000000000000000000000000000000000000000000000000000000000000",
			verifiedBy:         "user@example.com",
			wantValid:          false,
			wantHasReference:   true,
			wantError:          false,
		},
		{
			name:  "no reference checksum",
			docID: "doc-003",
			document: &models.Document{
				DocID:             "doc-003",
				Checksum:          "",
				ChecksumAlgorithm: "SHA-256",
			},
			calculatedChecksum: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			verifiedBy:         "user@example.com",
			wantValid:          false,
			wantHasReference:   false,
			wantError:          false,
		},
		{
			name:  "case insensitive comparison",
			docID: "doc-004",
			document: &models.Document{
				DocID:             "doc-004",
				Checksum:          "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
				ChecksumAlgorithm: "SHA-256",
			},
			calculatedChecksum: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			verifiedBy:         "user@example.com",
			wantValid:          true,
			wantHasReference:   true,
			wantError:          false,
		},
		{
			name:               "document not found",
			docID:              "non-existent",
			document:           nil,
			calculatedChecksum: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			verifiedBy:         "user@example.com",
			wantValid:          false,
			wantHasReference:   false,
			wantError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verificationRepo := newFakeVerificationRepository()
			documentRepo := newFakeDocumentRepository()

			if tt.document != nil {
				documentRepo.documents[tt.docID] = tt.document
			}

			service := NewChecksumService(verificationRepo, documentRepo)

			result, err := service.VerifyChecksum(ctx, tt.docID, tt.calculatedChecksum, tt.verifiedBy)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Valid != tt.wantValid {
				t.Errorf("expected Valid=%v, got %v", tt.wantValid, result.Valid)
			}

			if result.HasReferenceHash != tt.wantHasReference {
				t.Errorf("expected HasReferenceHash=%v, got %v", tt.wantHasReference, result.HasReferenceHash)
			}

			// Check that verification was recorded (if document has checksum)
			if tt.wantHasReference {
				if len(verificationRepo.verifications) != 1 {
					t.Errorf("expected 1 verification recorded, got %d", len(verificationRepo.verifications))
				} else {
					v := verificationRepo.verifications[0]
					if v.IsValid != tt.wantValid {
						t.Errorf("recorded verification IsValid=%v, expected %v", v.IsValid, tt.wantValid)
					}
					if v.VerifiedBy != tt.verifiedBy {
						t.Errorf("recorded verification VerifiedBy=%s, expected %s", v.VerifiedBy, tt.verifiedBy)
					}
				}
			}
		})
	}
}

func TestChecksumService_GetVerificationHistory(t *testing.T) {
	ctx := context.Background()
	verificationRepo := newFakeVerificationRepository()
	documentRepo := newFakeDocumentRepository()
	service := NewChecksumService(verificationRepo, documentRepo)

	// Add test verifications
	for i := 0; i < 5; i++ {
		v := &models.ChecksumVerification{
			DocID:              "doc-001",
			VerifiedBy:         "user@example.com",
			VerifiedAt:         time.Now(),
			StoredChecksum:     "abc123",
			CalculatedChecksum: "abc123",
			Algorithm:          "SHA-256",
			IsValid:            true,
		}
		_ = verificationRepo.RecordVerification(ctx, v)
	}

	// Test get all
	history, err := service.GetVerificationHistory(ctx, "doc-001", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 5 {
		t.Errorf("expected 5 verifications, got %d", len(history))
	}

	// Test with limit
	limited, err := service.GetVerificationHistory(ctx, "doc-001", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(limited) != 2 {
		t.Errorf("expected 2 verifications with limit, got %d", len(limited))
	}
}

func TestChecksumService_GetChecksumInfo(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		docID     string
		document  *models.Document
		wantError bool
	}{
		{
			name:  "document with checksum",
			docID: "doc-001",
			document: &models.Document{
				DocID:             "doc-001",
				Checksum:          "abc123",
				ChecksumAlgorithm: "SHA-256",
			},
			wantError: false,
		},
		{
			name:  "document without checksum",
			docID: "doc-002",
			document: &models.Document{
				DocID:             "doc-002",
				Checksum:          "",
				ChecksumAlgorithm: "SHA-256",
			},
			wantError: false,
		},
		{
			name:      "document not found",
			docID:     "non-existent",
			document:  nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			documentRepo := newFakeDocumentRepository()
			if tt.document != nil {
				documentRepo.documents[tt.docID] = tt.document
			}

			service := NewChecksumService(newFakeVerificationRepository(), documentRepo)

			info, err := service.GetChecksumInfo(ctx, tt.docID)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if info["doc_id"] != tt.docID {
				t.Errorf("expected doc_id %s, got %v", tt.docID, info["doc_id"])
			}

			if _, ok := info["has_checksum"]; !ok {
				t.Error("expected has_checksum field")
			}

			if _, ok := info["supported_algorithms"]; !ok {
				t.Error("expected supported_algorithms field")
			}
		})
	}
}
