package admin

import (
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
	}{
		TemplateName: "admin_dashboard",
		User:         user,
		BaseURL:      h.baseURL,
		Documents:    documents,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
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

	data := struct {
		TemplateName string
		User         *models.User
		BaseURL      string
		DocID        string
		Signatures   []*models.Signature
	}{
		TemplateName: "admin_doc_details",
		User:         user,
		BaseURL:      h.baseURL,
		DocID:        docID,
		Signatures:   signatures,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}
