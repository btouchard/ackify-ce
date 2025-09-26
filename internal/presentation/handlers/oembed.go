package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

// OEmbedHandler handles oEmbed requests
type OEmbedHandler struct {
	signatureService signatureService
	template         *template.Template
	baseURL          string
	organisation     string
}

// NewOEmbedHandler creates a new oEmbed handler
func NewOEmbedHandler(signatureService signatureService, tmpl *template.Template, baseURL, organisation string) *OEmbedHandler {
	return &OEmbedHandler{
		signatureService: signatureService,
		template:         tmpl,
		baseURL:          baseURL,
		organisation:     organisation,
	}
}

// OEmbedResponse represents the oEmbed JSON response format
type OEmbedResponse struct {
	Type         string `json:"type"`
	Version      string `json:"version"`
	Title        string `json:"title"`
	AuthorName   string `json:"author_name,omitempty"`
	AuthorURL    string `json:"author_url,omitempty"`
	ProviderName string `json:"provider_name"`
	ProviderURL  string `json:"provider_url"`
	CacheAge     int    `json:"cache_age,omitempty"`
	HTML         string `json:"html"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
}

// SignatoryData represents data for rendering signatories
type SignatoryData struct {
	DocID        string
	Signatures   []SignatoryInfo
	Count        int
	LastSignedAt string
	EmbedURL     string
	SignURL      string
}

// SignatoryInfo represents a signatory's information
type SignatoryInfo struct {
	Name     string
	Email    string
	SignedAt string
}

// HandleOEmbed handles oEmbed requests for signature lists
func (h *OEmbedHandler) HandleOEmbed(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	format := r.URL.Query().Get("format")
	maxWidth := r.URL.Query().Get("maxwidth")
	maxHeight := r.URL.Query().Get("maxheight")

	if targetURL == "" {
		HandleError(w, models.ErrInvalidDocument)
		return
	}

	if format == "" {
		format = "json"
	}

	if format != "json" {
		http.Error(w, "Only JSON format is supported", http.StatusNotImplemented)
		return
	}

	docID, err := h.extractDocIDFromURL(targetURL)
	if err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	signatures, err := h.signatureService.GetDocumentSignatures(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to retrieve signatures", http.StatusInternalServerError)
		return
	}

	// Convert to signatory info
	signatories := make([]SignatoryInfo, len(signatures))
	var lastSignedAt string
	for i, sig := range signatures {
		name := ""
		if sig.UserName != nil {
			name = *sig.UserName
		}
		signatories[i] = SignatoryInfo{
			Name:     name,
			Email:    sig.UserEmail,
			SignedAt: sig.SignedAtUTC.Format("02/01/2006 à 15:04"),
		}
		if i == 0 { // First signature (most recent due to ORDER BY in repository)
			lastSignedAt = signatories[i].SignedAt
		}
	}

	embedHTML, err := h.renderEmbeddedHTML(SignatoryData{
		DocID:        docID,
		Signatures:   signatories,
		Count:        len(signatories),
		LastSignedAt: lastSignedAt,
		EmbedURL:     targetURL,
		SignURL:      fmt.Sprintf("%s/sign?doc=%s", h.baseURL, url.QueryEscape(docID)),
	})
	if err != nil {
		http.Error(w, "Failed to render embedded content", http.StatusInternalServerError)
		return
	}

	width := 480  // Default width
	height := 320 // Default height

	if maxWidth != "" {
		if w, err := strconv.Atoi(maxWidth); err == nil && w > 0 && w < 2000 {
			width = w
		}
	}

	if maxHeight != "" {
		if h, err := strconv.Atoi(maxHeight); err == nil && h > 0 && h < 2000 {
			height = h
		}
	}

	// Create oEmbed response
	response := OEmbedResponse{
		Type:         "rich",
		Version:      "1.0",
		Title:        fmt.Sprintf("Signataires du document %s", docID),
		AuthorName:   h.organisation,
		AuthorURL:    h.baseURL,
		ProviderName: "Service de validation de lecture",
		ProviderURL:  h.baseURL,
		CacheAge:     3600, // Cache for 1 hour
		HTML:         embedHTML,
		Width:        width,
		Height:       height,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleEmbedView handles direct embed view requests
func (h *OEmbedHandler) HandleEmbedView(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimSpace(r.URL.Query().Get("doc"))
	if docID == "" {
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	signatures, err := h.signatureService.GetDocumentSignatures(ctx, docID)
	if err != nil {
		http.Error(w, "Failed to retrieve signatures", http.StatusInternalServerError)
		return
	}

	// Convert to signatory info
	signatories := make([]SignatoryInfo, len(signatures))
	var lastSignedAt string
	for i, sig := range signatures {
		name := ""
		if sig.UserName != nil {
			name = *sig.UserName
		}
		signatories[i] = SignatoryInfo{
			Name:     name,
			Email:    sig.UserEmail,
			SignedAt: sig.SignedAtUTC.Format("02/01/2006 à 15:04"),
		}
		if i == 0 {
			lastSignedAt = signatories[i].SignedAt
		}
	}

	data := SignatoryData{
		DocID:        docID,
		Signatures:   signatories,
		Count:        len(signatories),
		LastSignedAt: lastSignedAt,
		EmbedURL:     fmt.Sprintf("%s/embed?doc=%s", h.baseURL, url.QueryEscape(docID)),
		SignURL:      fmt.Sprintf("%s/sign?doc=%s", h.baseURL, url.QueryEscape(docID)),
	}

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Header().Set("X-Frame-Options", "ALLOWALL") // Allow embedding in iframes
    // Override default CSP to allow framing from any parent (widget use-case)
    w.Header().Set("Content-Security-Policy",
        "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; "+
            "script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; "+
            "img-src 'self' data: https://cdn.simpleicons.org; connect-src 'self'; "+
            "frame-ancestors *")

	if err := h.template.ExecuteTemplate(w, "embed", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// extractDocIDFromURL extracts document ID from various URL formats
func (h *OEmbedHandler) extractDocIDFromURL(targetURL string) (string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}

	if docID := parsedURL.Query().Get("doc"); docID != "" {
		return docID, nil
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) >= 2 && (pathParts[0] == "embed" || pathParts[0] == "status" || pathParts[0] == "sign") {
		return pathParts[1], nil
	}

	return "", fmt.Errorf("could not extract document ID from URL")
}

// renderEmbeddedHTML renders the embedded HTML content
func (h *OEmbedHandler) renderEmbeddedHTML(data SignatoryData) (string, error) {
	var buf strings.Builder
	if err := h.template.ExecuteTemplate(&buf, "embed", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
