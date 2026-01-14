// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// OEmbedResponse represents the oEmbed JSON response format
// Specification: https://oembed.com/
type OEmbedResponse struct {
	Type         string `json:"type"`            // Must be "rich" for iframe embeds
	Version      string `json:"version"`         // oEmbed version (always "1.0")
	Title        string `json:"title"`           // Document title
	ProviderName string `json:"provider_name"`   // Service name
	ProviderURL  string `json:"provider_url"`    // Service homepage URL
	HTML         string `json:"html"`            // HTML embed code (iframe)
	Width        int    `json:"width,omitempty"` // Recommended width (optional)
	Height       int    `json:"height"`          // Recommended height
}

// HandleOEmbed handles GET /oembed?url=<document_url>
// Returns oEmbed JSON for embedding Ackify signature widgets in external platforms
func HandleOEmbed(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlParam := r.URL.Query().Get("url")
		if urlParam == "" {
			logger.Logger.Warn("oEmbed request missing url parameter",
				"remote_addr", r.RemoteAddr)
			http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
			return
		}

		parsedURL, err := url.Parse(urlParam)
		if err != nil {
			logger.Logger.Warn("oEmbed request with invalid url",
				"url", urlParam,
				"error", err.Error(),
				"remote_addr", r.RemoteAddr)
			http.Error(w, "Invalid 'url' parameter", http.StatusBadRequest)
			return
		}

		// Extract doc ID from query parameters
		docID := parsedURL.Query().Get("doc")
		if docID == "" {
			logger.Logger.Warn("oEmbed request missing doc parameter in url",
				"url", urlParam,
				"remote_addr", r.RemoteAddr)
			http.Error(w, "URL must contain 'doc' parameter", http.StatusBadRequest)
			return
		}

		embedURL := baseURL + "/embed?doc=" + url.QueryEscape(docID)

		referrer := parsedURL.Query().Get("referrer")
		if referrer != "" {
			embedURL += "&referrer=" + url.QueryEscape(referrer)
		}

		iframeHTML := `<iframe src="` + embedURL + `" width="100%" height="200" frameborder="0" style="border: 1px solid #ddd; border-radius: 6px;" allowtransparency="true"></iframe>`

		response := OEmbedResponse{
			Type:         "rich",
			Version:      "1.0",
			Title:        "Document " + docID + " - Confirmations de lecture",
			ProviderName: "Ackify",
			ProviderURL:  baseURL,
			HTML:         iframeHTML,
			Height:       200,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Logger.Error("Failed to encode oEmbed response",
				"doc_id", docID,
				"error", err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		logger.Logger.Info("oEmbed response served",
			"doc_id", docID,
			"url", urlParam,
			"remote_addr", r.RemoteAddr)
	}
}

// ValidateOEmbedURL checks if the provided URL is a valid Ackify document URL
func ValidateOEmbedURL(urlStr string, baseURL string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	// Normalize hosts for comparison (remove ports if present)
	urlHost := strings.Split(parsedURL.Host, ":")[0]
	baseHost := strings.Split(baseURLParsed.Host, ":")[0]

	// Allow localhost variations
	if urlHost == "localhost" || urlHost == "127.0.0.1" {
		if baseHost == "localhost" || baseHost == "127.0.0.1" {
			return true
		}
	}

	return urlHost == baseHost
}
