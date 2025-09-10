package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"ackify/internal/domain/models"
)

// Mock repository implementation
type fakeRepository struct {
	signatures        map[string]*models.Signature // key: docID_userSub
	allSignatures     []*models.Signature
	shouldFailCreate  bool
	shouldFailGet     bool
	shouldFailExists  bool
	shouldFailGetLast bool
	shouldFailGetAll  bool
	shouldFailCheck   bool
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		signatures:    make(map[string]*models.Signature),
		allSignatures: make([]*models.Signature, 0),
	}
}

func (f *fakeRepository) Create(ctx context.Context, signature *models.Signature) error {
	if f.shouldFailCreate {
		return errors.New("repository create failed")
	}

	signature.ID = int64(len(f.allSignatures) + 1)
	signature.CreatedAt = time.Now().UTC()

	key := signature.DocID + "_" + signature.UserSub
	f.signatures[key] = signature
	f.allSignatures = append(f.allSignatures, signature)

	return nil
}

func (f *fakeRepository) GetByDocAndUser(ctx context.Context, docID, userSub string) (*models.Signature, error) {
	if f.shouldFailGet {
		return nil, errors.New("repository get failed")
	}

	key := docID + "_" + userSub
	signature, exists := f.signatures[key]
	if !exists {
		return nil, models.ErrSignatureNotFound
	}

	return signature, nil
}

func (f *fakeRepository) GetByDoc(ctx context.Context, docID string) ([]*models.Signature, error) {
	if f.shouldFailGet {
		return nil, errors.New("repository get failed")
	}

	var result []*models.Signature
	for _, sig := range f.signatures {
		if sig.DocID == docID {
			result = append(result, sig)
		}
	}

	return result, nil
}

func (f *fakeRepository) GetByUser(ctx context.Context, userSub string) ([]*models.Signature, error) {
	if f.shouldFailGet {
		return nil, errors.New("repository get failed")
	}

	var result []*models.Signature
	for _, sig := range f.signatures {
		if sig.UserSub == userSub {
			result = append(result, sig)
		}
	}

	return result, nil
}

func (f *fakeRepository) ExistsByDocAndUser(ctx context.Context, docID, userSub string) (bool, error) {
	if f.shouldFailExists {
		return false, errors.New("repository exists failed")
	}

	key := docID + "_" + userSub
	_, exists := f.signatures[key]
	return exists, nil
}

func (f *fakeRepository) CheckUserSignatureStatus(ctx context.Context, docID, userIdentifier string) (bool, error) {
	if f.shouldFailCheck {
		return false, errors.New("repository check failed")
	}

	for _, sig := range f.signatures {
		if sig.DocID == docID && (sig.UserSub == userIdentifier || sig.UserEmail == userIdentifier) {
			return true, nil
		}
	}

	return false, nil
}

func (f *fakeRepository) GetLastSignature(ctx context.Context) (*models.Signature, error) {
	if f.shouldFailGetLast {
		return nil, errors.New("repository get last failed")
	}

	if len(f.allSignatures) == 0 {
		return nil, nil
	}

	return f.allSignatures[len(f.allSignatures)-1], nil
}

func (f *fakeRepository) GetAllSignaturesOrdered(ctx context.Context) ([]*models.Signature, error) {
	if f.shouldFailGetAll {
		return nil, errors.New("repository get all failed")
	}

	return f.allSignatures, nil
}

// Mock crypto signer implementation
type fakeCryptoSigner struct {
	shouldFail bool
}

func newFakeCryptoSigner() *fakeCryptoSigner {
	return &fakeCryptoSigner{}
}

func (f *fakeCryptoSigner) CreateSignature(docID string, user *models.User, timestamp time.Time, nonce string) (string, string, error) {
	if f.shouldFail {
		return "", "", errors.New("crypto signing failed")
	}

	payloadHash := "fake-payload-hash-" + docID
	signature := "fake-signature-" + user.Sub
	return payloadHash, signature, nil
}

// Test NewSignatureService
func TestNewSignatureService(t *testing.T) {
	repo := newFakeRepository()
	signer := newFakeCryptoSigner()

	service := NewSignatureService(repo, signer)

	if service == nil {
		t.Error("NewSignatureService should not return nil")
	} else if service.repo != repo {
		t.Error("Service repository not set correctly")
	} else if service.signer != signer {
		t.Error("Service signer not set correctly")
	}
}

