package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// validateDocID extracts and validates document ID from request
func validateDocID(r *http.Request) (string, error) {
	var docID string

	// Try query parameter first, then form value
	docID = strings.TrimSpace(r.URL.Query().Get("doc"))
	if docID == "" {
		docID = strings.TrimSpace(r.FormValue("doc"))
	}

	if docID == "" {
		return "", fmt.Errorf("missing document ID")
	}

	return docID, nil
}

// buildSignURL constructs a sign URL with proper escaping
func buildSignURL(baseURL, docID string) string {
	return fmt.Sprintf("%s/sign?doc=%s", baseURL, url.QueryEscape(docID))
}

// buildLoginURL constructs a login URL with next parameter
func buildLoginURL(nextURL string) string {
	return "/login?next=" + url.QueryEscape(nextURL)
}

// validateUserIdentifier extracts and validates user identifier from request
func validateUserIdentifier(r *http.Request) (string, error) {
	userIdentifier := strings.TrimSpace(r.URL.Query().Get("user"))
	if userIdentifier == "" {
		return "", fmt.Errorf("missing user parameter")
	}
	return userIdentifier, nil
}
