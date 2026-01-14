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
	"github.com/btouchard/ackify-ce/backend/pkg/providers"
)

// documentService defines the interface for document operations
type documentService interface {
	CreateDocument(ctx context.Context, req services.CreateDocumentRequest) (*models.Document, error)
	FindOrCreateDocument(ctx context.Context, ref string) (*models.Document, bool, error)
	FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error)
	List(ctx context.Context, limit, offset int) ([]*models.Document, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Document, error)
	Count(ctx context.Context, searchQuery string) (int, error)
	GetByDocID(ctx context.Context, docID string) (*models.Document, error)
	GetExpectedSignerStats(ctx context.Context, docID string) (*models.DocCompletionStats, error)
	ListExpectedSigners(ctx context.Context, docID string) ([]*models.ExpectedSigner, error)
	ListByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Document, error)
	SearchByCreatedBy(ctx context.Context, createdBy, query string, limit, offset int) ([]*models.Document, error)
	CountByCreatedBy(ctx context.Context, createdBy, searchQuery string) (int, error)
}

// webhookPublisher defines minimal publish capability
type webhookPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]interface{}) error
}

// signatureService defines the interface for signature operations
type signatureService interface {
	GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error)
}

// Handler handles document API requests
type Handler struct {
	signatureService signatureService
	documentService  documentService
	webhookPublisher webhookPublisher
	authorizer       providers.Authorizer
}

// NewHandler creates a handler with all dependencies for full functionality
func NewHandler(
	signatureService signatureService,
	documentService documentService,
	publisher webhookPublisher,
	authorizer providers.Authorizer,
) *Handler {
	return &Handler{
		signatureService: signatureService,
		documentService:  documentService,
		webhookPublisher: publisher,
		authorizer:       authorizer,
	}
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

	user, authenticated := shared.GetUserFromContext(ctx)
	userEmail := ""
	if authenticated && user != nil {
		userEmail = user.Email
	}

	if !h.authorizer.CanCreateDocument(ctx, userEmail) {
		if !authenticated {
			logger.Logger.Warn("Unauthenticated user attempted to create document",
				"remote_addr", r.RemoteAddr)
			shared.WriteError(w, http.StatusUnauthorized, shared.ErrCodeUnauthorized, "Authentication required to create document", nil)
			return
		}
		logger.Logger.Warn("Non-admin user attempted to create document",
			"user_email", user.Email,
			"remote_addr", r.RemoteAddr)
		shared.WriteError(w, http.StatusForbidden, shared.ErrCodeForbidden, "Only administrators can create documents", nil)
		return
	}

	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Warn("Invalid document creation request body",
			"error", err.Error(),
			"remote_addr", r.RemoteAddr)
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

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

	docRequest := services.CreateDocumentRequest{
		Reference: req.Reference,
		Title:     req.Title,
	}

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
	ctx := r.Context()

	pagination := shared.ParsePaginationParams(r, 20, 100)
	searchQuery := r.URL.Query().Get("search")

	var docs []*models.Document
	var err error

	if searchQuery != "" {
		// Use search if query is provided
		docs, err = h.documentService.Search(ctx, searchQuery, pagination.PageSize, pagination.Offset)
		logger.Logger.Debug("Public document search request",
			"query", searchQuery,
			"limit", pagination.PageSize,
			"offset", pagination.Offset)
	} else {
		// Otherwise, list all documents
		docs, err = h.documentService.List(ctx, pagination.PageSize, pagination.Offset)
		logger.Logger.Debug("Public document list request",
			"limit", pagination.PageSize,
			"offset", pagination.Offset)
	}

	if err != nil {
		logger.Logger.Error("Failed to fetch documents",
			"search", searchQuery,
			"error", err.Error())
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to fetch documents", nil)
		return
	}

	totalCount, err := h.documentService.Count(ctx, searchQuery)
	if err != nil {
		logger.Logger.Warn("Failed to count documents, using result count",
			"error", err.Error(),
			"search", searchQuery)
		totalCount = len(docs)
	}

	documents := make([]DocumentDTO, 0, len(docs))
	for _, doc := range docs {
		dto := DocumentDTO{
			ID:          doc.DocID,
			Title:       doc.Title,
			Description: doc.Description,
			CreatedAt:   doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if sigs, err := h.signatureService.GetDocumentSignatures(ctx, doc.DocID); err == nil {
			dto.SignatureCount = len(sigs)
		}

		if stats, err := h.documentService.GetExpectedSignerStats(ctx, doc.DocID); err == nil {
			dto.ExpectedSignerCount = stats.ExpectedCount
		}

		documents = append(documents, dto)
	}

	shared.WritePaginatedJSON(w, documents, pagination.Page, pagination.PageSize, totalCount)
}

// HandleGetDocument handles GET /api/v1/documents/{docId}
func (h *Handler) HandleGetDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteValidationError(w, "Document ID is required", nil)
		return
	}

	doc, err := h.documentService.GetByDocID(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to get document", "doc_id", docID, "error", err.Error())
		shared.WriteInternalError(w)
		return
	}
	if doc == nil {
		shared.WriteError(w, http.StatusNotFound, shared.ErrCodeNotFound, "Document not found", nil)
		return
	}

	signatures, err := h.signatureService.GetDocumentSignatures(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to get signatures", "doc_id", docID, "error", err.Error())
		shared.WriteInternalError(w)
		return
	}

	// Build response
	response := DocumentDTO{
		ID:             docID,
		Title:          doc.Title,
		Description:    doc.Description,
		CreatedAt:      doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		SignatureCount: len(signatures),
	}

	// Get expected signer count
	if stats, err := h.documentService.GetExpectedSignerStats(ctx, docID); err == nil {
		response.ExpectedSignerCount = stats.ExpectedCount
	}

	shared.WriteJSON(w, http.StatusOK, response)
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

