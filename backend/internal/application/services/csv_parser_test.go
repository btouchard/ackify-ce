// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"strings"
	"testing"
)

func TestCSVParser_Parse_WithHeader(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name
jane@example.com,Jane Doe
john@example.com,John Smith`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.HasHeader {
		t.Error("expected HasHeader=true")
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if len(result.Signers) != 2 {
		t.Errorf("expected 2 signers, got %d", len(result.Signers))
	}

	if result.Signers[0].Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", result.Signers[0].Email)
	}

	if result.Signers[0].Name != "Jane Doe" {
		t.Errorf("expected name Jane Doe, got %s", result.Signers[0].Name)
	}
}

func TestCSVParser_Parse_WithHeaderReversedColumns(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `name,email
Jane Doe,jane@example.com
John Smith,john@example.com`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.HasHeader {
		t.Error("expected HasHeader=true")
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.Signers[0].Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", result.Signers[0].Email)
	}

	if result.Signers[0].Name != "Jane Doe" {
		t.Errorf("expected name Jane Doe, got %s", result.Signers[0].Name)
	}
}

func TestCSVParser_Parse_WithoutHeader(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `jane@example.com,Jane Doe
john@example.com,John Smith`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HasHeader {
		t.Error("expected HasHeader=false")
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.Signers[0].Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", result.Signers[0].Email)
	}
}

func TestCSVParser_Parse_SingleColumn(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email
jane@example.com
john@example.com`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.HasHeader {
		t.Error("expected HasHeader=true")
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.Signers[0].Name != "" {
		t.Errorf("expected empty name, got %s", result.Signers[0].Name)
	}
}

func TestCSVParser_Parse_SingleColumnNoHeader(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `jane@example.com
john@example.com`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HasHeader {
		t.Error("expected HasHeader=false")
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}
}

func TestCSVParser_Parse_SemicolonSeparator(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email;name
jane@example.com;Jane Doe
john@example.com;John Smith`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.Signers[0].Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", result.Signers[0].Email)
	}
}

func TestCSVParser_Parse_InvalidEmails(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name
jane@example.com,Jane Doe
not-an-email,Invalid User
john@example.com,John Smith
@missing.local,Bad Email`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.InvalidCount != 2 {
		t.Errorf("expected InvalidCount=2, got %d", result.InvalidCount)
	}

	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}

	for _, parseErr := range result.Errors {
		if parseErr.Error != "invalid_email_format" {
			t.Errorf("expected error 'invalid_email_format', got '%s'", parseErr.Error)
		}
	}
}

func TestCSVParser_Parse_EmptyLines(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name
jane@example.com,Jane Doe

john@example.com,John Smith
,`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	// Empty lines (including ",") are skipped, so InvalidCount should be 0
	if result.InvalidCount != 0 {
		t.Errorf("expected InvalidCount=0, got %d", result.InvalidCount)
	}

	// TotalLines should only count non-empty data rows
	if result.TotalLines != 2 {
		t.Errorf("expected TotalLines=2, got %d", result.TotalLines)
	}
}

func TestCSVParser_Parse_MaxLimit(t *testing.T) {
	parser := NewCSVParser(2)

	csvContent := `email,name
jane@example.com,Jane Doe
john@example.com,John Smith
alice@example.com,Alice Brown
bob@example.com,Bob Wilson`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2 (max limit), got %d", result.ValidCount)
	}

	if result.InvalidCount != 2 {
		t.Errorf("expected InvalidCount=2 (exceeds limit), got %d", result.InvalidCount)
	}

	// Check that the exceeded entries have the right error
	for _, parseErr := range result.Errors {
		if parseErr.Error != "max_signers_exceeded" {
			t.Errorf("expected error 'max_signers_exceeded', got '%s'", parseErr.Error)
		}
	}
}

func TestCSVParser_Parse_TrimWhitespace(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name
  jane@example.com  ,  Jane Doe
	john@example.com	,	John Smith	`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.Signers[0].Email != "jane@example.com" {
		t.Errorf("expected trimmed email jane@example.com, got '%s'", result.Signers[0].Email)
	}

	if result.Signers[0].Name != "Jane Doe" {
		t.Errorf("expected trimmed name 'Jane Doe', got '%s'", result.Signers[0].Name)
	}
}

func TestCSVParser_Parse_EmptyFile(t *testing.T) {
	parser := NewCSVParser(500)

	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalLines != 0 {
		t.Errorf("expected TotalLines=0, got %d", result.TotalLines)
	}

	if result.ValidCount != 0 {
		t.Errorf("expected ValidCount=0, got %d", result.ValidCount)
	}
}

func TestCSVParser_Parse_HeaderOnly(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.HasHeader {
		t.Error("expected HasHeader=true")
	}

	if result.TotalLines != 0 {
		t.Errorf("expected TotalLines=0, got %d", result.TotalLines)
	}

	if result.ValidCount != 0 {
		t.Errorf("expected ValidCount=0, got %d", result.ValidCount)
	}
}

