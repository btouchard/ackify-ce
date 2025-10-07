// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"context"
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

type reminderService interface {
	SendReminders(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string) (*models.ReminderSendResult, error)
	GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error)
	GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error)
}

type ExpectedSignersHandlers struct {
	expectedRepo    *database.ExpectedSignerRepository
	adminRepo       *database.AdminRepository
	userService     userService
	reminderService reminderService
	templates       *template.Template
	baseURL         string
}

func NewExpectedSignersHandlers(
	expectedRepo *database.ExpectedSignerRepository,
	adminRepo *database.AdminRepository,
	userService userService,
	reminderService reminderService,
	templates *template.Template,
	baseURL string,
) *ExpectedSignersHandlers {
	return &ExpectedSignersHandlers{
		expectedRepo:    expectedRepo,
		adminRepo:       adminRepo,
		userService:     userService,
		reminderService: reminderService,
		templates:       templates,
		baseURL:         baseURL,
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

	// Get reminder stats
	var reminderStats *models.ReminderStats
	if h.reminderService != nil {
		reminderStats, err = h.reminderService.GetReminderStats(ctx, docID)
		if err != nil {
			logger.Logger.Error("Failed to retrieve reminder stats", "error", err.Error())
			reminderStats = &models.ReminderStats{
				TotalSent:    0,
				PendingCount: stats.PendingCount,
			}
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
		ReminderStats        *models.ReminderStats
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
		ReminderStats:        reminderStats,
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

	contacts := parseContactsFromText(emailsText)

	if len(contacts) == 0 {
		http.Redirect(w, r, "/admin/docs/"+docID, http.StatusSeeOther)
		return
	}

	// Validate emails and build ContactInfo list
	validContacts := []models.ContactInfo{}
	for _, contact := range contacts {
		if isValidEmail(contact.Email) {
			validContacts = append(validContacts, models.ContactInfo{
				Name:  contact.Name,
				Email: contact.Email,
			})
		} else {
			logger.Logger.Warn("Invalid email format", "email", contact.Email)
		}
	}

	if len(validContacts) == 0 {
		http.Error(w, "No valid emails provided", http.StatusBadRequest)
		return
	}

	err = h.expectedRepo.AddExpected(ctx, docID, validContacts, user.Email)
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

// ParsedContact represents a contact with optional name and email
type ParsedContact struct {
	Name  string
	Email string
}

// parseContactsFromText extracts contacts from text supporting formats:
// - "Name <email@example.com>" (with name)
// - "email@example.com" (email only)
func parseContactsFromText(text string) []ParsedContact {
	// Split by newlines first to preserve individual contacts
	lines := strings.Split(text, "\n")

	contacts := []ParsedContact{}

	// Regex for "Name <email>" format
	nameEmailRegex := regexp.MustCompile(`^\s*(.+?)\s*<([^>]+)>\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to match "Name <email>" format
		if matches := nameEmailRegex.FindStringSubmatch(line); len(matches) == 3 {
			name := strings.TrimSpace(matches[1])
			email := strings.TrimSpace(matches[2])
			contacts = append(contacts, ParsedContact{
				Name:  name,
				Email: email,
			})
		} else {
			// Split by commas, semicolons, or spaces for plain emails
			separators := regexp.MustCompile(`[,;\s]+`)
			parts := separators.Split(line, -1)

			for _, part := range parts {
				email := strings.TrimSpace(part)
				if email != "" {
					contacts = append(contacts, ParsedContact{
						Name:  "",
						Email: email,
					})
				}
			}
		}
	}

	return contacts
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

// HandleSendReminders sends reminder emails to pending signers
func (h *ExpectedSignersHandlers) HandleSendReminders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	if h.reminderService == nil {
		http.Error(w, "Reminder service not configured", http.StatusInternalServerError)
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

	sendMode := r.FormValue("send_mode")
	docURL := r.FormValue("doc_url")
	var selectedEmails []string

	if sendMode == "selected" {
		selectedEmails = r.Form["emails"]
		if len(selectedEmails) == 0 {
			http.Error(w, "No emails selected", http.StatusBadRequest)
			return
		}
	}

	result, err := h.reminderService.SendReminders(ctx, docID, user.Email, selectedEmails, docURL)
	if err != nil {
		logger.Logger.Error("Failed to send reminders", "error", err.Error())
		http.Error(w, "Failed to send reminders", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("Reminders sent", "doc_id", docID, "sent_by", user.Email, "total", result.TotalAttempted, "success", result.SuccessfullySent, "failed", result.Failed)

	http.Redirect(w, r, "/admin/docs/"+docID, http.StatusSeeOther)
}

// HandleGetReminderHistory returns reminder history as JSON
func (h *ExpectedSignersHandlers) HandleGetReminderHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docID")

	if docID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}

	if h.reminderService == nil {
		http.Error(w, "Reminder service not configured", http.StatusInternalServerError)
		return
	}

	history, err := h.reminderService.GetReminderHistory(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to get reminder history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
