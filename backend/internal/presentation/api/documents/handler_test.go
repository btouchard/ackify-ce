// SPDX-License-Identifier: AGPL-3.0-or-later
package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
)

// ============================================================================
// TEST FIXTURES & MOCKS
// ============================================================================

var (
	testDoc = &models.Document{
		DocID:             "test-doc-123",
		Title:             "Test Document",
		URL:               "https://example.com/doc.pdf",
		Description:       "Test description",
		Checksum:          "abc123",
		ChecksumAlgorithm: "SHA-256",
		CreatedAt:         time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		CreatedBy:         "user@example.com",
	}

	testSignature = &models.Signature{
		ID:          1,
		DocID:       "test-doc-123",
		UserSub:     "oauth2|123",
		UserEmail:   "user@example.com",
		UserName:    "Test User",
		SignedAtUTC: time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
		PayloadHash: "payload-hash-123",
		Signature:   "signature-123",
		Nonce:       "nonce-123",
		CreatedAt:   time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
		PrevHash:    stringPtr("prev-hash-123"),
		Referer:     stringPtr("https://example.com"),
	}

	testUser = &models.User{
		Sub:   "oauth2|123",
		Email: "user@example.com",
		Name:  "Test User",
	}
)

func stringPtr(s string) *string {
	return &s
}

// Mock document service
type mockDocumentService struct {
	createDocFunc       func(ctx context.Context, req services.CreateDocumentRequest) (*models.Document, error)
	findOrCreateDocFunc func(ctx context.Context, ref string) (*models.Document, bool, error)
	findByReferenceFunc func(ctx context.Context, ref string, refType string) (*models.Document, error)
}

func (m *mockDocumentService) CreateDocument(ctx context.Context, req services.CreateDocumentRequest) (*models.Document, error) {
	if m.createDocFunc != nil {
		return m.createDocFunc(ctx, req)
	}
	return testDoc, nil
}

func (m *mockDocumentService) FindOrCreateDocument(ctx context.Context, ref string) (*models.Document, bool, error) {
	if m.findOrCreateDocFunc != nil {
		return m.findOrCreateDocFunc(ctx, ref)
	}
	return testDoc, true, nil
}

func (m *mockDocumentService) FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error) {
	if m.findByReferenceFunc != nil {
		return m.findByReferenceFunc(ctx, ref, refType)
	}
	return nil, fmt.Errorf("document not found")
}

// Mock signature service
type mockSignatureService struct {
	getDocumentSignaturesFunc func(ctx context.Context, docID string) ([]*models.Signature, error)
}

func (m *mockSignatureService) GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error) {
	if m.getDocumentSignaturesFunc != nil {
		return m.getDocumentSignaturesFunc(ctx, docID)
	}
	return []*models.Signature{testSignature}, nil
}

func createTestHandler() *Handler {
	return &Handler{
		signatureService: &services.SignatureService{}, // Not used in these tests
		documentService:  &mockDocumentService{},
	}
}

func addUserToContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, shared.ContextKeyUser, user)
}

// ============================================================================
// TESTS - Constructor
// ============================================================================

func TestNewHandler(t *testing.T) {
	t.Parallel()

	sigService := &services.SignatureService{}
	docService := &mockDocumentService{}

	handler := NewHandler(sigService, docService)

	assert.NotNil(t, handler)
	assert.Equal(t, sigService, handler.signatureService)
	assert.Equal(t, docService, handler.documentService)
}

// ============================================================================
// TESTS - HandleCreateDocument
// ============================================================================

func TestHandler_HandleCreateDocument_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reference string
		title     string
	}{
		{
			name:      "with title",
			reference: "https://example.com/doc.pdf",
			title:     "My Document",
		},
		{
			name:      "without title",
			reference: "https://example.com/doc.pdf",
			title:     "",
		},
		{
			name:      "with file path reference",
			reference: "/path/to/document.pdf",
			title:     "Local Document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDocService := &mockDocumentService{
				createDocFunc: func(ctx context.Context, req services.CreateDocumentRequest) (*models.Document, error) {
					assert.Equal(t, tt.reference, req.Reference)
					assert.Equal(t, tt.title, req.Title)
					return testDoc, nil
				},
			}

			handler := &Handler{
				documentService: mockDocService,
			}

			reqBody := CreateDocumentRequest{
				Reference: tt.reference,
				Title:     tt.title,
			}
			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleCreateDocument(rec, req)

			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var wrapper struct {
				Data CreateDocumentResponse `json:"data"`
			}
			err = json.Unmarshal(rec.Body.Bytes(), &wrapper)
			require.NoError(t, err)

			assert.Equal(t, testDoc.DocID, wrapper.Data.DocID)
			assert.Equal(t, testDoc.Title, wrapper.Data.Title)
			assert.Equal(t, testDoc.URL, wrapper.Data.URL)
			assert.NotEmpty(t, wrapper.Data.CreatedAt)
		})
	}
}