func TestCSVParser_Parse_EmailNormalization(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name
JANE@EXAMPLE.COM,Jane Doe
John@Example.COM,John Smith`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Signers[0].Email != "jane@example.com" {
		t.Errorf("expected lowercase email jane@example.com, got %s", result.Signers[0].Email)
	}

	if result.Signers[1].Email != "john@example.com" {
		t.Errorf("expected lowercase email john@example.com, got %s", result.Signers[1].Email)
	}
}

func TestCSVParser_Parse_FrenchHeaders(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `courriel,nom
jane@example.com,Jane Dupont`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.HasHeader {
		t.Error("expected HasHeader=true for French headers")
	}

	if result.ValidCount != 1 {
		t.Errorf("expected ValidCount=1, got %d", result.ValidCount)
	}

	if result.Signers[0].Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", result.Signers[0].Email)
	}
}

func TestCSVParser_Parse_CRLFLineEndings(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := "email,name\r\njane@example.com,Jane Doe\r\njohn@example.com,John Smith"

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}
}

func TestCSVParser_Parse_SpecialCharactersInName(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name
jane@example.com,"Jean-François Müller"
john@example.com,José García`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.Signers[0].Name != "Jean-François Müller" {
		t.Errorf("expected name 'Jean-François Müller', got '%s'", result.Signers[0].Name)
	}

	if result.Signers[1].Name != "José García" {
		t.Errorf("expected name 'José García', got '%s'", result.Signers[1].Name)
	}
}

func TestCSVParser_Parse_QuotedFields(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `email,name
"jane@example.com","Doe, Jane"
"john@example.com","Smith, John"`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ValidCount != 2 {
		t.Errorf("expected ValidCount=2, got %d", result.ValidCount)
	}

	if result.Signers[0].Name != "Doe, Jane" {
		t.Errorf("expected name 'Doe, Jane', got '%s'", result.Signers[0].Name)
	}
}

func TestCSVParser_Parse_MissingEmailColumn(t *testing.T) {
	parser := NewCSVParser(500)

	csvContent := `name
Jane Doe
John Smith`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fail because "name" header doesn't contain emails
	// Since only 'name' is detected as header, emailCol will be -1
	// All lines will fail email validation
	if result.ValidCount != 0 {
		t.Errorf("expected ValidCount=0, got %d", result.ValidCount)
	}

	if result.InvalidCount != 2 {
		t.Errorf("expected InvalidCount=2, got %d", result.InvalidCount)
	}
}

func TestCSVParser_Parse_EntryNumbering(t *testing.T) {
	parser := NewCSVParser(500)

	// CSV with header - entry numbers should start at 1
	csvContent := `email,name
jane@example.com,Jane Doe
john@example.com,John Smith
alice@example.com,Alice Brown`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First entry should be number 1, not 2 (even though it's file line 2)
	if result.Signers[0].LineNumber != 1 {
		t.Errorf("expected first entry LineNumber=1, got %d", result.Signers[0].LineNumber)
	}

	if result.Signers[1].LineNumber != 2 {
		t.Errorf("expected second entry LineNumber=2, got %d", result.Signers[1].LineNumber)
	}

	if result.Signers[2].LineNumber != 3 {
		t.Errorf("expected third entry LineNumber=3, got %d", result.Signers[2].LineNumber)
	}
}

func TestCSVParser_Parse_EntryNumberingWithErrors(t *testing.T) {
	parser := NewCSVParser(500)

	// CSV with header and some invalid entries
	csvContent := `email,name
jane@example.com,Jane Doe
invalid-email,Bad User
john@example.com,John Smith`

	result, err := parser.Parse(strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Valid entries should be numbered 1 and 3
	if result.Signers[0].LineNumber != 1 {
		t.Errorf("expected first valid entry LineNumber=1, got %d", result.Signers[0].LineNumber)
	}

	if result.Signers[1].LineNumber != 3 {
		t.Errorf("expected second valid entry LineNumber=3, got %d", result.Signers[1].LineNumber)
	}

	// Error should be at entry 2
	if result.Errors[0].LineNumber != 2 {
		t.Errorf("expected error at entry LineNumber=2, got %d", result.Errors[0].LineNumber)
	}
}

func TestDetectSeparator(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected rune
	}{
		{
			name:     "comma separator",
			content:  "email,name\njane@example.com,Jane",
			expected: ',',
		},
		{
			name:     "semicolon separator",
			content:  "email;name\njane@example.com;Jane",
			expected: ';',
		},
		{
			name:     "more semicolons",
			content:  "email;name;extra\njane@example.com,Jane",
			expected: ';',
		},
		{
			name:     "equal count defaults to comma",
			content:  "email,name;extra",
			expected: ',',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSeparator(tt.content)
			if got != tt.expected {
				t.Errorf("expected separator '%c', got '%c'", tt.expected, got)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"test.name@example.com", true},
		{"test+tag@example.com", true},
		{"test@sub.example.com", true},
		{"TEST@EXAMPLE.COM", true},
		{"invalid", false},
		{"invalid@", false},
		{"@example.com", false},
		{"test@.com", false},
		{"test@example", false},
		{"", false},
		{"test@@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := isValidEmail(tt.email)
			if got != tt.expected {
				t.Errorf("isValidEmail(%s) = %v, expected %v", tt.email, got, tt.expected)
			}
		})
	}
}
