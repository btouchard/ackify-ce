// SPDX-License-Identifier: AGPL-3.0-or-later
package signatures

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/go-chi/chi/v5"
)

// signatureService defines the interface for signature operations
type signatureService interface {
	CreateSignature(ctx context.Context, request *models.SignatureRequest) error
	GetSignatureStatus(ctx context.Context, docID string, user *models.User) (*models.SignatureStatus, error)
	GetSignatureByDocAndUser(ctx context.Context, docID string, user *models.User) (*models.Signature, error)
	GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error)
	GetUserSignatures(ctx context.Context, user *models.User) ([]*models.Signature, error)
}

// Handler handles signature-related requests
type Handler struct {
	signatureService signatureService
}

// NewHandler creates a new signature handler
func NewHandler(signatureService signatureService) *Handler {
	return &Handler{
		signatureService: signatureService,
	}
}

// CreateSignatureRequest represents the request body for creating a signature
type CreateSignatureRequest struct {
	DocID   string  `json:"docId"`
	Referer *string `json:"referer,omitempty"`
}

// SignatureResponse represents a signature in API responses
type SignatureResponse struct {
	ID           int64              `json:"id"`
	DocID        string             `json:"docId"`
	UserSub      string             `json:"userSub"`
	UserEmail    string             `json:"userEmail"`
	UserName     string             `json:"userName,omitempty"`
	SignedAt     string             `json:"signedAt"`
	PayloadHash  string             `json:"payloadHash"`
	Signature    string             `json:"signature"`
	Nonce        string             `json:"nonce"`
	CreatedAt    string             `json:"createdAt"`
	Referer      *string            `json:"referer,omitempty"`
	PrevHash     *string            `json:"prevHash,omitempty"`
	ServiceInfo  *ServiceInfoResult `json:"serviceInfo,omitempty"`
	DocDeletedAt *string            `json:"docDeletedAt,omitempty"`
	// Document metadata
	DocTitle *string `json:"docTitle,omitempty"`
	DocUrl   *string `json:"docUrl,omitempty"`
}

// ServiceInfoResult represents service detection information
type ServiceInfoResult struct {
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Type     string `json:"type"`
	Referrer string `json:"referrer"`
}

// SignatureStatusResponse represents the signature status for a document
type SignatureStatusResponse struct {
	DocID     string  `json:"docId"`
	UserEmail string  `json:"userEmail"`
	IsSigned  bool    `json:"isSigned"`
	SignedAt  *string `json:"signedAt,omitempty"`
}

// HandleCreateSignature handles POST /api/v1/signatures
func (h *Handler) HandleCreateSignature(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (set by RequireAuth middleware)
	user, ok := shared.GetUserFromContext(ctx)
	if !ok || user == nil {
		shared.WriteUnauthorized(w, "Authentication required")
		return
	}

	// Parse request body
	var req CreateSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Validate document ID
	if req.DocID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Create signature request
	sigRequest := &models.SignatureRequest{
		DocID:   req.DocID,
		User:    user,
		Referer: req.Referer,
	}

	// Create signature
	err := h.signatureService.CreateSignature(ctx, sigRequest)
	if err != nil {
		if err == models.ErrSignatureAlreadyExists {
			shared.WriteConflict(w, "You have already signed this document")
			return
		}

		if err == models.ErrInvalidDocument {
			shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid document", nil)
			return
		}

		if err == models.ErrDocumentModified {
			shared.WriteError(w, http.StatusConflict, "DOCUMENT_MODIFIED", "The document has been modified since it was created. Please verify the current version before signing.", map[string]interface{}{
				"docId": req.DocID,
			})
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to create signature", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get the created signature to return it
	signature, err := h.signatureService.GetSignatureByDocAndUser(ctx, req.DocID, user)
	if err != nil {
		// Signature was created but we couldn't retrieve it
		shared.WriteJSON(w, http.StatusCreated, map[string]interface{}{
			"message": "Signature created successfully",
			"docId":   req.DocID,
		})
		return
	}

	// Return the created signature
	shared.WriteJSON(w, http.StatusCreated, h.toSignatureResponse(ctx, signature))
}

// HandleGetUserSignatures handles GET /api/v1/signatures
func (h *Handler) HandleGetUserSignatures(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context
	user, ok := shared.GetUserFromContext(ctx)
	if !ok || user == nil {
		shared.WriteUnauthorized(w, "Authentication required")
		return
	}

	// Get user's signatures
	signatures, err := h.signatureService.GetUserSignatures(ctx, user)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to fetch signatures", map[string]interface{}{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]*SignatureResponse, 0, len(signatures))
	for _, sig := range signatures {
		response = append(response, h.toSignatureResponse(ctx, sig))
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

// HandleGetDocumentSignatures handles GET /api/v1/documents/{docId}/signatures
func (h *Handler) HandleGetDocumentSignatures(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get document ID from URL
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Get document signatures
	signatures, err := h.signatureService.GetDocumentSignatures(ctx, docID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to fetch signatures", map[string]interface{}{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]*SignatureResponse, 0, len(signatures))
	for _, sig := range signatures {
		response = append(response, h.toSignatureResponse(ctx, sig))
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

// HandleGetSignatureStatus handles GET /api/v1/documents/{docId}/signatures/status
func (h *Handler) HandleGetSignatureStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context
	user, ok := shared.GetUserFromContext(ctx)
	if !ok || user == nil {
		shared.WriteUnauthorized(w, "Authentication required")
		return
	}

	// Get document ID from URL
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Document ID is required", nil)
		return
	}

	// Get signature status
	status, err := h.signatureService.GetSignatureStatus(ctx, docID, user)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrCodeInternal, "Failed to fetch signature status", map[string]interface{}{"error": err.Error()})
		return
	}

	// Convert to response format
	response := SignatureStatusResponse{
		DocID:     status.DocID,
		UserEmail: status.UserEmail,
		IsSigned:  status.IsSigned,
	}

	if status.SignedAt != nil {
		signedAt := status.SignedAt.Format("2006-01-02T15:04:05Z07:00")
		response.SignedAt = &signedAt
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

// toSignatureResponse converts a domain signature to API response format
func (h *Handler) toSignatureResponse(ctx context.Context, sig *models.Signature) *SignatureResponse {
	response := &SignatureResponse{
		ID:          sig.ID,
		DocID:       sig.DocID,
		UserSub:     sig.UserSub,
		UserEmail:   sig.UserEmail,
		UserName:    sig.UserName,
		SignedAt:    sig.SignedAtUTC.Format("2006-01-02T15:04:05Z07:00"),
		PayloadHash: sig.PayloadHash,
		Signature:   sig.Signature,
		Nonce:       sig.Nonce,
		CreatedAt:   sig.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Referer:     sig.Referer,
		PrevHash:    sig.PrevHash,
	}

	// Add doc_deleted_at if document was deleted
	if sig.DocDeletedAt != nil {
		deletedAt := sig.DocDeletedAt.Format("2006-01-02T15:04:05Z07:00")
		response.DocDeletedAt = &deletedAt
	}

	// Add service info if available
	if serviceInfo := sig.GetServiceInfo(); serviceInfo != nil {
		response.ServiceInfo = &ServiceInfoResult{
			Name:     serviceInfo.Name,
			Icon:     serviceInfo.Icon,
			Type:     serviceInfo.Type,
			Referrer: serviceInfo.Referrer,
		}
	}

	// Document metadata is enriched from LEFT JOIN in repository
	if sig.DocTitle != "" {
		response.DocTitle = &sig.DocTitle
	}
	if sig.DocURL != "" {
		response.DocUrl = &sig.DocURL
	}

	return response
}
