// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

type DocumentHandlers struct {
	documentRepo *database.DocumentRepository
	userService  userService
}

func NewDocumentHandlers(
	documentRepo *database.DocumentRepository,
	userService userService,
) *DocumentHandlers {
	return &DocumentHandlers{
		documentRepo: documentRepo,
		userService:  userService,
	}
}

// HandleGetDocumentMetadata retrieves document metadata as JSON
func (h *DocumentHandlers) HandleGetDocumentMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	doc, err := h.documentRepo.GetByDocID(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to get document metadata", "error", err.Error(), "doc_id", docID)
		http.Error(w, "Failed to get document metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleUpdateDocumentMetadata creates or updates document metadata
func (h *DocumentHandlers) HandleUpdateDocumentMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(r)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	input := models.DocumentInput{
		Title:             r.FormValue("title"),
		URL:               r.FormValue("url"),
		Checksum:          r.FormValue("checksum"),
		ChecksumAlgorithm: r.FormValue("checksum_algorithm"),
		Description:       r.FormValue("description"),
	}

	// Validate checksum algorithm
	validAlgorithms := map[string]bool{
		"SHA-256": true,
		"SHA-512": true,
		"MD5":     true,
	}

	if input.ChecksumAlgorithm != "" && !validAlgorithms[input.ChecksumAlgorithm] {
		http.Error(w, "Invalid checksum algorithm. Must be SHA-256, SHA-512, or MD5", http.StatusBadRequest)
		return
	}

	// Default to SHA-256 if not specified
	if input.ChecksumAlgorithm == "" {
		input.ChecksumAlgorithm = "SHA-256"
	}

	doc, err := h.documentRepo.CreateOrUpdate(ctx, docID, input, user.Email)
	if err != nil {
		logger.Logger.Error("Failed to update document metadata", "error", err.Error(), "doc_id", docID)
		http.Error(w, "Failed to update document metadata", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("Document metadata updated", "doc_id", docID, "updated_by", user.Email)

	// Return JSON response for AJAX requests
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Redirect back to document page for form submissions
	http.Redirect(w, r, "/admin/docs/"+docID, http.StatusSeeOther)
}

// HandleDeleteDocumentMetadata deletes document metadata
func (h *DocumentHandlers) HandleDeleteDocumentMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(r)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.documentRepo.Delete(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to delete document metadata", "error", err.Error(), "doc_id", docID)
		http.Error(w, "Failed to delete document metadata", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("Document metadata deleted", "doc_id", docID, "deleted_by", user.Email)

	// Return success for AJAX requests
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		return
	}

	// Redirect back to document page for form submissions
	http.Redirect(w, r, "/admin/docs/"+docID, http.StatusSeeOther)
}
