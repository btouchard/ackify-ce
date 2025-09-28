// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	// Allow safe doc identifiers: letters, digits, dot, underscore, colon, hyphen; max 128
	reDocID = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)
	// User identifier (sub or email): any non-whitespace chars, 1..254
	reUserIdentifier = regexp.MustCompile(`^[^\s]{1,254}$`)
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

	if !reDocID.MatchString(docID) {
		return "", fmt.Errorf("invalid document ID format")
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
	if !reUserIdentifier.MatchString(userIdentifier) {
		return "", fmt.Errorf("invalid user parameter")
	}
	return userIdentifier, nil
}