// PublicExpectedSigner represents an expected signer in public API (minimal info)
type PublicExpectedSigner struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// HandleGetExpectedSigners handles GET /api/v1/documents/{docId}/expected-signers
func (h *Handler) HandleGetExpectedSigners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	if docID == "" {
		shared.WriteValidationError(w, "Document ID is required", nil)
		return
	}

	// Get expected signers (public version - without internal notes/metadata)
	signers, err := h.documentService.ListExpectedSigners(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to get expected signers", "doc_id", docID, "error", err.Error())
		shared.WriteInternalError(w)
		return
	}

	// Convert to public DTO (minimal info - no notes, no internal metadata)
	response := make([]PublicExpectedSigner, 0, len(signers))
	for _, signer := range signers {
		response = append(response, PublicExpectedSigner{
			Email: signer.Email,
			Name:  signer.Name,
		})
	}

	shared.WriteJSON(w, http.StatusOK, response)
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
	ReadMode          string `json:"readMode"`
	AllowDownload     bool   `json:"allowDownload"`
	RequireFullRead   bool   `json:"requireFullRead"`
	VerifyChecksum    bool   `json:"verifyChecksum"`
	CreatedAt         string `json:"createdAt"`
	IsNew             bool   `json:"isNew"`
	// Storage fields for uploaded documents
	StorageKey string `json:"storageKey,omitempty"`
	MimeType   string `json:"mimeType,omitempty"`
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
			ReadMode:          existingDoc.ReadMode,
			AllowDownload:     existingDoc.AllowDownload,
			RequireFullRead:   existingDoc.RequireFullRead,
			VerifyChecksum:    existingDoc.VerifyChecksum,
			CreatedAt:         existingDoc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsNew:             false,
			StorageKey:        existingDoc.StorageKey,
			MimeType:          existingDoc.MimeType,
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

	// Check if user can create documents
	if !h.authorizer.CanCreateDocument(ctx, user.Email) {
		logger.Logger.Warn("Non-admin user attempted to create document via find-or-create",
			"user_email", user.Email,
			"reference", ref,
			"remote_addr", r.RemoteAddr)
		shared.WriteError(w, http.StatusForbidden, shared.ErrCodeForbidden, "Only administrators can create documents", nil)
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
		ReadMode:          doc.ReadMode,
		AllowDownload:     doc.AllowDownload,
		RequireFullRead:   doc.RequireFullRead,
		VerifyChecksum:    doc.VerifyChecksum,
		CreatedAt:         doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsNew:             isNew,
		StorageKey:        doc.StorageKey,
		MimeType:          doc.MimeType,
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

// MyDocumentDTO represents a document with stats for the current user's documents list
type MyDocumentDTO struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	URL                 string `json:"url,omitempty"`
	Description         string `json:"description"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
	SignatureCount      int    `json:"signatureCount"`
	ExpectedSignerCount int    `json:"expectedSignerCount"`
}

// HandleListMyDocuments handles GET /api/v1/users/me/documents
// Returns documents created by the current authenticated user
func (h *Handler) HandleListMyDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, authenticated := shared.GetUserFromContext(ctx)
	if !authenticated || user == nil {
		shared.WriteError(w, http.StatusUnauthorized, shared.ErrCodeUnauthorized, "Authentication required", nil)
		return
	}

	pagination := shared.ParsePaginationParams(r, 20, 100)
	searchQuery := r.URL.Query().Get("search")

	var docs []*models.Document
	var err error

	if searchQuery != "" {
		docs, err = h.documentService.SearchByCreatedBy(ctx, user.Email, searchQuery, pagination.PageSize, pagination.Offset)
		logger.Logger.Debug("User document search request",
			"user_email", user.Email,
			"query", searchQuery,
			"limit", pagination.PageSize,
			"offset", pagination.Offset)
	} else {
		docs, err = h.documentService.ListByCreatedBy(ctx, user.Email, pagination.PageSize, pagination.Offset)
		logger.Logger.Debug("User document list request",
			"user_email", user.Email,
			"limit", pagination.PageSize,
			"offset", pagination.Offset)
	}

	if err != nil {
		logger.Logger.Error("Failed to fetch user documents",
			"user_email", user.Email,
			"search", searchQuery,
			"error", err.Error())
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to fetch documents", nil)
		return
	}

	totalCount, err := h.documentService.CountByCreatedBy(ctx, user.Email, searchQuery)
	if err != nil {
		logger.Logger.Warn("Failed to count user documents, using result count",
			"error", err.Error(),
			"user_email", user.Email,
			"search", searchQuery)
		totalCount = len(docs)
	}

	documents := make([]MyDocumentDTO, 0, len(docs))
	for _, doc := range docs {
		dto := MyDocumentDTO{
			ID:          doc.DocID,
			Title:       doc.Title,
			URL:         doc.URL,
			Description: doc.Description,
			CreatedAt:   doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if sigs, err := h.signatureService.GetDocumentSignatures(ctx, doc.DocID); err == nil {
			dto.SignatureCount = len(sigs)
		}

		if stats, err := h.documentService.GetExpectedSignerStats(ctx, doc.DocID); err == nil {
			dto.ExpectedSignerCount = stats.ExpectedCount
		}

		documents = append(documents, dto)
	}

	shared.WritePaginatedJSON(w, documents, pagination.Page, pagination.PageSize, totalCount)
}