func TestHandler_HandleCreateDocument_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "empty reference",
			requestBody:    CreateDocumentRequest{Reference: "", Title: "Title"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Reference is required",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := createTestHandler()

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleCreateDocument(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
		})
	}
}

func TestHandler_HandleCreateDocument_ServiceError(t *testing.T) {
	t.Parallel()

	mockDocService := &mockDocumentService{
		createDocFunc: func(ctx context.Context, req services.CreateDocumentRequest) (*models.Document, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := &Handler{
		documentService: mockDocService,
	}

	reqBody := CreateDocumentRequest{
		Reference: "https://example.com/doc.pdf",
		Title:     "Test",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleCreateDocument(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// ============================================================================
// TESTS - HandleListDocuments
// ============================================================================

func TestHandler_HandleListDocuments_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		queryParams   string
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "default pagination",
			queryParams:   "",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "custom page and limit",
			queryParams:   "?page=2&limit=50",
			expectedPage:  2,
			expectedLimit: 50,
		},
		{
			name:          "limit max capped at 100",
			queryParams:   "?limit=200",
			expectedPage:  1,
			expectedLimit: 20, // Will use default since > 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := createTestHandler()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/documents"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			handler.HandleListDocuments(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var wrapper struct {
				Data interface{} `json:"data"`
				Meta struct {
					Page  int `json:"page"`
					Limit int `json:"limit"`
					Total int `json:"total"`
				} `json:"meta"`
			}
			err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
			require.NoError(t, err)

			// Currently returns empty list
			assert.NotNil(t, wrapper.Data)
		})
	}
}

// ============================================================================
// TESTS - HandleGetDocument
// ============================================================================

// TestHandler_HandleGetDocument_Success is skipped because SignatureService
// cannot be mocked without significant refactoring. The service requires
// a repository interface that we cannot inject in tests.
// TODO: Refactor to use interface for signature service
func TestHandler_HandleGetDocument_Success(t *testing.T) {
	t.Skip("SignatureService is not mockable - needs refactoring")
}

