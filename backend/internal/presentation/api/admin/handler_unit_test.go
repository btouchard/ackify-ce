// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MOCKS
// ============================================================================

type mockDocumentRepository struct {
	getByDocIDFunc     func(ctx context.Context, docID string) (*models.Document, error)
	listFunc           func(ctx context.Context, limit, offset int) ([]*models.Document, error)
	createOrUpdateFunc func(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error)
	deleteFunc         func(ctx context.Context, docID string) error
}

func (m *mockDocumentRepository) GetByDocID(ctx context.Context, docID string) (*models.Document, error) {
	if m.getByDocIDFunc != nil {
		return m.getByDocIDFunc(ctx, docID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDocumentRepository) List(ctx context.Context, limit, offset int) ([]*models.Document, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDocumentRepository) CreateOrUpdate(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	if m.createOrUpdateFunc != nil {
		return m.createOrUpdateFunc(ctx, docID, input, createdBy)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDocumentRepository) Delete(ctx context.Context, docID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, docID)
	}
	return errors.New("not implemented")
}

type mockExpectedSignerRepository struct {
	listByDocIDFunc           func(ctx context.Context, docID string) ([]*models.ExpectedSigner, error)
	listWithStatusByDocIDFunc func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error)
	addExpectedFunc           func(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error
	removeFunc                func(ctx context.Context, docID, email string) error
	getStatsFunc              func(ctx context.Context, docID string) (*models.DocCompletionStats, error)
}

func (m *mockExpectedSignerRepository) ListByDocID(ctx context.Context, docID string) ([]*models.ExpectedSigner, error) {
	if m.listByDocIDFunc != nil {
		return m.listByDocIDFunc(ctx, docID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExpectedSignerRepository) ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
	if m.listWithStatusByDocIDFunc != nil {
		return m.listWithStatusByDocIDFunc(ctx, docID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExpectedSignerRepository) AddExpected(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error {
	if m.addExpectedFunc != nil {
		return m.addExpectedFunc(ctx, docID, contacts, addedBy)
	}
	return errors.New("not implemented")
}

func (m *mockExpectedSignerRepository) Remove(ctx context.Context, docID, email string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, docID, email)
	}
	return errors.New("not implemented")
}

func (m *mockExpectedSignerRepository) GetStats(ctx context.Context, docID string) (*models.DocCompletionStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, docID)
	}
	return nil, errors.New("not implemented")
}

type mockReminderService struct {
	sendRemindersFunc      func(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string, locale string) (*models.ReminderSendResult, error)
	getReminderHistoryFunc func(ctx context.Context, docID string) ([]*models.ReminderLog, error)
	getReminderStatsFunc   func(ctx context.Context, docID string) (*models.ReminderStats, error)
}

func (m *mockReminderService) SendReminders(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string, locale string) (*models.ReminderSendResult, error) {
	if m.sendRemindersFunc != nil {
		return m.sendRemindersFunc(ctx, docID, sentBy, specificEmails, docURL, locale)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReminderService) GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
	if m.getReminderHistoryFunc != nil {
		return m.getReminderHistoryFunc(ctx, docID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReminderService) GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error) {
	if m.getReminderStatsFunc != nil {
		return m.getReminderStatsFunc(ctx, docID)
	}
	return nil, errors.New("not implemented")
}

type mockSignatureService struct {
	getDocumentSignaturesFunc func(ctx context.Context, docID string) ([]*models.Signature, error)
}

func (m *mockSignatureService) GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error) {
	if m.getDocumentSignaturesFunc != nil {
		return m.getDocumentSignaturesFunc(ctx, docID)
	}
	return nil, errors.New("not implemented")
}

// ============================================================================
// HELPERS
// ============================================================================

func createTestHandler(docRepo documentRepository, signerRepo expectedSignerRepository, reminderSvc reminderService, sigService signatureService) *Handler {
	return NewHandler(docRepo, signerRepo, reminderSvc, sigService, "https://test.example.com")
}

func createContextWithUser(email string, isAdmin bool) context.Context {
	user := &models.User{
		Sub:   "test-sub-123",
		Email: email,
		Name:  "Test User",
	}
	return context.WithValue(context.Background(), shared.ContextKeyUser, user)
}

func createTestDocument(docID string) *models.Document {
	now := time.Now()
	return &models.Document{
		DocID:             docID,
		Title:             "Test Document",
		URL:               "https://example.com/doc.pdf",
		Checksum:          "abc123",
		ChecksumAlgorithm: "SHA-256",
		Description:       "Test description",
		CreatedAt:         now,
		UpdatedAt:         now,
		CreatedBy:         "admin@example.com",
	}
}

func createTestExpectedSignerWithStatus(docID, email string, hasSigned bool) *models.ExpectedSignerWithStatus {
	now := time.Now()
	status := &models.ExpectedSignerWithStatus{
		ExpectedSigner: models.ExpectedSigner{
			ID:      1,
			DocID:   docID,
			Email:   email,
			Name:    "Test Signer",
			AddedAt: now,
			AddedBy: "admin@example.com",
		},
		HasSigned:             hasSigned,
		ReminderCount:         0,
		DaysSinceAdded:        5,
		DaysSinceLastReminder: nil,
	}
	if hasSigned {
		signedAt := now.Add(-2 * time.Hour)
		status.SignedAt = &signedAt
		userName := "Test Signer"
		status.UserName = &userName
	}
	return status
}

func createTestReminderLog(docID, email string) *models.ReminderLog {
	return &models.ReminderLog{
		ID:             1,
		DocID:          docID,
		RecipientEmail: email,
		SentAt:         time.Now(),
		SentBy:         "admin@example.com",
		TemplateUsed:   "reminder",
		Status:         "sent",
	}
}

// ============================================================================
// TESTS - HandleListDocuments
// ============================================================================

func TestHandleListDocuments_Success(t *testing.T) {
	t.Parallel()

	docs := []*models.Document{
		createTestDocument("doc1"),
		createTestDocument("doc2"),
	}

	docRepo := &mockDocumentRepository{
		listFunc: func(ctx context.Context, limit, offset int) ([]*models.Document, error) {
			assert.Equal(t, 100, limit)
			assert.Equal(t, 0, offset)
			return docs, nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents", nil)
	rec := httptest.NewRecorder()

	handler.HandleListDocuments(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data []DocumentResponse     `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, 2, int(response.Meta["total"].(float64)))
}

func TestHandleListDocuments_EmptyList(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		listFunc: func(ctx context.Context, limit, offset int) ([]*models.Document, error) {
			return []*models.Document{}, nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents", nil)
	rec := httptest.NewRecorder()

	handler.HandleListDocuments(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data []DocumentResponse     `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 0)
}

func TestHandleListDocuments_RepositoryError(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		listFunc: func(ctx context.Context, limit, offset int) ([]*models.Document, error) {
			return nil, errors.New("database error")
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents", nil)
	rec := httptest.NewRecorder()

	handler.HandleListDocuments(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================================
// TESTS - HandleGetDocument
// ============================================================================

func TestHandleGetDocument_Success(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			assert.Equal(t, "doc1", docID)
			return doc, nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}", handler.HandleGetDocument)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data DocumentResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "doc1", response.Data.DocID)
	assert.Equal(t, "Test Document", response.Data.Title)
}

func TestHandleGetDocument_NotFound(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}", handler.HandleGetDocument)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetDocument_EmptyDocID(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	// Without chi routing context, docId will be empty
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/", nil)
	rec := httptest.NewRecorder()

	handler.HandleGetDocument(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// TESTS - HandleGetDocumentWithSigners
// ============================================================================

func TestHandleGetDocumentWithSigners_Success(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	signers := []*models.ExpectedSignerWithStatus{
		createTestExpectedSignerWithStatus("doc1", "signer1@example.com", true),
		createTestExpectedSignerWithStatus("doc1", "signer2@example.com", false),
	}
	stats := &models.DocCompletionStats{
		DocID:          "doc1",
		ExpectedCount:  2,
		SignedCount:    1,
		PendingCount:   1,
		CompletionRate: 50.0,
	}

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
	}
	signerRepo := &mockExpectedSignerRepository{
		listWithStatusByDocIDFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return signers, nil
		},
		getStatsFunc: func(ctx context.Context, docID string) (*models.DocCompletionStats, error) {
			return stats, nil
		},
	}

	handler := createTestHandler(docRepo, signerRepo, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/signers", handler.HandleGetDocumentWithSigners)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/signers", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Data["document"])
	assert.NotNil(t, response.Data["signers"])
	assert.NotNil(t, response.Data["stats"])
}

func TestHandleGetDocumentWithSigners_DocumentNotFound(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/signers", handler.HandleGetDocumentWithSigners)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/nonexistent/signers", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetDocumentWithSigners_SignersError(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
	}
	signerRepo := &mockExpectedSignerRepository{
		listWithStatusByDocIDFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return nil, errors.New("database error")
		},
	}

	handler := createTestHandler(docRepo, signerRepo, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/signers", handler.HandleGetDocumentWithSigners)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/signers", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================================
// TESTS - HandleAddExpectedSigner
// ============================================================================

func TestHandleAddExpectedSigner_Success(t *testing.T) {
	t.Parallel()

	signerRepo := &mockExpectedSignerRepository{
		addExpectedFunc: func(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error {
			assert.Equal(t, "doc1", docID)
			assert.Len(t, contacts, 1)
			assert.Equal(t, "new@example.com", contacts[0].Email)
			assert.Equal(t, "admin@example.com", addedBy)
			return nil
		},
	}

	handler := createTestHandler(nil, signerRepo, nil, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/signers", handler.HandleAddExpectedSigner)

	reqBody := AddExpectedSignerRequest{
		Email: "new@example.com",
		Name:  "New Signer",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/signers", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response struct {
		Data map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", response.Data["email"])
}

func TestHandleAddExpectedSigner_MissingEmail(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/signers", handler.HandleAddExpectedSigner)

	reqBody := AddExpectedSignerRequest{
		Email: "",
		Name:  "New Signer",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/signers", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleAddExpectedSigner_NoUser(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/signers", handler.HandleAddExpectedSigner)

	reqBody := AddExpectedSignerRequest{
		Email: "new@example.com",
		Name:  "New Signer",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/signers", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleAddExpectedSigner_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/signers", handler.HandleAddExpectedSigner)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/signers", strings.NewReader("invalid json"))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// TESTS - HandleRemoveExpectedSigner
// ============================================================================

func TestHandleRemoveExpectedSigner_Success(t *testing.T) {
	t.Parallel()

	signerRepo := &mockExpectedSignerRepository{
		removeFunc: func(ctx context.Context, docID, email string) error {
			assert.Equal(t, "doc1", docID)
			assert.Equal(t, "remove@example.com", email)
			return nil
		},
	}

	handler := createTestHandler(nil, signerRepo, nil, nil)

	router := chi.NewRouter()
	router.Delete("/api/v1/admin/documents/{docId}/signers/{email}", handler.HandleRemoveExpectedSigner)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/documents/doc1/signers/remove@example.com", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleRemoveExpectedSigner_RepositoryError(t *testing.T) {
	t.Parallel()

	signerRepo := &mockExpectedSignerRepository{
		removeFunc: func(ctx context.Context, docID, email string) error {
			return errors.New("database error")
		},
	}

	handler := createTestHandler(nil, signerRepo, nil, nil)

	router := chi.NewRouter()
	router.Delete("/api/v1/admin/documents/{docId}/signers/{email}", handler.HandleRemoveExpectedSigner)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/documents/doc1/signers/remove@example.com", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleRemoveExpectedSigner_EmptyParams(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	// Without chi routing context, params will be empty
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/documents//signers/", nil)
	rec := httptest.NewRecorder()

	handler.HandleRemoveExpectedSigner(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// TESTS - HandleSendReminders
// ============================================================================

func TestHandleSendReminders_Success(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
	}

	reminderSvc := &mockReminderService{
		sendRemindersFunc: func(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string, locale string) (*models.ReminderSendResult, error) {
			assert.Equal(t, "doc1", docID)
			assert.Equal(t, "admin@example.com", sentBy)
			assert.Equal(t, "fr", locale)
			return &models.ReminderSendResult{
				TotalAttempted:   2,
				SuccessfullySent: 2,
				Failed:           0,
			}, nil
		},
	}

	handler := createTestHandler(docRepo, nil, reminderSvc, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/reminders", handler.HandleSendReminders)

	reqBody := SendRemindersRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/reminders", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleSendReminders_ServiceNotAvailable(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/reminders", handler.HandleSendReminders)

	reqBody := SendRemindersRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/reminders", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestHandleSendReminders_WithLocale(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
	}

	reminderSvc := &mockReminderService{
		sendRemindersFunc: func(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string, locale string) (*models.ReminderSendResult, error) {
			assert.Equal(t, "en", locale)
			return &models.ReminderSendResult{
				TotalAttempted:   1,
				SuccessfullySent: 1,
			}, nil
		},
	}

	handler := createTestHandler(docRepo, nil, reminderSvc, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/reminders", handler.HandleSendReminders)

	reqBody := SendRemindersRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/reminders", bytes.NewReader(body))
	req.Header.Set("Accept-Language", "en")
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleSendReminders_SpecificEmails(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
	}

	reminderSvc := &mockReminderService{
		sendRemindersFunc: func(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string, locale string) (*models.ReminderSendResult, error) {
			assert.Len(t, specificEmails, 2)
			assert.Contains(t, specificEmails, "user1@example.com")
			assert.Contains(t, specificEmails, "user2@example.com")
			return &models.ReminderSendResult{
				TotalAttempted:   2,
				SuccessfullySent: 2,
			}, nil
		},
	}

	handler := createTestHandler(docRepo, nil, reminderSvc, nil)

	router := chi.NewRouter()
	router.Post("/api/v1/admin/documents/{docId}/reminders", handler.HandleSendReminders)

	reqBody := SendRemindersRequest{
		Emails: []string{"user1@example.com", "user2@example.com"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents/doc1/reminders", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ============================================================================
// TESTS - HandleGetReminderHistory
// ============================================================================

func TestHandleGetReminderHistory_Success(t *testing.T) {
	t.Parallel()

	logs := []*models.ReminderLog{
		createTestReminderLog("doc1", "user1@example.com"),
		createTestReminderLog("doc1", "user2@example.com"),
	}

	reminderSvc := &mockReminderService{
		getReminderHistoryFunc: func(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
			assert.Equal(t, "doc1", docID)
			return logs, nil
		},
	}

	handler := createTestHandler(nil, nil, reminderSvc, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/reminders", handler.HandleGetReminderHistory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/reminders", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data []ReminderLogResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)
}

func TestHandleGetReminderHistory_ServiceNotAvailable(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/reminders", handler.HandleGetReminderHistory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/reminders", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestHandleGetReminderHistory_EmptyHistory(t *testing.T) {
	t.Parallel()

	reminderSvc := &mockReminderService{
		getReminderHistoryFunc: func(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
			return []*models.ReminderLog{}, nil
		},
	}

	handler := createTestHandler(nil, nil, reminderSvc, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/reminders", handler.HandleGetReminderHistory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/reminders", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data []ReminderLogResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 0)
}

// ============================================================================
// TESTS - HandleUpdateDocumentMetadata
// ============================================================================

func TestHandleUpdateDocumentMetadata_CreateNew(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
		createOrUpdateFunc: func(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
			assert.Equal(t, "new-doc", docID)
			assert.Equal(t, "New Document", input.Title)
			assert.Equal(t, "admin@example.com", createdBy)
			return createTestDocument(docID), nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Put("/api/v1/admin/documents/{docId}/metadata", handler.HandleUpdateDocumentMetadata)

	title := "New Document"
	reqBody := UpdateDocumentMetadataRequest{
		Title: &title,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/documents/new-doc/metadata", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleUpdateDocumentMetadata_UpdateExisting(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
		createOrUpdateFunc: func(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
			assert.Equal(t, "Updated Title", input.Title)
			doc.Title = input.Title
			return doc, nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Put("/api/v1/admin/documents/{docId}/metadata", handler.HandleUpdateDocumentMetadata)

	title := "Updated Title"
	reqBody := UpdateDocumentMetadataRequest{
		Title: &title,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/documents/doc1/metadata", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleUpdateDocumentMetadata_AllFields(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return createTestDocument(docID), nil
		},
		createOrUpdateFunc: func(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
			assert.Equal(t, "New Title", input.Title)
			assert.Equal(t, "https://new.example.com/doc.pdf", input.URL)
			assert.Equal(t, "xyz789", input.Checksum)
			assert.Equal(t, "SHA-512", input.ChecksumAlgorithm)
			assert.Equal(t, "New description", input.Description)
			return createTestDocument(docID), nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Put("/api/v1/admin/documents/{docId}/metadata", handler.HandleUpdateDocumentMetadata)

	title := "New Title"
	url := "https://new.example.com/doc.pdf"
	checksum := "xyz789"
	algorithm := "SHA-512"
	description := "New description"
	reqBody := UpdateDocumentMetadataRequest{
		Title:             &title,
		URL:               &url,
		Checksum:          &checksum,
		ChecksumAlgorithm: &algorithm,
		Description:       &description,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/documents/doc1/metadata", bytes.NewReader(body))
	req = req.WithContext(createContextWithUser("admin@example.com", true))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleUpdateDocumentMetadata_NoUser(t *testing.T) {
	t.Parallel()

	handler := createTestHandler(nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Put("/api/v1/admin/documents/{docId}/metadata", handler.HandleUpdateDocumentMetadata)

	title := "New Title"
	reqBody := UpdateDocumentMetadataRequest{
		Title: &title,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/documents/doc1/metadata", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ============================================================================
// TESTS - HandleGetDocumentStatus
// ============================================================================

func TestHandleGetDocumentStatus_Complete(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	signers := []*models.ExpectedSignerWithStatus{
		createTestExpectedSignerWithStatus("doc1", "expected@example.com", true),
	}
	stats := &models.DocCompletionStats{
		DocID:          "doc1",
		ExpectedCount:  1,
		SignedCount:    1,
		PendingCount:   0,
		CompletionRate: 100.0,
	}
	signatures := []*models.Signature{
		{
			ID:          1,
			DocID:       "doc1",
			UserSub:     "unexpected-sub",
			UserEmail:   "unexpected@example.com",
			UserName:    "Unexpected User",
			SignedAtUTC: time.Now(),
		},
	}
	lastSent := time.Now()
	reminderStats := &models.ReminderStats{
		TotalSent:    5,
		PendingCount: 0,
		LastSentAt:   &lastSent,
	}

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
	}
	signerRepo := &mockExpectedSignerRepository{
		listWithStatusByDocIDFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return signers, nil
		},
		getStatsFunc: func(ctx context.Context, docID string) (*models.DocCompletionStats, error) {
			return stats, nil
		},
	}
	sigService := &mockSignatureService{
		getDocumentSignaturesFunc: func(ctx context.Context, docID string) ([]*models.Signature, error) {
			return signatures, nil
		},
	}
	reminderSvc := &mockReminderService{
		getReminderStatsFunc: func(ctx context.Context, docID string) (*models.ReminderStats, error) {
			return reminderStats, nil
		},
	}

	handler := createTestHandler(docRepo, signerRepo, reminderSvc, sigService)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/status", handler.HandleGetDocumentStatus)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/status", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data DocumentStatusResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "doc1", response.Data.DocID)
	assert.NotNil(t, response.Data.Document)
	assert.Len(t, response.Data.ExpectedSigners, 1)
	assert.Len(t, response.Data.UnexpectedSignatures, 1)
	assert.Equal(t, "unexpected@example.com", response.Data.UnexpectedSignatures[0].UserEmail)
	assert.NotNil(t, response.Data.Stats)
	assert.NotNil(t, response.Data.ReminderStats)
	assert.Contains(t, response.Data.ShareLink, "doc1")
}

func TestHandleGetDocumentStatus_MinimalData(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	signerRepo := &mockExpectedSignerRepository{
		listWithStatusByDocIDFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return []*models.ExpectedSignerWithStatus{}, nil
		},
		getStatsFunc: func(ctx context.Context, docID string) (*models.DocCompletionStats, error) {
			return nil, errors.New("no stats")
		},
	}

	handler := createTestHandler(docRepo, signerRepo, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/status", handler.HandleGetDocumentStatus)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/status", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data DocumentStatusResponse `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "doc1", response.Data.DocID)
	assert.Nil(t, response.Data.Document)
	assert.Empty(t, response.Data.ExpectedSigners)
	assert.Empty(t, response.Data.UnexpectedSignatures)
	assert.NotNil(t, response.Data.Stats)
	assert.Equal(t, 0.0, response.Data.Stats.CompletionRate)
}

// ============================================================================
// TESTS - HandleDeleteDocument
// ============================================================================

func TestHandleDeleteDocument_Success(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		deleteFunc: func(ctx context.Context, docID string) error {
			assert.Equal(t, "doc1", docID)
			return nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Delete("/api/v1/admin/documents/{docId}", handler.HandleDeleteDocument)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/documents/doc1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Data map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Data["message"], "deleted successfully")
}

func TestHandleDeleteDocument_RepositoryError(t *testing.T) {
	t.Parallel()

	docRepo := &mockDocumentRepository{
		deleteFunc: func(ctx context.Context, docID string) error {
			return errors.New("database error")
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	router := chi.NewRouter()
	router.Delete("/api/v1/admin/documents/{docId}", handler.HandleDeleteDocument)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/documents/doc1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================================
// TESTS - Helper Functions
// ============================================================================

func TestToDocumentResponse(t *testing.T) {
	t.Parallel()

	doc := createTestDocument("doc1")
	response := toDocumentResponse(doc)

	assert.Equal(t, "doc1", response.DocID)
	assert.Equal(t, "Test Document", response.Title)
	assert.Equal(t, "https://example.com/doc.pdf", response.URL)
	assert.Equal(t, "abc123", response.Checksum)
	assert.Equal(t, "SHA-256", response.ChecksumAlgorithm)
	assert.Equal(t, "Test description", response.Description)
	assert.NotEmpty(t, response.CreatedAt)
	assert.NotEmpty(t, response.UpdatedAt)
	assert.Equal(t, "admin@example.com", response.CreatedBy)
}

func TestToExpectedSignerResponse_WithSignature(t *testing.T) {
	t.Parallel()

	signer := createTestExpectedSignerWithStatus("doc1", "test@example.com", true)
	response := toExpectedSignerResponse(signer)

	assert.Equal(t, "test@example.com", response.Email)
	assert.True(t, response.HasSigned)
	assert.NotNil(t, response.SignedAt)
	assert.NotNil(t, response.UserName)
}

func TestToExpectedSignerResponse_NoSignature(t *testing.T) {
	t.Parallel()

	signer := createTestExpectedSignerWithStatus("doc1", "test@example.com", false)
	response := toExpectedSignerResponse(signer)

	assert.Equal(t, "test@example.com", response.Email)
	assert.False(t, response.HasSigned)
	assert.Nil(t, response.SignedAt)
}

func TestToStatsResponse(t *testing.T) {
	t.Parallel()

	stats := &models.DocCompletionStats{
		DocID:          "doc1",
		ExpectedCount:  10,
		SignedCount:    7,
		PendingCount:   3,
		CompletionRate: 70.0,
	}

	response := toStatsResponse(stats)

	assert.Equal(t, "doc1", response.DocID)
	assert.Equal(t, 10, response.ExpectedCount)
	assert.Equal(t, 7, response.SignedCount)
	assert.Equal(t, 3, response.PendingCount)
	assert.Equal(t, 70.0, response.CompletionRate)
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkHandleListDocuments(b *testing.B) {
	docs := []*models.Document{
		createTestDocument("doc1"),
		createTestDocument("doc2"),
		createTestDocument("doc3"),
	}

	docRepo := &mockDocumentRepository{
		listFunc: func(ctx context.Context, limit, offset int) ([]*models.Document, error) {
			return docs, nil
		},
	}

	handler := createTestHandler(docRepo, nil, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents", nil)
		rec := httptest.NewRecorder()
		handler.HandleListDocuments(rec, req)
	}
}

func BenchmarkHandleGetDocumentStatus(b *testing.B) {
	doc := createTestDocument("doc1")
	signers := []*models.ExpectedSignerWithStatus{
		createTestExpectedSignerWithStatus("doc1", "signer1@example.com", true),
		createTestExpectedSignerWithStatus("doc1", "signer2@example.com", false),
	}
	stats := &models.DocCompletionStats{
		DocID:          "doc1",
		ExpectedCount:  2,
		SignedCount:    1,
		PendingCount:   1,
		CompletionRate: 50.0,
	}

	docRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			return doc, nil
		},
	}
	signerRepo := &mockExpectedSignerRepository{
		listWithStatusByDocIDFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return signers, nil
		},
		getStatsFunc: func(ctx context.Context, docID string) (*models.DocCompletionStats, error) {
			return stats, nil
		},
	}

	handler := createTestHandler(docRepo, signerRepo, nil, nil)

	router := chi.NewRouter()
	router.Get("/api/v1/admin/documents/{docId}/status", handler.HandleGetDocumentStatus)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/doc1/status", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}
}

func BenchmarkToDocumentResponse(b *testing.B) {
	doc := createTestDocument("doc1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = toDocumentResponse(doc)
	}
}

func BenchmarkToExpectedSignerResponse(b *testing.B) {
	signer := createTestExpectedSignerWithStatus("doc1", "test@example.com", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = toExpectedSignerResponse(signer)
	}
}