func TestSignatureService_CreateSignature(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.SignatureRequest
		setupRepo     func(*fakeRepository)
		setupSigner   func(*fakeCryptoSigner)
		expectError   bool
		expectedError error
	}{
		{
			name: "valid signature creation",
			request: &models.SignatureRequest{
				DocID: "test-doc-123",
				User: &models.User{
					Sub:   "user-123",
					Email: "test@example.com",
					Name:  "Test User",
				},
				Referer: stringPtr("github"),
			},
			setupRepo:   func(r *fakeRepository) {},
			setupSigner: func(s *fakeCryptoSigner) {},
			expectError: false,
		},
		{
			name: "invalid user - nil",
			request: &models.SignatureRequest{
				DocID: "test-doc-123",
				User:  nil,
			},
			expectError:   true,
			expectedError: models.ErrInvalidUser,
		},
		{
			name: "invalid user - invalid data",
			request: &models.SignatureRequest{
				DocID: "test-doc-123",
				User: &models.User{
					Sub:   "",
					Email: "test@example.com",
				},
			},
			expectError:   true,
			expectedError: models.ErrInvalidUser,
		},
		{
			name: "empty document ID",
			request: &models.SignatureRequest{
				DocID: "",
				User: &models.User{
					Sub:   "user-123",
					Email: "test@example.com",
				},
			},
			expectError:   true,
			expectedError: models.ErrInvalidDocument,
		},
		{
			name: "signature already exists",
			request: &models.SignatureRequest{
				DocID: "existing-doc",
				User: &models.User{
					Sub:   "existing-user",
					Email: "existing@example.com",
				},
			},
			setupRepo: func(r *fakeRepository) {
				// Pre-populate with existing signature
				r.signatures["existing-doc_existing-user"] = &models.Signature{
					ID:      1,
					DocID:   "existing-doc",
					UserSub: "existing-user",
				}
			},
			setupSigner:   func(s *fakeCryptoSigner) {},
			expectError:   true,
			expectedError: models.ErrSignatureAlreadyExists,
		},
		{
			name: "repository exists check fails",
			request: &models.SignatureRequest{
				DocID: "test-doc",
				User: &models.User{
					Sub:   "user-123",
					Email: "test@example.com",
				},
			},
			setupRepo: func(r *fakeRepository) {
				r.shouldFailExists = true
			},
			setupSigner: func(s *fakeCryptoSigner) {},
			expectError: true,
		},
		{
			name: "crypto signing fails",
			request: &models.SignatureRequest{
				DocID: "test-doc",
				User: &models.User{
					Sub:   "user-123",
					Email: "test@example.com",
				},
			},
			setupRepo: func(r *fakeRepository) {},
			setupSigner: func(s *fakeCryptoSigner) {
				s.shouldFail = true
			},
			expectError: true,
		},
		{
			name: "repository get last signature fails",
			request: &models.SignatureRequest{
				DocID: "test-doc",
				User: &models.User{
					Sub:   "user-123",
					Email: "test@example.com",
				},
			},
			setupRepo: func(r *fakeRepository) {
				r.shouldFailGetLast = true
			},
			setupSigner: func(s *fakeCryptoSigner) {},
			expectError: true,
		},
		{
			name: "repository create fails",
			request: &models.SignatureRequest{
				DocID: "test-doc",
				User: &models.User{
					Sub:   "user-123",
					Email: "test@example.com",
				},
			},
			setupRepo: func(r *fakeRepository) {
				r.shouldFailCreate = true
			},
			setupSigner: func(s *fakeCryptoSigner) {},
			expectError: true,
		},
		{
			name: "user without name",
			request: &models.SignatureRequest{
				DocID: "test-doc",
				User: &models.User{
					Sub:   "user-123",
					Email: "test@example.com",
					Name:  "",
				},
			},
			setupRepo:   func(r *fakeRepository) {},
			setupSigner: func(s *fakeCryptoSigner) {},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			signer := newFakeCryptoSigner()

			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			if tt.setupSigner != nil {
				tt.setupSigner(signer)
			}

			service := NewSignatureService(repo, signer)

			err := service.CreateSignature(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
					t.Errorf("Error = %v, expected %v", err, tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify signature was created
			key := tt.request.DocID + "_" + tt.request.User.Sub
			signature, exists := repo.signatures[key]
			if !exists {
				t.Error("Signature should have been created")
				return
			}

			if signature.DocID != tt.request.DocID {
				t.Errorf("DocID = %v, expected %v", signature.DocID, tt.request.DocID)
			}
			if signature.UserSub != tt.request.User.Sub {
				t.Errorf("UserSub = %v, expected %v", signature.UserSub, tt.request.User.Sub)
			}
			if signature.UserEmail != tt.request.User.NormalizedEmail() {
				t.Errorf("UserEmail = %v, expected %v", signature.UserEmail, tt.request.User.NormalizedEmail())
			}
		})
	}
}

