// SPDX-License-Identifier: AGPL-3.0-or-later
package documents

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// documentService defines the interface for document operations
type documentService interface {
	CreateDocument(ctx context.Context, req services.CreateDocumentRequest) (*models.Document, error)
	FindOrCreateDocument(ctx context.Context, ref string) (*models.Document, bool, error)
	FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error)
}

// webhookPublisher defines minimal publish capability
type webhookPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]interface{}) error
}

// Handler handles document API requests
type Handler struct {
	signatureService *services.SignatureService
	documentService  documentService
	webhookPublisher webhookPublisher
}

// NewHandler creates a new documents handler
// Backward-compatible constructor used by tests and existing code
func NewHandler(signatureService *services.SignatureService, documentService documentService) *Handler {
	return &Handler{signatureService: signatureService, documentService: documentService}
}

// Extended constructor with webhook publisher
func NewHandlerWithPublisher(signatureService *services.SignatureService, documentService documentService, publisher webhookPublisher) *Handler {
	return &Handler{signatureService: signatureService, documentService: documentService, webhookPublisher: publisher}
}

// DocumentDTO represents a document data transfer object
type DocumentDTO struct {
	ID                  string                 `json:"id"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	CreatedAt           string                 `json:"createdAt,omitempty"`
	UpdatedAt           string                 `json:"updatedAt,omitempty"`
	SignatureCount      int                    `json:"signatureCount"`
	ExpectedSignerCount int                    `json:"expectedSignerCount"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// SignatureDTO represents a signature data transfer object
type SignatureDTO struct {
	ID          string `json:"id"`
	DocID       string `json:"docId"`
	UserEmail   string `json:"userEmail"`
	UserName    string `json:"userName,omitempty"`
	SignedAt    string `json:"signedAt"`
	Signature   string `json:"signature"`
	PayloadHash string `json:"payloadHash"`
	Nonce       string `json:"nonce"`
	PrevHash    string `json:"prevHash,omitempty"`
}

// CreateDocumentRequest represents the request body for creating a document
type CreateDocumentRequest struct {
	Reference string `json:"reference"`
	Title     string `json:"title,omitempty"`
}

// CreateDocumentResponse represents the response for creating a document
type CreateDocumentResponse struct {
	DocID     string `json:"docId"`
	URL       string `json:"url,omitempty"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
}

// HandleCreateDocument handles POST /api/v1/documents
func (h *Handler) HandleCreateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Warn("Invalid document creation request body",
			"error", err.Error(),
			"remote_addr", r.RemoteAddr)
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Validate reference field
	if req.Reference == "" {
		logger.Logger.Warn("Document creation request missing reference field",
			"remote_addr", r.RemoteAddr)
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Reference is required", nil)
		return
	}

	logger.Logger.Info("Document creation request received",
		"reference", req.Reference,
		"has_title", req.Title != "",
		"remote_addr", r.RemoteAddr)

	// Create document request
	docRequest := services.CreateDocumentRequest{
		Reference: req.Reference,
		Title:     req.Title,
	}

	// Create document
	doc, err := h.documentService.CreateDocument(ctx, docRequest)
	if err != nil {
		logger.Logger.Error("Document creation failed in handler",
			"reference", req.Reference,
			"error", err.Error())
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to create document", map[string]interface{}{"error": err.Error()})
		return
	}

	logger.Logger.Info("Document creation succeeded",
		"doc_id", doc.DocID,
		"title", doc.Title,
		"has_url", doc.URL != "")

	// Publish webhook event
	if h.webhookPublisher != nil {
		_ = h.webhookPublisher.Publish(ctx, "document.created", map[string]interface{}{
			"doc_id":             doc.DocID,
			"title":              doc.Title,
			"url":                doc.URL,
			"checksum":           doc.Checksum,
			"checksum_algorithm": doc.ChecksumAlgorithm,
		})
	}

	// Return the created document
	response := CreateDocumentResponse{
		DocID:     doc.DocID,
		URL:       doc.URL,
		Title:     doc.Title,
		CreatedAt: doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	shared.WriteJSON(w, http.StatusCreated, response)
}

// HandleListDocuments handles GET /api/v1/documents
func (h *Handler) HandleListDocuments(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page := 1
	limit := 20
	_ = r.URL.Query().Get("search") // TODO: implement search

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// For now, return empty list (we'll implement document listing later)
	documents := []DocumentDTO{}

	// TODO: Implement actual document listing from database
	// This would require adding a document repository and service

	total := 0
	shared.WritePaginatedJSON(w, documents, page, limit, total)
}

// HandleGetDocument handles GET /api/v1/documents/{docId}
func (h *Handler) HandleGetDocument(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		shared.WriteValidationError(w, "Document ID is required", nil)
		return
	}

	// Get signatures for the document
	signatures, err := h.signatureService.GetDocumentSignatures(r.Context(), docID)
	if err != nil {
		shared.WriteInternalError(w)
		return
	}

	// Build document response
	// TODO: Get actual document metadata from database
	document := DocumentDTO{
		ID:             docID,
		Title:          "Document " + docID, // Placeholder
		Description:    "",
		SignatureCount: len(signatures),
		// ExpectedSignerCount will be populated when we have the expected signers repository
	}

	shared.WriteJSON(w, http.StatusOK, document)
}

// HandleGetDocumentSignatures handles GET /api/v1/documents/{docId}/signatures
func (h *Handler) HandleGetDocumentSignatures(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		shared.WriteValidationError(w, "Document ID is required", nil)
		return
	}

	ctx := r.Context()

	signatures, err := h.signatureService.GetDocumentSignatures(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to get signatures",
			"doc_id", docID,
			"error", err.Error())
		shared.WriteInternalError(w)
		return
	}

	// Convert to DTOs
	dtos := make([]SignatureDTO, len(signatures))
	for i := range signatures {
		dtos[i] = signatureToDTO(signatures[i])
	}

	shared.WriteJSON(w, http.StatusOK, dtos)
}

// HandleGetExpectedSigners handles GET /api/v1/documents/{docId}/expected-signers
func (h *Handler) HandleGetExpectedSigners(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		shared.WriteValidationError(w, "Document ID is required", nil)
		return
	}

	// TODO: Implement with expected signers repository
	expectedSigners := []interface{}{}

	shared.WriteJSON(w, http.StatusOK, expectedSigners)
}

// Helper function to convert signature model to DTO
func signatureToDTO(sig *models.Signature) SignatureDTO {
	dto := SignatureDTO{
		ID:          strconv.FormatInt(sig.ID, 10),
		DocID:       sig.DocID,
		UserEmail:   sig.UserEmail,
		UserName:    sig.UserName,
		SignedAt:    sig.SignedAtUTC.Format("2006-01-02T15:04:05Z07:00"),
		Signature:   sig.Signature,
		PayloadHash: sig.PayloadHash,
		Nonce:       sig.Nonce,
	}

	if sig.PrevHash != nil && *sig.PrevHash != "" {
		dto.PrevHash = *sig.PrevHash
	}

	return dto
}

// FindOrCreateDocumentResponse represents the response for finding or creating a document
type FindOrCreateDocumentResponse struct {
	DocID             string `json:"docId"`
	URL               string `json:"url,omitempty"`
	Title             string `json:"title"`
	Checksum          string `json:"checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksumAlgorithm,omitempty"`
	Description       string `json:"description,omitempty"`
	CreatedAt         string `json:"createdAt"`
	IsNew             bool   `json:"isNew"`
}

// HandleFindOrCreateDocument handles GET /api/v1/documents/find-or-create?ref={reference}
func (h *Handler) HandleFindOrCreateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get reference from query parameter
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		logger.Logger.Warn("Find or create request missing ref parameter",
			"remote_addr", r.RemoteAddr)
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "ref parameter is required", nil)
		return
	}

	logger.Logger.Info("Find or create document request",
		"reference", ref,
		"remote_addr", r.RemoteAddr)

	// Check if user is authenticated
	user, isAuthenticated := shared.GetUserFromContext(ctx)

	// First, try to find the document (without creating)
	refType := detectReferenceType(ref)
	existingDoc, err := h.documentService.FindByReference(ctx, ref, string(refType))
	if err != nil {
		logger.Logger.Error("Failed to search for document",
			"reference", ref,
			"error", err.Error())
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to search for document", map[string]interface{}{"error": err.Error()})
		return
	}

	// If document exists, return it
	if existingDoc != nil {
		logger.Logger.Info("Document found",
			"doc_id", existingDoc.DocID,
			"reference", ref)

		response := FindOrCreateDocumentResponse{
			DocID:             existingDoc.DocID,
			URL:               existingDoc.URL,
			Title:             existingDoc.Title,
			Checksum:          existingDoc.Checksum,
			ChecksumAlgorithm: existingDoc.ChecksumAlgorithm,
			Description:       existingDoc.Description,
			CreatedAt:         existingDoc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsNew:             false,
		}

		shared.WriteJSON(w, http.StatusOK, response)
		return
	}

	// Document doesn't exist - check authentication before creating
	if !isAuthenticated {
		logger.Logger.Warn("Unauthenticated user attempted to create document",
			"reference", ref,
			"remote_addr", r.RemoteAddr)
		shared.WriteError(w, http.StatusUnauthorized, shared.ErrCodeUnauthorized, "Authentication required to create document", nil)
		return
	}

	// User is authenticated, create the document
	doc, isNew, err := h.documentService.FindOrCreateDocument(ctx, ref)
	if err != nil {
		logger.Logger.Error("Failed to create document",
			"reference", ref,
			"error", err.Error())
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to create document", map[string]interface{}{"error": err.Error()})
		return
	}

	logger.Logger.Info("Document created",
		"doc_id", doc.DocID,
		"reference", ref,
		"user_email", user.Email)

	// Build response
	response := FindOrCreateDocumentResponse{
		DocID:             doc.DocID,
		URL:               doc.URL,
		Title:             doc.Title,
		Checksum:          doc.Checksum,
		ChecksumAlgorithm: doc.ChecksumAlgorithm,
		Description:       doc.Description,
		CreatedAt:         doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsNew:             isNew,
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func detectReferenceType(ref string) ReferenceType {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return "url"
	}

	if strings.Contains(ref, "/") || strings.Contains(ref, "\\") {
		return "path"
	}

	return "reference"
}

type ReferenceType string