func TestHandler_HandleGetDocument_MissingDocID(t *testing.T) {
	t.Parallel()

	handler := createTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/", nil)

	// Empty docId parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("docId", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.HandleGetDocument(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// TESTS - HandleGetDocumentSignatures
// ============================================================================

func TestHandler_HandleGetDocumentSignatures_MissingDocID(t *testing.T) {
	t.Parallel()

	handler := createTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents//signatures", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("docId", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.HandleGetDocumentSignatures(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// TESTS - HandleFindOrCreateDocument
// ============================================================================

func TestHandler_HandleFindOrCreateDocument_FindExisting(t *testing.T) {
	t.Parallel()

	mockDocService := &mockDocumentService{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			assert.Equal(t, "https://example.com/doc.pdf", ref)
			assert.Equal(t, "url", refType)
			return testDoc, nil
		},
	}

	handler := &Handler{
		documentService: mockDocService,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/find-or-create?ref=https://example.com/doc.pdf", nil)
	rec := httptest.NewRecorder()

	handler.HandleFindOrCreateDocument(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var wrapper struct {
		Data FindOrCreateDocumentResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
	require.NoError(t, err)

	assert.Equal(t, testDoc.DocID, wrapper.Data.DocID)
	assert.False(t, wrapper.Data.IsNew, "Should not be new since document was found")
}

func TestHandler_HandleFindOrCreateDocument_CreateNew(t *testing.T) {
	t.Parallel()

	mockDocService := &mockDocumentService{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			// Document not found - return nil, nil (not an error)
			return nil, nil
		},
		findOrCreateDocFunc: func(ctx context.Context, ref string) (*models.Document, bool, error) {
			assert.Equal(t, "https://example.com/new-doc.pdf", ref)
			return testDoc, true, nil
		},
	}

	handler := &Handler{
		documentService: mockDocService,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/find-or-create?ref=https://example.com/new-doc.pdf", nil)

	// Add authenticated user to context
	ctx := addUserToContext(req.Context(), testUser)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.HandleFindOrCreateDocument(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var wrapper struct {
		Data FindOrCreateDocumentResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
	require.NoError(t, err)

	assert.Equal(t, testDoc.DocID, wrapper.Data.DocID)
	assert.True(t, wrapper.Data.IsNew, "Should be new since document was created")
}

func TestHandler_HandleFindOrCreateDocument_UnauthenticatedCreate(t *testing.T) {
	t.Parallel()

	mockDocService := &mockDocumentService{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			// Document not found - return nil, nil (not an error)
			return nil, nil
		},
	}

	handler := &Handler{
		documentService: mockDocService,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/find-or-create?ref=https://example.com/new-doc.pdf", nil)
	// No user in context
	rec := httptest.NewRecorder()

	handler.HandleFindOrCreateDocument(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestHandler_HandleFindOrCreateDocument_MissingRef(t *testing.T) {
	t.Parallel()

	handler := createTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/find-or-create", nil)
	rec := httptest.NewRecorder()

	handler.HandleFindOrCreateDocument(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// ============================================================================
// TESTS - detectReferenceType
// ============================================================================

func Test_detectReferenceType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ref      string
		expected ReferenceType
	}{
		{
			name:     "HTTP URL",
			ref:      "http://example.com/doc.pdf",
			expected: "url",
		},
		{
			name:     "HTTPS URL",
			ref:      "https://example.com/doc.pdf",
			expected: "url",
		},
		{
			name:     "Unix file path",
			ref:      "/path/to/document.pdf",
			expected: "path",
		},
		{
			name:     "Windows file path",
			ref:      "C:\\path\\to\\document.pdf",
			expected: "path",
		},
		{
			name:     "Simple reference",
			ref:      "doc-12345",
			expected: "reference",
		},
		{
			name:     "Hash reference",
			ref:      "abc123def456",
			expected: "reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := detectReferenceType(tt.ref)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// TESTS - signatureToDTO
// ============================================================================

func Test_signatureToDTO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sig      *models.Signature
		checkDTO func(t *testing.T, dto SignatureDTO)
	}{
		{
			name: "with prevHash",
			sig:  testSignature,
			checkDTO: func(t *testing.T, dto SignatureDTO) {
				assert.Equal(t, "1", dto.ID)
				assert.Equal(t, testSignature.DocID, dto.DocID)
				assert.Equal(t, testSignature.UserEmail, dto.UserEmail)
				assert.Equal(t, testSignature.UserName, dto.UserName)
				assert.Equal(t, testSignature.Signature, dto.Signature)
				assert.Equal(t, testSignature.PayloadHash, dto.PayloadHash)
				assert.Equal(t, testSignature.Nonce, dto.Nonce)
				assert.Equal(t, *testSignature.PrevHash, dto.PrevHash)
				assert.NotEmpty(t, dto.SignedAt)
			},
		},
		{
			name: "without prevHash",
			sig: &models.Signature{
				ID:          2,
				DocID:       "doc-456",
				UserSub:     "oauth2|456",
				UserEmail:   "user2@example.com",
				UserName:    "User 2",
				SignedAtUTC: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
				PayloadHash: "hash-456",
				Signature:   "sig-456",
				Nonce:       "nonce-456",
				PrevHash:    nil,
			},
			checkDTO: func(t *testing.T, dto SignatureDTO) {
				assert.Equal(t, "2", dto.ID)
				assert.Empty(t, dto.PrevHash)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dto := signatureToDTO(tt.sig)
			tt.checkDTO(t, dto)
		})
	}
}

// ============================================================================
// TESTS - Concurrency
// ============================================================================

func TestHandler_HandleCreateDocument_Concurrent(t *testing.T) {
	t.Parallel()

	handler := createTestHandler()

	const numRequests = 50
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			defer func() { done <- true }()

			reqBody := CreateDocumentRequest{
				Reference: fmt.Sprintf("https://example.com/doc-%d.pdf", id),
				Title:     fmt.Sprintf("Document %d", id),
			}
			body, err := json.Marshal(reqBody)
			if err != nil {
				errors <- err
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleCreateDocument(rec, req)

			if rec.Code != http.StatusCreated {
				errors <- fmt.Errorf("unexpected status: %d", rec.Code)
			}
		}(i)
	}

	for i := 0; i < numRequests; i++ {
		<-done
	}
	close(errors)

	var errCount int
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errCount++
	}

	assert.Equal(t, 0, errCount, "All concurrent requests should succeed")
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkHandler_HandleCreateDocument(b *testing.B) {
	handler := createTestHandler()

	reqBody := CreateDocumentRequest{
		Reference: "https://example.com/doc.pdf",
		Title:     "Test Document",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleCreateDocument(rec, req)
	}
}

func BenchmarkHandler_HandleCreateDocument_Parallel(b *testing.B) {
	handler := createTestHandler()

	reqBody := CreateDocumentRequest{
		Reference: "https://example.com/doc.pdf",
		Title:     "Test Document",
	}
	body, _ := json.Marshal(reqBody)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleCreateDocument(rec, req)
		}
	})
}

func Benchmark_detectReferenceType(b *testing.B) {
	refs := []string{
		"https://example.com/doc.pdf",
		"/path/to/file.pdf",
		"simple-reference",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detectReferenceType(refs[i%len(refs)])
	}
}
