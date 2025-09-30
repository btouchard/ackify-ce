// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/btouchard/ackify-ce/internal/infrastructure/i18n"
	"github.com/btouchard/ackify-ce/pkg/logger"
	"github.com/go-chi/chi/v5"
)

type LangHandlers struct {
	secureCookies bool
}

func NewLangHandlers(secureCookies bool) *LangHandlers {
	return &LangHandlers{
		secureCookies: secureCookies,
	}
}

// HandleLangSwitch changes the user's language preference
func (h *LangHandlers) HandleLangSwitch(w http.ResponseWriter, r *http.Request) {
	lang := chi.URLParam(r, "code")

	// Set language cookie
	i18n.SetLangCookie(w, lang, h.secureCookies)

	// Get redirect URL from query parameter first, then referer
	redirectTo := r.URL.Query().Get("redirect")

	if redirectTo == "" {
		// Try to get referer
		referer := r.Header.Get("Referer")
		if referer != "" {
			// Parse referer to get just the path
			if refererURL, err := url.Parse(referer); err == nil {
				// Only use path + query, ignore host to prevent open redirect
				redirectTo = refererURL.Path
				if refererURL.RawQuery != "" {
					redirectTo += "?" + refererURL.RawQuery
				}
			}
		}
	}

	// Default to home if no valid redirect
	if redirectTo == "" || redirectTo == "/lang/fr" || redirectTo == "/lang/en" || strings.HasPrefix(redirectTo, "/lang/") {
		redirectTo = "/"
	}

	logger.Logger.Debug("Language switch", "lang", lang, "redirect", redirectTo)

	http.Redirect(w, r, redirectTo, http.StatusFound)
}
