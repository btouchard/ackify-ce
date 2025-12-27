// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// SignatureRepository defines minimal signature operations for meta tags
type SignatureRepository interface {
	GetByDoc(ctx context.Context, docID string) ([]*models.Signature, error)
}

// EmbedFolder returns an http.HandlerFunc that serves an embedded filesystem
// with SPA fallback support (serves index.html for non-existent routes)
// For index.html, it replaces __ACKIFY_BASE_URL__ placeholder with the actual base URL,
// __ACKIFY_VERSION__ with the application version,
// __ACKIFY_OAUTH_ENABLED__ and __ACKIFY_MAGICLINK_ENABLED__ with auth method flags,
// __ACKIFY_SMTP_ENABLED__ with SMTP availability flag,
// __ACKIFY_ONLY_ADMIN_CAN_CREATE__ with document creation restriction flag,
// and __META_TAGS__ with dynamic meta tags based on query parameters
func EmbedFolder(fsEmbed embed.FS, targetPath string, baseURL string, version string, oauthEnabled bool, magicLinkEnabled bool, smtpEnabled bool, onlyAdminCanCreate bool, signatureRepo SignatureRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fsys, err := fs.Sub(fsEmbed, targetPath)
		if err != nil {
			logger.Logger.Error("Failed to load embedded files",
				"target_path", targetPath,
				"error", err.Error())
			http.Error(w, "Failed to load embedded files", http.StatusInternalServerError)
			return
		}

		urlPath := r.URL.Path

		cleanPath := path.Clean(urlPath)
		shouldServeIndex := false

		if cleanPath == "/" {
			cleanPath = "index.html"
			shouldServeIndex = true
		} else {
			cleanPath = cleanPath[1:]
		}

		file, err := fsys.Open(cleanPath)
		if err != nil {
			logger.Logger.Debug("SPA fallback: file not found, serving index.html",
				"requested_path", urlPath,
				"clean_path", cleanPath)
			cleanPath = "index.html"
			shouldServeIndex = true
			file, err = fsys.Open(cleanPath)
			if err != nil {
				http.Error(w, "index.html not found", http.StatusInternalServerError)
				return
			}
		}
		defer file.Close()

		if shouldServeIndex || strings.HasSuffix(cleanPath, "index.html") {
			serveIndexTemplate(w, r, file, baseURL, version, oauthEnabled, magicLinkEnabled, smtpEnabled, onlyAdminCanCreate, signatureRepo)
			return
		}

		fileServer := http.FileServer(http.FS(fsys))
		fileServer.ServeHTTP(w, r)
	}
}

func serveIndexTemplate(w http.ResponseWriter, r *http.Request, file fs.File, baseURL string, version string, oauthEnabled bool, magicLinkEnabled bool, smtpEnabled bool, onlyAdminCanCreate bool, signatureRepo SignatureRepository) {
	content, err := io.ReadAll(file)
	if err != nil {
		logger.Logger.Error("Failed to read index.html", "error", err.Error())
		http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
		return
	}

	processedContent := strings.ReplaceAll(string(content), "__ACKIFY_BASE_URL__", baseURL)
	processedContent = strings.ReplaceAll(processedContent, "__ACKIFY_VERSION__", version)

	oauthEnabledStr := "false"
	if oauthEnabled {
		oauthEnabledStr = "true"
	}
	magicLinkEnabledStr := "false"
	if magicLinkEnabled {
		magicLinkEnabledStr = "true"
	}
	smtpEnabledStr := "false"
	if smtpEnabled {
		smtpEnabledStr = "true"
	}
	onlyAdminCanCreateStr := "false"
	if onlyAdminCanCreate {
		onlyAdminCanCreateStr = "true"
	}

	processedContent = strings.ReplaceAll(processedContent, "__ACKIFY_OAUTH_ENABLED__", oauthEnabledStr)
	processedContent = strings.ReplaceAll(processedContent, "__ACKIFY_MAGICLINK_ENABLED__", magicLinkEnabledStr)
	processedContent = strings.ReplaceAll(processedContent, "__ACKIFY_SMTP_ENABLED__", smtpEnabledStr)
	processedContent = strings.ReplaceAll(processedContent, "__ACKIFY_ONLY_ADMIN_CAN_CREATE__", onlyAdminCanCreateStr)

	metaTags := generateMetaTags(r, baseURL, signatureRepo)
	processedContent = strings.ReplaceAll(processedContent, "__META_TAGS__", metaTags)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, bytes.NewBufferString(processedContent)); err != nil {
		logger.Logger.Error("Failed to write response", "error", err.Error())
	}
}

func generateMetaTags(r *http.Request, baseURL string, signatureRepo SignatureRepository) string {
	docID := r.URL.Query().Get("doc")
	if docID == "" {
		return ""
	}

	ctx := context.Background()
	signatures, err := signatureRepo.GetByDoc(ctx, docID)
	if err != nil {
		logger.Logger.Warn("Failed to fetch signatures for meta tags", "doc_id", docID, "error", err.Error())
		return generateBasicMetaTags(docID, baseURL, 0)
	}

	signatureCount := len(signatures)
	return generateBasicMetaTags(docID, baseURL, signatureCount)
}

func generateBasicMetaTags(docID string, baseURL string, signatureCount int) string {
	escapedDocID := html.EscapeString(docID)
	currentURL := fmt.Sprintf("%s/?doc=%s", baseURL, docID)
	escapedURL := html.EscapeString(currentURL)

	var title, description string
	if signatureCount == 0 {
		title = fmt.Sprintf("Document: %s - Aucune confirmation", escapedDocID)
		description = fmt.Sprintf("Confirmations de lecture pour le document %s", escapedDocID)
	} else if signatureCount == 1 {
		title = fmt.Sprintf("Document: %s - 1 confirmation", escapedDocID)
		description = fmt.Sprintf("1 personne a confirmé avoir lu le document %s", escapedDocID)
	} else {
		title = fmt.Sprintf("Document: %s - %d confirmations", escapedDocID, signatureCount)
		description = fmt.Sprintf("%d personnes ont confirmé avoir lu le document %s", signatureCount, escapedDocID)
	}

	var metaTags strings.Builder
	metaTags.WriteString(fmt.Sprintf(`<meta property="og:title" content="%s" />`, html.EscapeString(title)))
	metaTags.WriteString("\n    ")
	metaTags.WriteString(fmt.Sprintf(`<meta property="og:description" content="%s" />`, html.EscapeString(description)))
	metaTags.WriteString("\n    ")
	metaTags.WriteString(fmt.Sprintf(`<meta property="og:url" content="%s" />`, escapedURL))
	metaTags.WriteString("\n    ")
	metaTags.WriteString(`<meta property="og:type" content="website" />`)
	metaTags.WriteString("\n    ")

	// Twitter Card tags
	metaTags.WriteString(`<meta name="twitter:card" content="summary" />`)
	metaTags.WriteString("\n    ")
	metaTags.WriteString(fmt.Sprintf(`<meta name="twitter:title" content="%s" />`, html.EscapeString(title)))
	metaTags.WriteString("\n    ")
	metaTags.WriteString(fmt.Sprintf(`<meta name="twitter:description" content="%s" />`, html.EscapeString(description)))
	metaTags.WriteString("\n    ")

	// oEmbed discovery tag
	oembedURL := fmt.Sprintf("%s/oembed?url=%s", baseURL, escapedURL)
	metaTags.WriteString(fmt.Sprintf(`<link rel="alternate" type="application/json+oembed" href="%s" title="%s" />`,
		html.EscapeString(oembedURL),
		html.EscapeString(title)))

	return metaTags.String()
}
