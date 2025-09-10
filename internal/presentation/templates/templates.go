package templates

import (
	"fmt"
	"html/template"
	"path/filepath"
)

// InitTemplates initializes the HTML templates from files
func InitTemplates() (*template.Template, error) {
	// Get the templates directory path relative to the binary
	templatesDir := "web/templates"

	// Parse the base template first
	tmpl, err := template.New("base").ParseFiles(filepath.Join(templatesDir, "base.html.tpl"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %w", err)
	}

	// Parse the additional templates
	additionalTemplates := []string{"index.html.tpl", "sign.html.tpl", "signatures.html.tpl", "embed.html.tpl"}
	for _, templateFile := range additionalTemplates {
		_, err = tmpl.ParseFiles(filepath.Join(templatesDir, templateFile))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
		}
	}

	return tmpl, nil
}