func TestSignatureService_GetSignatureStatus(t *testing.T) {
	tests := []struct {
		name           string
		docID          string
		user           *models.User
		setupRepo      func(*fakeRepository)
		expectError    bool
		expectedError  error
		expectedSigned bool
	}{
		{
			name:  "user has signed",
			docID: "test-doc",
			user: &models.User{
				Sub:   "user-123",
				Email: "test@example.com",
			},
			setupRepo: func(r *fakeRepository) {
				r.signatures["test-doc_user-123"] = &models.Signature{
					ID:          1,
					DocID:       "test-doc",
					UserSub:     "user-123",
					SignedAtUTC: time.Now().UTC(),
				}
			},
			expectError:    false,
			expectedSigned: true,
		},
		{
			name:  "user has not signed",
			docID: "test-doc",
			user: &models.User{
				Sub:   "user-123",
				Email: "test@example.com",
			},
			setupRepo:      func(r *fakeRepository) {},
			expectError:    false,
			expectedSigned: false,
		},
		{
			name:          "invalid user - nil",
			docID:         "test-doc",
			user:          nil,
			expectError:   true,
			expectedError: models.ErrInvalidUser,
		},
		{
			name:  "invalid user - invalid data",
			docID: "test-doc",
			user: &models.User{
				Sub:   "",
				Email: "test@example.com",
			},
			expectError:   true,
			expectedError: models.ErrInvalidUser,
		},
		{
			name:  "repository get fails",
			docID: "test-doc",
			user: &models.User{
				Sub:   "user-123",
				Email: "test@example.com",
			},
			setupRepo: func(r *fakeRepository) {
				r.shouldFailGet = true
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			signer := newFakeCryptoSigner()

			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			service := NewSignatureService(repo, signer)

			status, err := service.GetSignatureStatus(context.Background(), tt.docID, tt.user)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
					t.Errorf("Error = %v, expected %v", err, tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if status == nil {
				t.Error("Status should not be nil")
				return
			}

			if status.DocID != tt.docID {
				t.Errorf("Status.DocID = %v, expected %v", status.DocID, tt.docID)
			}
			if status.UserEmail != tt.user.Email {
				t.Errorf("Status.UserEmail = %v, expected %v", status.UserEmail, tt.user.Email)
			}
			if status.IsSigned != tt.expectedSigned {
				t.Errorf("Status.IsSigned = %v, expected %v", status.IsSigned, tt.expectedSigned)
			}
		})
	}
}

func TestSignatureService_GetDocumentSignatures(t *testing.T) {
	repo := newFakeRepository()
	signer := newFakeCryptoSigner()
	service := NewSignatureService(repo, signer)

	// Setup test data
	sig1 := &models.Signature{ID: 1, DocID: "doc1", UserSub: "user1"}
	sig2 := &models.Signature{ID: 2, DocID: "doc1", UserSub: "user2"}
	sig3 := &models.Signature{ID: 3, DocID: "doc2", UserSub: "user1"}

	repo.signatures["doc1_user1"] = sig1
	repo.signatures["doc1_user2"] = sig2
	repo.signatures["doc2_user1"] = sig3

	t.Run("get signatures for document", func(t *testing.T) {
		signatures, err := service.GetDocumentSignatures(context.Background(), "doc1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(signatures) != 2 {
			t.Errorf("Expected 2 signatures, got %d", len(signatures))
		}
	})

	t.Run("repository fails", func(t *testing.T) {
		repo.shouldFailGet = true
		_, err := service.GetDocumentSignatures(context.Background(), "doc1")
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestSignatureService_GetUserSignatures(t *testing.T) {
	repo := newFakeRepository()
	signer := newFakeCryptoSigner()
	service := NewSignatureService(repo, signer)

	// Setup test data
	sig1 := &models.Signature{ID: 1, DocID: "doc1", UserSub: "user1"}
	sig2 := &models.Signature{ID: 2, DocID: "doc2", UserSub: "user1"}
	sig3 := &models.Signature{ID: 3, DocID: "doc1", UserSub: "user2"}

	repo.signatures["doc1_user1"] = sig1
	repo.signatures["doc2_user1"] = sig2
	repo.signatures["doc1_user2"] = sig3

	t.Run("get signatures for user", func(t *testing.T) {
		user := &models.User{Sub: "user1", Email: "user1@example.com"}
		signatures, err := service.GetUserSignatures(context.Background(), user)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(signatures) != 2 {
			t.Errorf("Expected 2 signatures, got %d", len(signatures))
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		_, err := service.GetUserSignatures(context.Background(), nil)
		if err != models.ErrInvalidUser {
			t.Errorf("Error = %v, expected %v", err, models.ErrInvalidUser)
		}
	})

	t.Run("repository fails", func(t *testing.T) {
		user := &models.User{Sub: "user1", Email: "user1@example.com"}
		repo.shouldFailGet = true
		_, err := service.GetUserSignatures(context.Background(), user)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestSignatureService_GetSignatureByDocAndUser(t *testing.T) {
	repo := newFakeRepository()
	signer := newFakeCryptoSigner()
	service := NewSignatureService(repo, signer)

	// Setup test data
	sig := &models.Signature{ID: 1, DocID: "doc1", UserSub: "user1"}
	repo.signatures["doc1_user1"] = sig

	t.Run("get existing signature", func(t *testing.T) {
		user := &models.User{Sub: "user1", Email: "user1@example.com"}
		signature, err := service.GetSignatureByDocAndUser(context.Background(), "doc1", user)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if signature.ID != 1 {
			t.Errorf("Signature.ID = %v, expected 1", signature.ID)
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		_, err := service.GetSignatureByDocAndUser(context.Background(), "doc1", nil)
		if err != models.ErrInvalidUser {
			t.Errorf("Error = %v, expected %v", err, models.ErrInvalidUser)
		}
	})

	t.Run("repository fails", func(t *testing.T) {
		user := &models.User{Sub: "user1", Email: "user1@example.com"}
		repo.shouldFailGet = true
		_, err := service.GetSignatureByDocAndUser(context.Background(), "doc1", user)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestSignatureService_CheckUserSignature(t *testing.T) {
	repo := newFakeRepository()
	signer := newFakeCryptoSigner()
	service := NewSignatureService(repo, signer)

	// Setup test data
	sig := &models.Signature{ID: 1, DocID: "doc1", UserSub: "user1", UserEmail: "user1@example.com"}
	repo.signatures["doc1_user1"] = sig

	t.Run("check by user sub", func(t *testing.T) {
		exists, err := service.CheckUserSignature(context.Background(), "doc1", "user1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !exists {
			t.Error("Should find signature by user sub")
		}
	})

	t.Run("check by email", func(t *testing.T) {
		exists, err := service.CheckUserSignature(context.Background(), "doc1", "user1@example.com")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !exists {
			t.Error("Should find signature by email")
		}
	})

	t.Run("signature not found", func(t *testing.T) {
		exists, err := service.CheckUserSignature(context.Background(), "doc1", "nonexistent")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if exists {
			t.Error("Should not find nonexistent signature")
		}
	})

	t.Run("repository fails", func(t *testing.T) {
		repo.shouldFailCheck = true
		_, err := service.CheckUserSignature(context.Background(), "doc1", "user1")
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestSignatureService_VerifyChainIntegrity(t *testing.T) {
	tests := []struct {
		name            string
		setupSignatures func(*fakeRepository)
		expectValid     bool
		expectBreakAtID *int64
		expectDetails   string
	}{
		{
			name:            "empty chain",
			setupSignatures: func(r *fakeRepository) {},
			expectValid:     true,
			expectDetails:   "No signatures found",
		},
		{
			name: "valid chain with single signature",
			setupSignatures: func(r *fakeRepository) {
				sig1 := &models.Signature{
					ID:       1,
					DocID:    "doc1",
					UserSub:  "user1",
					PrevHash: nil, // Genesis
				}
				r.allSignatures = []*models.Signature{sig1}
			},
			expectValid:   true,
			expectDetails: "Chain integrity verified successfully",
		},
		{
			name: "valid chain with multiple signatures",
			setupSignatures: func(r *fakeRepository) {
				sig1 := &models.Signature{
					ID:       1,
					DocID:    "doc1",
					UserSub:  "user1",
					PrevHash: nil, // Genesis
				}
				hash1 := sig1.ComputeRecordHash()
				sig2 := &models.Signature{
					ID:       2,
					DocID:    "doc2",
					UserSub:  "user2",
					PrevHash: &hash1,
				}
				r.allSignatures = []*models.Signature{sig1, sig2}
			},
			expectValid:   true,
			expectDetails: "Chain integrity verified successfully",
		},
		{
			name: "invalid chain - genesis has prev hash",
			setupSignatures: func(r *fakeRepository) {
				hash := "invalid-genesis-hash"
				sig1 := &models.Signature{
					ID:       1,
					DocID:    "doc1",
					UserSub:  "user1",
					PrevHash: &hash,
				}
				r.allSignatures = []*models.Signature{sig1}
			},
			expectValid:     false,
			expectBreakAtID: int64Ptr(1),
			expectDetails:   "Genesis signature has non-null previous hash",
		},
		{
			name: "invalid chain - missing prev hash",
			setupSignatures: func(r *fakeRepository) {
				sig1 := &models.Signature{
					ID:       1,
					DocID:    "doc1",
					UserSub:  "user1",
					PrevHash: nil, // Genesis
				}
				sig2 := &models.Signature{
					ID:       2,
					DocID:    "doc2",
					UserSub:  "user2",
					PrevHash: nil, // Should have prev hash
				}
				r.allSignatures = []*models.Signature{sig1, sig2}
			},
			expectValid:     false,
			expectBreakAtID: int64Ptr(2),
		},
		{
			name: "invalid chain - wrong prev hash",
			setupSignatures: func(r *fakeRepository) {
				sig1 := &models.Signature{
					ID:       1,
					DocID:    "doc1",
					UserSub:  "user1",
					PrevHash: nil, // Genesis
				}
				wrongHash := "wrong-hash-that-is-long-enough-for-display"
				sig2 := &models.Signature{
					ID:       2,
					DocID:    "doc2",
					UserSub:  "user2",
					PrevHash: &wrongHash,
				}
				r.allSignatures = []*models.Signature{sig1, sig2}
			},
			expectValid:     false,
			expectBreakAtID: int64Ptr(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			signer := newFakeCryptoSigner()
			service := NewSignatureService(repo, signer)

			tt.setupSignatures(repo)

			result, err := service.VerifyChainIntegrity(context.Background())
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.IsValid != tt.expectValid {
				t.Errorf("IsValid = %v, expected %v", result.IsValid, tt.expectValid)
			}

			if tt.expectBreakAtID != nil {
				if result.BreakAtID == nil {
					t.Error("Expected BreakAtID to be set")
				} else if *result.BreakAtID != *tt.expectBreakAtID {
					t.Errorf("BreakAtID = %v, expected %v", *result.BreakAtID, *tt.expectBreakAtID)
				}
			}

			if tt.expectDetails != "" && !contains(result.Details, tt.expectDetails) {
				t.Errorf("Details should contain %v, got %v", tt.expectDetails, result.Details)
			}
		})
	}

	t.Run("repository fails", func(t *testing.T) {
		repo := newFakeRepository()
		signer := newFakeCryptoSigner()
		service := NewSignatureService(repo, signer)

		repo.shouldFailGetAll = true

		_, err := service.VerifyChainIntegrity(context.Background())
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestSignatureService_RebuildChain(t *testing.T) {
	t.Run("empty chain", func(t *testing.T) {
		repo := newFakeRepository()
		signer := newFakeCryptoSigner()
		service := NewSignatureService(repo, signer)

		err := service.RebuildChain(context.Background())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("chain with signatures", func(t *testing.T) {
		repo := newFakeRepository()
		signer := newFakeCryptoSigner()
		service := NewSignatureService(repo, signer)

		// Setup signatures that need rebuilding
		hash := "wrong-hash"
		sig1 := &models.Signature{
			ID:       1,
			DocID:    "doc1",
			UserSub:  "user1",
			PrevHash: &hash, // Should be nil for genesis
		}
		sig2 := &models.Signature{
			ID:       2,
			DocID:    "doc2",
			UserSub:  "user2",
			PrevHash: nil, // Should have correct hash
		}
		repo.allSignatures = []*models.Signature{sig1, sig2}

		err := service.RebuildChain(context.Background())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("repository fails", func(t *testing.T) {
		repo := newFakeRepository()
		signer := newFakeCryptoSigner()
		service := NewSignatureService(repo, signer)

		repo.shouldFailGetAll = true

		err := service.RebuildChain(context.Background())
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestChainIntegrityResult_Structure(t *testing.T) {
	result := &ChainIntegrityResult{
		IsValid:      true,
		TotalRecords: 5,
		BreakAtID:    int64Ptr(3),
		Details:      "Test details",
	}

	if !result.IsValid {
		t.Error("IsValid should be true")
	}
	if result.TotalRecords != 5 {
		t.Errorf("TotalRecords = %v, expected 5", result.TotalRecords)
	}
	if result.BreakAtID == nil || *result.BreakAtID != 3 {
		t.Errorf("BreakAtID = %v, expected 3", result.BreakAtID)
	}
	if result.Details != "Test details" {
		t.Errorf("Details = %v, expected 'Test details'", result.Details)
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if s[start+i] != substr[i] {
			if start+1 < len(s) {
				return containsAt(s, substr, start+1)
			}
			return false
		}
	}
	return true
}
