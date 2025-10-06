// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/internal/infrastructure/i18n"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

const maxTextareaSize = 10000

type ExpectedSignersHandlers struct {
	expectedRepo *database.ExpectedSignerRepository
	adminRepo    *database.AdminRepository
	userService  userService
	templates    *template.Template
	baseURL      string
}

func NewExpectedSignersHandlers(
	expectedRepo *database.ExpectedSignerRepository,
	adminRepo *database.AdminRepository,
	userService userService,
	templates *template.Template,
	baseURL string,
) *ExpectedSignersHandlers {
	return &ExpectedSignersHandlers{
		expectedRepo: expectedRepo,
		adminRepo:    adminRepo,
		userService:  userService,
		templates:    templates,
		baseURL:      baseURL,
	}
}

// HandleDocumentDetailsWithExpected displays document details with expected signers
func (h *ExpectedSignersHandlers) HandleDocumentDetailsWithExpected(w http.ResponseWriter, r *http.Request) {
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

	// Get signatures
	signatures, err := h.adminRepo.ListSignaturesByDoc(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to retrieve signatures", http.StatusInternalServerError)
		return
	}

	// Get expected signers with status
	expectedSigners, err := h.expectedRepo.ListWithStatusByDocID(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to retrieve expected signers", "error", err.Error())
		expectedSigners = []*models.ExpectedSignerWithStatus{}
	}

	// Get stats
	stats, err := h.expectedRepo.GetStats(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to retrieve stats", "error", err.Error())
		stats = &models.DocCompletionStats{
			DocID:          docID,
			ExpectedCount:  0,
			SignedCount:    0,
			PendingCount:   0,
			CompletionRate: 0,
		}
	}

	// Check chain integrity
	chainIntegrity, err := h.adminRepo.VerifyDocumentChainIntegrity(ctx, docID)
	if err != nil {
		chainIntegrity = &database.ChainIntegrityResult{
			IsValid:     false,
			TotalSigs:   len(signatures),
			ValidSigs:   0,
			InvalidSigs: len(signatures),
			Errors:      []string{"Failed to verify chain integrity: " + err.Error()},
			DocID:       docID,
		}
	}

	// Find unexpected signatures (signed but not in expected list)
	unexpectedSignatures := []*models.Signature{}
	if len(expectedSigners) > 0 {
		expectedEmails := make(map[string]bool)
		for _, es := range expectedSigners {
			expectedEmails[es.Email] = true
		}

		for _, sig := range signatures {
			if !expectedEmails[sig.UserEmail] {
				unexpectedSignatures = append(unexpectedSignatures, sig)
			}
		}
	}

	data := struct {
		TemplateName         string
		User                 *models.User
		BaseURL              string
		DocID                *string
		Signatures           []*models.Signature
		ExpectedSigners      []*models.ExpectedSignerWithStatus
		Stats                *models.DocCompletionStats
		UnexpectedSignatures []*models.Signature
		ChainIntegrity       *database.ChainIntegrityResult
		ShareLink            string
		IsAdmin              bool
		Lang                 string
		T                    map[string]string
	}{
		TemplateName:         "admin_document_expected_signers",
		User:                 user,
		BaseURL:              h.baseURL,
		DocID:                &docID,
		Signatures:           signatures,
		ExpectedSigners:      expectedSigners,
		Stats:                stats,
		UnexpectedSignatures: unexpectedSignatures,
		ChainIntegrity:       chainIntegrity,
		ShareLink:            h.baseURL + "/sign?doc=" + docID,
		IsAdmin:              true,
		Lang:                 i18n.GetLang(ctx),
		T:                    i18n.GetTranslations(ctx),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleAddExpectedSigners adds expected signers to a document
func (h *ExpectedSignersHandlers) HandleAddExpectedSigners(w http.ResponseWriter, r *http.Request) {
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

	emailsText := r.FormValue("emails")
	if len(emailsText) > maxTextareaSize {
		http.Error(w, "Input too large", http.StatusBadRequest)
		return
	}

	emails := parseEmailsFromText(emailsText)

	if len(emails) == 0 {
		http.Redirect(w, r, "/admin/docs/"+docID, http.StatusSeeOther)
		return
	}

	// Validate emails
	validEmails := []string{}
	for _, email := range emails {
		if isValidEmail(email) {
			validEmails = append(validEmails, email)
		} else {
			logger.Logger.Warn("Invalid email format", "email", email)
		}
	}

	if len(validEmails) == 0 {
		http.Error(w, "No valid emails provided", http.StatusBadRequest)
		return
	}

	err = h.expectedRepo.AddExpected(ctx, docID, validEmails, user.Email)
	if err != nil {
		logger.Logger.Error("Failed to add expected signers", "error", err.Error())
		http.Error(w, "Failed to add expected signers", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/docs/"+docID, http.StatusSeeOther)
}

// HandleRemoveExpectedSigner removes an expected signer from a document
func (h *ExpectedSignersHandlers) HandleRemoveExpectedSigner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}

	err := h.expectedRepo.Remove(ctx, docID, email)
	if err != nil {
		logger.Logger.Error("Failed to remove expected signer", "error", err.Error())
		http.Error(w, "Failed to remove expected signer", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/docs/"+docID, http.StatusSeeOther)
}

// HandleGetDocumentStatusJSON returns document status as JSON for AJAX requests
func (h *ExpectedSignersHandlers) HandleGetDocumentStatusJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	stats, err := h.expectedRepo.GetStats(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	signers, err := h.expectedRepo.ListWithStatusByDocID(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to get signers", http.StatusInternalServerError)
		return
	}

	response := struct {
		Stats   *models.DocCompletionStats         `json:"stats"`
		Signers []*models.ExpectedSignerWithStatus `json:"signers"`
	}{
		Stats:   stats,
		Signers: signers,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// parseEmailsFromText extracts emails from text (separated by newlines, commas, semicolons)
func parseEmailsFromText(text string) []string {
	// Split by multiple separators: newline, comma, semicolon, space
	separators := regexp.MustCompile(`[\n,;\s]+`)
	parts := separators.Split(text, -1)

	emails := []string{}
	for _, part := range parts {
		email := strings.TrimSpace(part)
		if email != "" {
			emails = append(emails, email)
		}
	}

	return emails
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// Basic regex: has @ and . after @
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return emailRegex.MatchString(email)
}
