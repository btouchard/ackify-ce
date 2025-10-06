// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	txtTemplate "text/template"
)

type Renderer struct {
	templateDir   string
	baseURL       string
	organisation  string
	fromName      string
	fromMail      string
	defaultLocale string
}

type TemplateData struct {
	Organisation string
	BaseURL      string
	FromName     string
	FromMail     string
	Data         map[string]any
}

func NewRenderer(templateDir, baseURL, organisation, fromName, fromMail, defaultLocale string) *Renderer {
	return &Renderer{
		templateDir:   templateDir,
		baseURL:       baseURL,
		organisation:  organisation,
		fromName:      fromName,
		fromMail:      fromMail,
		defaultLocale: defaultLocale,
	}
}

func (r *Renderer) Render(templateName, locale string, data map[string]any) (htmlBody, textBody string, err error) {
	if locale == "" {
		locale = r.defaultLocale
	}

	templateData := TemplateData{
		Organisation: r.organisation,
		BaseURL:      r.baseURL,
		FromName:     r.fromName,
		FromMail:     r.fromMail,
		Data:         data,
	}

	htmlBody, err = r.renderHTML(templateName, locale, templateData)
	if err != nil {
		return "", "", fmt.Errorf("failed to render HTML: %w", err)
	}

	textBody, err = r.renderText(templateName, locale, templateData)
	if err != nil {
		return "", "", fmt.Errorf("failed to render text: %w", err)
	}

	return htmlBody, textBody, nil
}

func (r *Renderer) renderHTML(templateName, locale string, data TemplateData) (string, error) {
	baseTemplatePath := filepath.Join(r.templateDir, "base.html.tmpl")
	templatePath := r.resolveTemplatePath(templateName, locale, "html.tmpl")

	if templatePath == "" {
		return "", fmt.Errorf("template not found: %s (locale: %s)", templateName, locale)
	}

	tmpl, err := htmlTemplate.ParseFiles(baseTemplatePath, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (r *Renderer) renderText(templateName, locale string, data TemplateData) (string, error) {
	baseTemplatePath := filepath.Join(r.templateDir, "base.txt.tmpl")
	templatePath := r.resolveTemplatePath(templateName, locale, "txt.tmpl")

	if templatePath == "" {
		return "", fmt.Errorf("template not found: %s (locale: %s)", templateName, locale)
	}

	tmpl, err := txtTemplate.ParseFiles(baseTemplatePath, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (r *Renderer) resolveTemplatePath(templateName, locale, extension string) string {
	localizedPath := filepath.Join(r.templateDir, fmt.Sprintf("%s.%s.%s", templateName, locale, extension))
	if _, err := os.Stat(localizedPath); err == nil {
		return localizedPath
	}

	fallbackPath := filepath.Join(r.templateDir, fmt.Sprintf("%s.en.%s", templateName, extension))
	if _, err := os.Stat(fallbackPath); err == nil {
		return fallbackPath
	}

	return ""
}
