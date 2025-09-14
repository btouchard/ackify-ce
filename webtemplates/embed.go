package webtemplates

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.tpl
var TemplatesFS embed.FS

// InitTemplates initializes the HTML templates from embedded files
func InitTemplates() (*template.Template, error) {
	// Parse the base template first
	tmpl, err := template.New("base").ParseFS(TemplatesFS, "templates/base.html.tpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %w", err)
	}

	// Parse the additional templates
	additionalTemplates := []string{"templates/index.html.tpl", "templates/sign.html.tpl", "templates/signatures.html.tpl", "templates/embed.html.tpl"}
	for _, templateFile := range additionalTemplates {
		_, err = tmpl.ParseFS(TemplatesFS, templateFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
		}
	}

	return tmpl, nil
}
