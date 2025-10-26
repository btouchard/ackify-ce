// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/i18n"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/go-chi/chi/v5"
)

// documentRepository defines the interface for document operations
type documentRepository interface {
	GetByDocID(ctx context.Context, docID string) (*models.Document, error)
	List(ctx context.Context, limit, offset int) ([]*models.Document, error)
	CreateOrUpdate(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error)
	Delete(ctx context.Context, docID string) error
}

// expectedSignerRepository defines the interface for expected signer operations
type expectedSignerRepository interface {
	ListByDocID(ctx context.Context, docID string) ([]*models.ExpectedSigner, error)
	ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error)
	AddExpected(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error
	Remove(ctx context.Context, docID, email string) error
	GetStats(ctx context.Context, docID string) (*models.DocCompletionStats, error)
}

// reminderService defines the interface for reminder operations
type reminderService interface {
	SendReminders(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string, locale string) (*models.ReminderSendResult, error)
	GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error)
	GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error)
}

// signatureService defines the interface for signature operations
type signatureService interface {
	GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error)
}

// Handler handles admin API requests
type Handler struct {
	documentRepo       documentRepository
	expectedSignerRepo expectedSignerRepository
	reminderService    reminderService
	signatureService   signatureService
	baseURL            string
}

// NewHandler creates a new admin handler
func NewHandler(documentRepo documentRepository, expectedSignerRepo expectedSignerRepository, reminderService reminderService, signatureService signatureService, baseURL string) *Handler {
	return &Handler{
		documentRepo:       documentRepo,
		expectedSignerRepo: expectedSignerRepo,
		reminderService:    reminderService,
		signatureService:   signatureService,
		baseURL:            baseURL,
	}
}

