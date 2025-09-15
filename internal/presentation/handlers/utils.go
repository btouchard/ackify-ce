package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func validateDocID(r *http.Request) (string, error) {
	var docID string

	docID = strings.TrimSpace(r.URL.Query().Get("doc"))
	if docID == "" {
		docID = strings.TrimSpace(r.FormValue("doc"))
	}

	if docID == "" {
		return "", fmt.Errorf("missing document ID")
	}

	return docID, nil
}

func buildSignURL(baseURL, docID string) string {
	return fmt.Sprintf("%s/sign?doc=%s", baseURL, url.QueryEscape(docID))
}

func buildLoginURL(nextURL string) string {
	return "/login?next=" + url.QueryEscape(nextURL)
}

func validateUserIdentifier(r *http.Request) (string, error) {
	userIdentifier := strings.TrimSpace(r.URL.Query().Get("user"))
	if userIdentifier == "" {
		return "", fmt.Errorf("missing user parameter")
	}
	return userIdentifier, nil
}
