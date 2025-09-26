package admin

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
)

// AdminHandlers handles admin-specific HTTP requests
type AdminHandlers struct {
	adminRepo   *database.AdminRepository
	userService userService
	templates   *template.Template
	baseURL     string
}

// NewAdminHandlers creates new admin handlers
func NewAdminHandlers(
	adminRepo *database.AdminRepository,
	userService userService,
	templates *template.Template,
	baseURL string,
) *AdminHandlers {
	return &AdminHandlers{
		adminRepo:   adminRepo,
		userService: userService,
		templates:   templates,
		baseURL:     baseURL,
	}
}

// HandleDashboard handles GET /admin - lists documents with signature counts
func (h *AdminHandlers) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := h.userService.GetUser(r)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	documents, err := h.adminRepo.ListDocumentsWithCounts(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve documents", http.StatusInternalServerError)
		return
	}

	data := struct {
		TemplateName string
		User         *models.User
		BaseURL      string
		Documents    []database.DocumentAgg
		DocID        *string
		IsAdmin      bool
	}{
		TemplateName: "admin_dashboard",
		User:         user,
		BaseURL:      h.baseURL,
		Documents:    documents,
		DocID:        nil,
		IsAdmin:      true, // L'utilisateur est forcément admin pour accéder à cette page
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleDocumentDetails handles GET /admin/docs/{docID} - shows document signataires
func (h *AdminHandlers) HandleDocumentDetails(w http.ResponseWriter, r *http.Request) {
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

	signatures, err := h.adminRepo.ListSignaturesByDoc(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to retrieve signatures", http.StatusInternalServerError)
		return
	}

	// Vérifier l'intégrité de la chaîne pour ce document
	chainIntegrity, err := h.adminRepo.VerifyDocumentChainIntegrity(ctx, docID)
	if err != nil {
		// Log l'erreur mais continue l'affichage
		chainIntegrity = &database.ChainIntegrityResult{
			IsValid:     false,
			TotalSigs:   len(signatures),
			ValidSigs:   0,
			InvalidSigs: len(signatures),
			Errors:      []string{"Failed to verify chain integrity: " + err.Error()},
			DocID:       docID,
		}
	}

	data := struct {
		TemplateName   string
		User           *models.User
		BaseURL        string
		DocID          *string
		Signatures     []*models.Signature
		ChainIntegrity *database.ChainIntegrityResult
		IsAdmin        bool
	}{
		TemplateName:   "admin_doc_details",
		User:           user,
		BaseURL:        h.baseURL,
		DocID:          &docID,
		Signatures:     signatures,
		ChainIntegrity: chainIntegrity,
		IsAdmin:        true, // L'utilisateur est forcément admin pour accéder à cette page
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleChainIntegrityAPI handles GET /admin/api/chain-integrity/{docID} - returns JSON
func (h *AdminHandlers) HandleChainIntegrityAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	result, err := h.adminRepo.VerifyDocumentChainIntegrity(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to verify chain integrity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
