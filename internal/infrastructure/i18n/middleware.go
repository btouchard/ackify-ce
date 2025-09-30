// SPDX-License-Identifier: AGPL-3.0-or-later
package i18n

import (
	"context"
	"net/http"
)

type contextKey string

const (
	langContextKey  = contextKey("lang")
	i18nContextKey  = contextKey("i18n")
	transContextKey = contextKey("translations")
)

// Middleware injects language and i18n service into request context
func Middleware(i18n *I18n) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := GetLangFromRequest(r)

			// Inject language and i18n service into context
			ctx := context.WithValue(r.Context(), langContextKey, lang)
			ctx = context.WithValue(ctx, i18nContextKey, i18n)
			ctx = context.WithValue(ctx, transContextKey, i18n.GetTranslations(lang))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLang extracts language from context
func GetLang(ctx context.Context) string {
	if lang, ok := ctx.Value(langContextKey).(string); ok {
		return lang
	}
	return DefaultLang
}

// GetI18n extracts i18n service from context
func GetI18n(ctx context.Context) *I18n {
	if i18n, ok := ctx.Value(i18nContextKey).(*I18n); ok {
		return i18n
	}
	return nil
}

// GetTranslations extracts translations map from context
func GetTranslations(ctx context.Context) map[string]string {
	if trans, ok := ctx.Value(transContextKey).(map[string]string); ok {
		return trans
	}
	return make(map[string]string)
}