// DocumentResponse represents a document in API responses
type DocumentResponse struct {
	DocID             string `json:"docId"`
	Title             string `json:"title"`
	URL               string `json:"url"`
	Checksum          string `json:"checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksumAlgorithm,omitempty"`
	Description       string `json:"description"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	CreatedBy         string `json:"createdBy"`
}

// ExpectedSignerResponse represents an expected signer in API responses
type ExpectedSignerResponse struct {
	ID                    int64   `json:"id"`
	DocID                 string  `json:"docId"`
	Email                 string  `json:"email"`
	Name                  string  `json:"name"`
	AddedAt               string  `json:"addedAt"`
	AddedBy               string  `json:"addedBy"`
	Notes                 *string `json:"notes,omitempty"`
	HasSigned             bool    `json:"hasSigned"`
	SignedAt              *string `json:"signedAt,omitempty"`
	UserName              *string `json:"userName,omitempty"`
	LastReminderSent      *string `json:"lastReminderSent,omitempty"`
	ReminderCount         int     `json:"reminderCount"`
	DaysSinceAdded        int     `json:"daysSinceAdded"`
	DaysSinceLastReminder *int    `json:"daysSinceLastReminder,omitempty"`
}

// DocumentStatsResponse represents document statistics
type DocumentStatsResponse struct {
	DocID          string  `json:"docId"`
	ExpectedCount  int     `json:"expectedCount"`
	SignedCount    int     `json:"signedCount"`
	PendingCount   int     `json:"pendingCount"`
	CompletionRate float64 `json:"completionRate"`
}

// UnexpectedSignatureResponse represents an unexpected signature
type UnexpectedSignatureResponse struct {
	UserEmail   string  `json:"userEmail"`
	UserName    *string `json:"userName,omitempty"`
	SignedAtUTC string  `json:"signedAtUTC"`
}

// HandleListDocuments handles GET /api/v1/admin/documents
func (h *Handler) HandleListDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: Add pagination parameters
	limit := 100
	offset := 0

	documents, err := h.documentRepo.List(ctx, limit, offset)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to list documents", nil)
		return
	}

	response := make([]*DocumentResponse, 0, len(documents))
	for _, doc := range documents {
		response = append(response, toDocumentResponse(doc))
	}

	meta := map[string]interface{}{
		"total":  len(documents), // For now, just return count of results
		"limit":  limit,
		"offset": offset,
	}

	shared.WriteJSONWithMeta(w, http.StatusOK, response, meta)
}

// HandleGetDocument handles GET /api/v1/admin/documents/{docId}
func (h *Handler) HandleGetDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	document, err := h.documentRepo.GetByDocID(ctx, docID)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, shared.ErrCodeNotFound, "Document not found", nil)
		return
	}

	shared.WriteJSON(w, http.StatusOK, toDocumentResponse(document))
}

// HandleGetDocumentWithSigners handles GET /api/v1/admin/documents/{docId}/signers
func (h *Handler) HandleGetDocumentWithSigners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Get document
	document, err := h.documentRepo.GetByDocID(ctx, docID)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, shared.ErrCodeNotFound, "Document not found", nil)
		return
	}

	// Get expected signers with status
	signers, err := h.expectedSignerRepo.ListWithStatusByDocID(ctx, docID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to get signers", nil)
		return
	}

	// Get completion stats
	stats, err := h.expectedSignerRepo.GetStats(ctx, docID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to get stats", nil)
		return
	}

	signersResponse := make([]*ExpectedSignerResponse, 0, len(signers))
	for _, signer := range signers {
		signersResponse = append(signersResponse, toExpectedSignerResponse(signer))
	}

	response := map[string]interface{}{
		"document": toDocumentResponse(document),
		"signers":  signersResponse,
		"stats":    toStatsResponse(stats),
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

// AddExpectedSignerRequest represents the request body for adding an expected signer
type AddExpectedSignerRequest struct {
	Email string  `json:"email"`
	Name  string  `json:"name"`
	Notes *string `json:"notes,omitempty"`
}

// HandleAddExpectedSigner handles POST /api/v1/admin/documents/{docId}/signers
func (h *Handler) HandleAddExpectedSigner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Get user from context
	user, ok := shared.GetUserFromContext(ctx)
	if !ok {
		shared.WriteUnauthorized(w, "")
		return
	}

	// Parse request body
	var req AddExpectedSignerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	// Validate
	if req.Email == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Email is required", nil)
		return
	}

	// Add expected signer
	contacts := []models.ContactInfo{{Email: req.Email, Name: req.Name}}
	err := h.expectedSignerRepo.AddExpected(ctx, docID, contacts, user.Email)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to add expected signer", nil)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Expected signer added successfully",
		"email":   req.Email,
	})
}

// HandleRemoveExpectedSigner handles DELETE /api/v1/admin/documents/{docId}/signers/{email}
func (h *Handler) HandleRemoveExpectedSigner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")
	email := chi.URLParam(r, "email")

	if docID == "" || email == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID and email are required", nil)
		return
	}

	// Remove expected signer
	err := h.expectedSignerRepo.Remove(ctx, docID, email)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to remove expected signer", nil)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Expected signer removed successfully",
	})
}

// Helper functions to convert models to API responses
func toDocumentResponse(doc *models.Document) *DocumentResponse {
	return &DocumentResponse{
		DocID:             doc.DocID,
		Title:             doc.Title,
		URL:               doc.URL,
		Checksum:          doc.Checksum,
		ChecksumAlgorithm: doc.ChecksumAlgorithm,
		Description:       doc.Description,
		CreatedAt:         doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:         doc.CreatedBy,
	}
}

func toExpectedSignerResponse(signer *models.ExpectedSignerWithStatus) *ExpectedSignerResponse {
	response := &ExpectedSignerResponse{
		ID:                    signer.ID,
		DocID:                 signer.DocID,
		Email:                 signer.Email,
		Name:                  signer.Name,
		AddedAt:               signer.AddedAt.Format("2006-01-02T15:04:05Z07:00"),
		AddedBy:               signer.AddedBy,
		Notes:                 signer.Notes,
		HasSigned:             signer.HasSigned,
		UserName:              signer.UserName,
		ReminderCount:         signer.ReminderCount,
		DaysSinceAdded:        signer.DaysSinceAdded,
		DaysSinceLastReminder: signer.DaysSinceLastReminder,
	}

	if signer.SignedAt != nil {
		signedAt := signer.SignedAt.Format("2006-01-02T15:04:05Z07:00")
		response.SignedAt = &signedAt
	}

	if signer.LastReminderSent != nil {
		lastReminder := signer.LastReminderSent.Format("2006-01-02T15:04:05Z07:00")
		response.LastReminderSent = &lastReminder
	}

	return response
}

func toStatsResponse(stats *models.DocCompletionStats) *DocumentStatsResponse {
	return &DocumentStatsResponse{
		DocID:          stats.DocID,
		ExpectedCount:  stats.ExpectedCount,
		SignedCount:    stats.SignedCount,
		PendingCount:   stats.PendingCount,
		CompletionRate: stats.CompletionRate,
	}
}

// SendRemindersRequest represents the request body for sending reminders
type SendRemindersRequest struct {
	Emails []string `json:"emails,omitempty"` // If empty, send to all pending signers
}

// HandleSendReminders handles POST /api/v1/admin/documents/{docId}/reminders
func (h *Handler) HandleSendReminders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Check if reminder service is available
	if h.reminderService == nil {
		shared.WriteError(w, http.StatusServiceUnavailable, shared.ErrCodeInternal, "Reminder service not configured", nil)
		return
	}

	// Get user from context
	user, ok := shared.GetUserFromContext(ctx)
	if !ok {
		shared.WriteUnauthorized(w, "")
		return
	}

	// Parse request body
	var req SendRemindersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	// Get document URL from metadata
	var docURL string
	if doc, err := h.documentRepo.GetByDocID(ctx, docID); err == nil && doc != nil && doc.URL != "" {
		docURL = doc.URL
	}

	// Get locale from request using i18n helper
	locale := i18n.GetLangFromRequest(r)

	// Send reminders
	result, err := h.reminderService.SendReminders(ctx, docID, user.Email, req.Emails, docURL, locale)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to send reminders", nil)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Reminders sent",
		"result":  result,
	})
}

// ReminderLogResponse represents a reminder log entry in API responses
type ReminderLogResponse struct {
	ID             int64   `json:"id"`
	DocID          string  `json:"docId"`
	RecipientEmail string  `json:"recipientEmail"`
	SentAt         string  `json:"sentAt"`
	SentBy         string  `json:"sentBy"`
	TemplateUsed   string  `json:"templateUsed"`
	Status         string  `json:"status"`
	ErrorMessage   *string `json:"errorMessage,omitempty"`
}

// HandleGetReminderHistory handles GET /api/v1/admin/documents/{docId}/reminders
func (h *Handler) HandleGetReminderHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Check if reminder service is available
	if h.reminderService == nil {
		shared.WriteError(w, http.StatusServiceUnavailable, shared.ErrCodeInternal, "Reminder service not configured", nil)
		return
	}

	history, err := h.reminderService.GetReminderHistory(ctx, docID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to get reminder history", nil)
		return
	}

	response := make([]*ReminderLogResponse, 0, len(history))
	for _, log := range history {
		response = append(response, &ReminderLogResponse{
			ID:             log.ID,
			DocID:          log.DocID,
			RecipientEmail: log.RecipientEmail,
			SentAt:         log.SentAt.Format("2006-01-02T15:04:05Z07:00"),
			SentBy:         log.SentBy,
			TemplateUsed:   log.TemplateUsed,
			Status:         log.Status,
			ErrorMessage:   log.ErrorMessage,
		})
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

// UpdateDocumentMetadataRequest represents the request body for updating document metadata
type UpdateDocumentMetadataRequest struct {
	Title             *string `json:"title,omitempty"`
	URL               *string `json:"url,omitempty"`
	Checksum          *string `json:"checksum,omitempty"`
	ChecksumAlgorithm *string `json:"checksumAlgorithm,omitempty"`
	Description       *string `json:"description,omitempty"`
}

// HandleUpdateDocumentMetadata handles PUT /api/v1/admin/documents/{docId}/metadata
func (h *Handler) HandleUpdateDocumentMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Get user from context
	user, ok := shared.GetUserFromContext(ctx)
	if !ok {
		shared.WriteUnauthorized(w, "")
		return
	}

	// Parse request body
	var req UpdateDocumentMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	// Get existing document or create new one
	doc, err := h.documentRepo.GetByDocID(ctx, docID)
	if err != nil || doc == nil {
		// Document doesn't exist, create a new one
		doc = &models.Document{
			DocID:     docID,
			CreatedBy: user.Email,
		}
	}

	// Update fields if provided
	if req.Title != nil {
		doc.Title = *req.Title
	}
	if req.URL != nil {
		doc.URL = *req.URL
	}
	if req.Checksum != nil {
		doc.Checksum = *req.Checksum
	}
	if req.ChecksumAlgorithm != nil {
		doc.ChecksumAlgorithm = *req.ChecksumAlgorithm
	}
	if req.Description != nil {
		doc.Description = *req.Description
	}

	// Save document using CreateOrUpdate
	input := models.DocumentInput{
		Title:             doc.Title,
		URL:               doc.URL,
		Checksum:          doc.Checksum,
		ChecksumAlgorithm: doc.ChecksumAlgorithm,
		Description:       doc.Description,
	}
	doc, err = h.documentRepo.CreateOrUpdate(ctx, docID, input, user.Email)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to update document metadata", nil)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Document metadata updated successfully",
		"document": toDocumentResponse(doc),
	})
}

// DocumentStatusResponse represents complete document status including everything
type DocumentStatusResponse struct {
	DocID                string                         `json:"docId"`
	Document             *DocumentResponse              `json:"document,omitempty"`
	ExpectedSigners      []*ExpectedSignerResponse      `json:"expectedSigners"`
	UnexpectedSignatures []*UnexpectedSignatureResponse `json:"unexpectedSignatures"`
	Stats                *DocumentStatsResponse         `json:"stats"`
	ReminderStats        *ReminderStatsResponse         `json:"reminderStats,omitempty"`
	ShareLink            string                         `json:"shareLink"`
}

// ReminderStatsResponse represents reminder statistics
type ReminderStatsResponse struct {
	TotalSent    int     `json:"totalSent"`
	PendingCount int     `json:"pendingCount"`
	LastSentAt   *string `json:"lastSentAt,omitempty"`
}

// HandleGetDocumentStatus handles GET /api/v1/admin/documents/{docId}/status
func (h *Handler) HandleGetDocumentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	response := &DocumentStatusResponse{
		DocID:                docID,
		ExpectedSigners:      []*ExpectedSignerResponse{},
		UnexpectedSignatures: []*UnexpectedSignatureResponse{},
		ShareLink:            h.baseURL + "/?doc=" + docID,
	}

	// Get document (optional)
	if doc, err := h.documentRepo.GetByDocID(ctx, docID); err == nil && doc != nil {
		response.Document = toDocumentResponse(doc)
	}

	// Get expected signers with status
	expectedEmails := make(map[string]bool)
	if signers, err := h.expectedSignerRepo.ListWithStatusByDocID(ctx, docID); err == nil {
		for _, signer := range signers {
			response.ExpectedSigners = append(response.ExpectedSigners, toExpectedSignerResponse(signer))
			expectedEmails[signer.Email] = true
		}
	}

	// Get all signatures for this document and find unexpected ones
	if h.signatureService != nil {
		if signatures, err := h.signatureService.GetDocumentSignatures(ctx, docID); err == nil {
			for _, sig := range signatures {
				// If this signature's email is not in the expected list, it's unexpected
				if !expectedEmails[sig.UserEmail] {
					userName := sig.UserName
					response.UnexpectedSignatures = append(response.UnexpectedSignatures, &UnexpectedSignatureResponse{
						UserEmail:   sig.UserEmail,
						UserName:    &userName,
						SignedAtUTC: sig.SignedAtUTC.Format("2006-01-02T15:04:05Z07:00"),
					})
				}
			}
		}
	}

	// Get completion stats
	if stats, err := h.expectedSignerRepo.GetStats(ctx, docID); err == nil {
		response.Stats = toStatsResponse(stats)
	} else {
		// Default stats if no expected signers
		response.Stats = &DocumentStatsResponse{
			DocID:          docID,
			ExpectedCount:  0,
			SignedCount:    0,
			PendingCount:   0,
			CompletionRate: 0,
		}
	}

	// Get reminder stats if service available
	if h.reminderService != nil {
		if reminderStats, err := h.reminderService.GetReminderStats(ctx, docID); err == nil {
			var lastSentAt *string
			if reminderStats.LastSentAt != nil {
				formatted := reminderStats.LastSentAt.Format("2006-01-02T15:04:05Z07:00")
				lastSentAt = &formatted
			}
			response.ReminderStats = &ReminderStatsResponse{
				TotalSent:    reminderStats.TotalSent,
				PendingCount: reminderStats.PendingCount,
				LastSentAt:   lastSentAt,
			}
		}
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

// HandleDeleteDocument handles DELETE /api/v1/admin/documents/{docId}
func (h *Handler) HandleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Delete document (this will cascade delete signatures and expected signers due to DB constraints)
	err := h.documentRepo.Delete(ctx, docID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to delete document", nil)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Document deleted successfully",
	})
}
