package webtemplates

import (
	"testing"
)

func TestTemplatesFS(t *testing.T) {
	// Test that the embedded filesystem contains the expected files
	expectedFiles := []string{
		"templates/base.html.tpl",
		"templates/index.html.tpl",
		"templates/sign.html.tpl",
		"templates/signatures.html.tpl",
		"templates/embed.html.tpl",
	}

	for _, file := range expectedFiles {
		data, err := TemplatesFS.ReadFile(file)
		if err != nil {
			t.Errorf("Failed to read embedded file %s: %v", file, err)
		}
		if len(data) == 0 {
			t.Errorf("Embedded file %s is empty", file)
		}
	}
}

func TestInitTemplates(t *testing.T) {
	// Test that InitTemplates works correctly
	tmpl, err := InitTemplates()
	if err != nil {
		t.Fatalf("InitTemplates failed: %v", err)
	}

	if tmpl == nil {
		t.Fatal("InitTemplates returned nil template")
	}

	// Test that all expected templates are parsed
	expectedTemplateNames := []string{
		"base",
		"base.html.tpl",
		"index.html.tpl",
		"sign.html.tpl",
		"signatures.html.tpl",
		"embed.html.tpl",
	}

	for _, name := range expectedTemplateNames {
		if tmpl.Lookup(name) == nil {
			t.Errorf("Template %s not found in parsed templates", name)
		}
	}
}
