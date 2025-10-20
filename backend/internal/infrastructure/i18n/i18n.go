// SPDX-License-Identifier: AGPL-3.0-or-later
package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"
)

const (
	LangCookieName = "lang"
	DefaultLang    = "en"
)

var (
	SupportedLangs = []language.Tag{
		language.English,
		language.French,
		language.Italian,
		language.German,
		language.Spanish,
	}
	matcher = language.NewMatcher(SupportedLangs)
)

type I18n struct {
	translations map[string]map[string]string // lang -> key -> value
}

func NewI18n(localesDir string) (*I18n, error) {
	i18n := &I18n{
		translations: make(map[string]map[string]string),
	}

	// Load all supported language translations
	languages := []string{"en", "fr", "it", "de", "es"}
	for _, lang := range languages {
		filePath := filepath.Join(localesDir, lang+".json")
		if err := i18n.loadTranslations(filePath, lang); err != nil {
			return nil, fmt.Errorf("failed to load %s translations: %w", lang, err)
		}
	}

	return i18n, nil
}

func (i *I18n) loadTranslations(filePath, lang string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	translations := make(map[string]string)
	if err := json.Unmarshal(data, &translations); err != nil {
		return err
	}

	i.translations[lang] = translations
	return nil
}

// T translates a key for a given language
func (i *I18n) T(lang, key string) string {
	if translations, ok := i.translations[lang]; ok {
		if value, ok := translations[key]; ok {
			return value
		}
	}

	// Fallback to English
	if lang != "en" {
		if translations, ok := i.translations["en"]; ok {
			if value, ok := translations[key]; ok {
				return value
			}
		}
	}

	// Return key if translation not found
	return key
}

// GetLangFromRequest extracts language from cookie or Accept-Language header
func GetLangFromRequest(r *http.Request) string {
	// First, check cookie
	if cookie, err := r.Cookie(LangCookieName); err == nil && cookie.Value != "" {
		lang := normalizeLang(cookie.Value)
		if isSupported(lang) {
			return lang
		}
	}

	// Then, check Accept-Language header
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		tags, _, _ := language.ParseAcceptLanguage(acceptLang)
		if len(tags) > 0 {
			_, index, _ := matcher.Match(tags...)
			if index < len(SupportedLangs) {
				return normalizeLang(SupportedLangs[index].String())
			}
		}
	}

	// Default to English
	return DefaultLang
}

// SetLangCookie sets the language preference cookie
func SetLangCookie(w http.ResponseWriter, lang string, secureCookies bool) {
	lang = normalizeLang(lang)
	if !isSupported(lang) {
		lang = DefaultLang
	}

	cookie := &http.Cookie{
		Name:     LangCookieName,
		Value:    lang,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60, // 1 year
		HttpOnly: true,
		Secure:   secureCookies,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

// normalizeLang normalizes language codes (en-US -> en, fr-FR -> fr, it-IT -> it, etc.)
func normalizeLang(lang string) string {
	lang = strings.ToLower(lang)
	// Extract base language code (before - or _)
	if idx := strings.IndexAny(lang, "-_"); idx > 0 {
		return lang[:idx]
	}
	return lang
}

// isSupported checks if a language is supported
func isSupported(lang string) bool {
	lang = normalizeLang(lang)
	supportedLangs := []string{"en", "fr", "it", "de", "es"}
	for _, supported := range supportedLangs {
		if lang == supported {
			return true
		}
	}
	return false
}

// GetTranslations returns all translations for a given language
func (i *I18n) GetTranslations(lang string) map[string]string {
	if translations, ok := i.translations[lang]; ok {
		return translations
	}
	return i.translations[DefaultLang]
}
