// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"encoding/csv"
	"errors"
	"io"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// CSVSignerEntry represents a valid signer entry parsed from CSV
type CSVSignerEntry struct {
	LineNumber int    `json:"lineNumber"`
	Email      string `json:"email"`
	Name       string `json:"name"`
}

// CSVParseError represents an error for a specific line in the CSV
type CSVParseError struct {
	LineNumber int    `json:"lineNumber"`
	Content    string `json:"content"`
	Error      string `json:"error"`
}

// CSVParseResult contains the complete result of parsing a CSV file
type CSVParseResult struct {
	Signers      []CSVSignerEntry `json:"signers"`
	Errors       []CSVParseError  `json:"errors"`
	TotalLines   int              `json:"totalLines"`
	ValidCount   int              `json:"validCount"`
	InvalidCount int              `json:"invalidCount"`
	HasHeader    bool             `json:"hasHeader"`
}

// CSVParserConfig holds configuration for CSV parsing
type CSVParserConfig struct {
	MaxSigners int
}

// CSVParser handles CSV file parsing for expected signers import
type CSVParser struct {
	config CSVParserConfig
}

// NewCSVParser creates a new CSV parser with the given configuration
func NewCSVParser(maxSigners int) *CSVParser {
	return &CSVParser{
		config: CSVParserConfig{
			MaxSigners: maxSigners,
		},
	}
}

// Parse reads and parses a CSV file from the given reader
func (p *CSVParser) Parse(reader io.Reader) (*CSVParseResult, error) {
	result := &CSVParseResult{
		Signers: []CSVSignerEntry{},
		Errors:  []CSVParseError{},
	}

	// Try to detect the separator by reading the first line
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return result, nil
	}

	separator := detectSeparator(string(content))
	csvReader := csv.NewReader(strings.NewReader(string(content)))
	csvReader.Comma = separator
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields
	csvReader.TrimLeadingSpace = true
	csvReader.LazyQuotes = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return result, nil
	}

	// Detect header and column positions
	emailCol, nameCol, hasHeader := detectColumns(records[0])

	result.HasHeader = hasHeader

	startRow := 0
	if hasHeader {
		startRow = 1
	}

	entryNumber := 0 // Counter for entry numbering (starts at 1 for first entry)

	for i := startRow; i < len(records); i++ {
		row := records[i]

		// Skip empty rows
		if isEmptyRow(row) {
			continue
		}

		entryNumber++ // Increment for each non-empty data row
		result.TotalLines++

		// Check max signers limit
		if p.config.MaxSigners > 0 && len(result.Signers) >= p.config.MaxSigners {
			result.Errors = append(result.Errors, CSVParseError{
				LineNumber: entryNumber,
				Content:    strings.Join(row, string(separator)),
				Error:      "max_signers_exceeded",
			})
			result.InvalidCount++
			continue
		}

		entry, parseErr := parseRow(row, emailCol, nameCol, entryNumber)
		if parseErr != nil {
			result.Errors = append(result.Errors, CSVParseError{
				LineNumber: entryNumber,
				Content:    strings.Join(row, string(separator)),
				Error:      parseErr.Error(),
			})
			result.InvalidCount++
			continue
		}

		result.Signers = append(result.Signers, *entry)
		result.ValidCount++
	}

	return result, nil
}

// detectSeparator determines if the CSV uses comma or semicolon
func detectSeparator(content string) rune {
	firstLine := strings.Split(content, "\n")[0]

	semicolonCount := strings.Count(firstLine, ";")
	commaCount := strings.Count(firstLine, ",")

	if semicolonCount > commaCount {
		return ';'
	}
	return ','
}

// detectColumns analyzes the first row to detect column positions and if it's a header
func detectColumns(firstRow []string) (emailCol, nameCol int, hasHeader bool) {
	emailCol = -1
	nameCol = -1
	hasHeader = false

	for i, field := range firstRow {
		normalized := strings.ToLower(strings.TrimSpace(field))
		switch normalized {
		case "email", "e-mail", "mail", "courriel":
			emailCol = i
			hasHeader = true
		case "name", "nom", "prenom", "prénom", "firstname", "lastname", "fullname", "full_name":
			nameCol = i
			hasHeader = true
		}
	}

	// If header detected, return found positions
	if hasHeader {
		return emailCol, nameCol, true
	}

	// No header detected - determine column positions from data
	// Try to identify email column by checking for @ symbol
	for i, field := range firstRow {
		trimmed := strings.TrimSpace(field)
		if isValidEmail(trimmed) {
			emailCol = i
			// If there's another column, assume it's the name
			if len(firstRow) > 1 {
				if i == 0 {
					nameCol = 1
				} else {
					nameCol = 0
				}
			}
			break
		}
	}

	// If we couldn't find an email column, assume first column is email
	if emailCol == -1 {
		emailCol = 0
		if len(firstRow) > 1 {
			nameCol = 1
		}
	}

	return emailCol, nameCol, false
}

// parseRow extracts email and name from a row
func parseRow(row []string, emailCol, nameCol, lineNumber int) (*CSVSignerEntry, error) {
	email := ""
	name := ""

	if emailCol >= 0 && emailCol < len(row) {
		email = strings.TrimSpace(row[emailCol])
	}

	if nameCol >= 0 && nameCol < len(row) {
		name = strings.TrimSpace(row[nameCol])
	}

	// Validate email
	if email == "" {
		return nil, errors.New("email_required")
	}

	email = strings.ToLower(email)

	if !isValidEmail(email) {
		return nil, errors.New("invalid_email_format")
	}

	return &CSVSignerEntry{
		LineNumber: lineNumber,
		Email:      email,
		Name:       name,
	}, nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// isEmptyRow checks if all fields in a row are empty
func isEmptyRow(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
