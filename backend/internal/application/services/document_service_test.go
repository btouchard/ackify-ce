// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"strings"
	"testing"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
)

// Test generateDocID function
func TestGenerateDocID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Generate first ID"},
		{"Generate second ID"},
		{"Generate third ID"},
	}

	seenIDs := make(map[string]bool)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := generateDocID()

			// Check length (timestamp in base36 + 4 random chars = ~10-11 chars)
			if len(id) < 10 || len(id) > 12 {
				t.Errorf("Expected ID length between 10-12 chars, got %d (%s)", len(id), id)
			}

			// Check all characters are alphanumeric lowercase
			for _, ch := range id {
				if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
					t.Errorf("ID contains invalid character: %c in %s", ch, id)
				}
			}

			// Check uniqueness (probabilistic, but should be unique in small sample)
			if seenIDs[id] {
				t.Errorf("Duplicate ID generated: %s", id)
			}
			seenIDs[id] = true
		})
	}
}

// Test extractTitleFromURL function
func TestExtractTitleFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with file extension",
			url:      "https://example.com/documents/report.pdf",
			expected: "report",
		},
		{
			name:     "URL without extension",
			url:      "https://example.com/documents/annual-report",
			expected: "annual-report",
		},
		{
			name:     "URL with query parameters",
			url:      "https://example.com/doc.pdf?version=2",
			expected: "doc",
		},
		{
			name:     "URL with fragment",
			url:      "https://example.com/guide.html#section1",
			expected: "guide",
		},
		{
			name:     "URL with trailing slash",
			url:      "https://example.com/page/",
			expected: "page",
		},
		{
			name:     "Domain only",
			url:      "https://example.com",
			expected: "example",
		},
		{
			name:     "Domain with trailing slash",
			url:      "https://example.com/",
			expected: "example",
		},
		{
			name:     "HTTP URL",
			url:      "http://example.com/test.txt",
			expected: "test",
		},
		{
			name:     "URL with path and extension",
			url:      "https://docs.example.com/v2/api/reference.json",
			expected: "reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTitleFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("extractTitleFromURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// mockDocumentRepository is a mock implementation for testing
type mockDocumentRepository struct {
	createFunc          func(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error)
	getByDocIDFunc      func(ctx context.Context, docID string) (*models.Document, error)
	findByReferenceFunc func(ctx context.Context, ref string, refType string) (*models.Document, error)
}

func (m *mockDocumentRepository) Create(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, docID, input, createdBy)
	}
	return &models.Document{
		DocID:             docID,
		Title:             input.Title,
		URL:               input.URL,
		Checksum:          input.Checksum,
		ChecksumAlgorithm: input.ChecksumAlgorithm,
		CreatedBy:         createdBy,
	}, nil
}

func (m *mockDocumentRepository) GetByDocID(ctx context.Context, docID string) (*models.Document, error) {
	if m.getByDocIDFunc != nil {
		return m.getByDocIDFunc(ctx, docID)
	}
	return nil, nil // Not found by default
}

func (m *mockDocumentRepository) FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error) {
	if m.findByReferenceFunc != nil {
		return m.findByReferenceFunc(ctx, ref, refType)
	}
	return nil, nil // Not found by default
}

// Test CreateDocument with URL reference
func TestDocumentService_CreateDocument_WithURL(t *testing.T) {
	mockRepo := &mockDocumentRepository{}
	service := NewDocumentService(mockRepo, nil) // nil config = no automatic checksum

	req := CreateDocumentRequest{
		Reference: "https://example.com/important-doc.pdf",
		Title:     "",
	}

	ctx := context.Background()
	doc, err := service.CreateDocument(ctx, req)

	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created, got nil")
	}

	// Check that URL was extracted
	if doc.URL != "https://example.com/important-doc.pdf" {
		t.Errorf("Expected URL to be %q, got %q", "https://example.com/important-doc.pdf", doc.URL)
	}

	// Check that title was extracted from URL
	if doc.Title != "important-doc" {
		t.Errorf("Expected title to be %q, got %q", "important-doc", doc.Title)
	}

	// Check that doc_id was generated
	if doc.DocID == "" {
		t.Error("Expected doc_id to be generated")
	}
}

// Test CreateDocument with URL reference and custom title
func TestDocumentService_CreateDocument_WithURLAndTitle(t *testing.T) {
	mockRepo := &mockDocumentRepository{}
	service := NewDocumentService(mockRepo, nil)

	req := CreateDocumentRequest{
		Reference: "https://example.com/doc.pdf",
		Title:     "My Custom Title",
	}

	ctx := context.Background()
	doc, err := service.CreateDocument(ctx, req)

	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}

	// Check that URL was extracted
	if doc.URL != "https://example.com/doc.pdf" {
		t.Errorf("Expected URL to be %q, got %q", "https://example.com/doc.pdf", doc.URL)
	}

	// Check that custom title was used
	if doc.Title != "My Custom Title" {
		t.Errorf("Expected title to be %q, got %q", "My Custom Title", doc.Title)
	}
}

// Test CreateDocument with plain text reference
func TestDocumentService_CreateDocument_WithPlainReference(t *testing.T) {
	mockRepo := &mockDocumentRepository{}
	service := NewDocumentService(mockRepo, nil)

	req := CreateDocumentRequest{
		Reference: "company-policy-2024",
		Title:     "",
	}

	ctx := context.Background()
	doc, err := service.CreateDocument(ctx, req)

	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}

	// Check that URL is empty
	if doc.URL != "" {
		t.Errorf("Expected URL to be empty, got %q", doc.URL)
	}

	// Check that reference was used as title
	if doc.Title != "company-policy-2024" {
		t.Errorf("Expected title to be %q, got %q", "company-policy-2024", doc.Title)
	}
}

// Test CreateDocument with plain reference and custom title
func TestDocumentService_CreateDocument_WithPlainReferenceAndTitle(t *testing.T) {
	mockRepo := &mockDocumentRepository{}
	service := NewDocumentService(mockRepo, nil)

	req := CreateDocumentRequest{
		Reference: "doc-ref-123",
		Title:     "Employee Handbook",
	}

	ctx := context.Background()
	doc, err := service.CreateDocument(ctx, req)

	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}

	// Check that URL is empty
	if doc.URL != "" {
		t.Errorf("Expected URL to be empty, got %q", doc.URL)
	}

	// Check that custom title was used
	if doc.Title != "Employee Handbook" {
		t.Errorf("Expected title to be %q, got %q", "Employee Handbook", doc.Title)
	}
}

// Test CreateDocument with HTTP URL
func TestDocumentService_CreateDocument_WithHTTPURL(t *testing.T) {
	mockRepo := &mockDocumentRepository{}
	service := NewDocumentService(mockRepo, nil)

	req := CreateDocumentRequest{
		Reference: "http://example.com/doc.html",
		Title:     "",
	}

	ctx := context.Background()
	doc, err := service.CreateDocument(ctx, req)

	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}

	// Check that URL was extracted (HTTP should work too)
	if doc.URL != "http://example.com/doc.html" {
		t.Errorf("Expected URL to be %q, got %q", "http://example.com/doc.html", doc.URL)
	}

	// Check that title was extracted
	if doc.Title != "doc" {
		t.Errorf("Expected title to be %q, got %q", "doc", doc.Title)
	}
}

// Test CreateDocument with ID collision retry
func TestDocumentService_CreateDocument_IDCollisionRetry(t *testing.T) {
	collisionCount := 0
	mockRepo := &mockDocumentRepository{
		getByDocIDFunc: func(ctx context.Context, docID string) (*models.Document, error) {
			// First two attempts return existing document (collision)
			if collisionCount < 2 {
				collisionCount++
				return &models.Document{DocID: docID}, nil
			}
			// Third attempt returns nil (ID is available)
			return nil, nil
		},
	}

	service := NewDocumentService(mockRepo, nil)

	req := CreateDocumentRequest{
		Reference: "test-doc",
		Title:     "",
	}

	ctx := context.Background()
	doc, err := service.CreateDocument(ctx, req)

	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}

	// Should have retried at least twice
	if collisionCount < 2 {
		t.Errorf("Expected at least 2 collision retries, got %d", collisionCount)
	}

	if doc == nil {
		t.Fatal("Expected document to be created after retries")
	}
}

// Test that generated IDs are URL-safe
func TestGenerateDocID_URLSafe(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := generateDocID()

		// Check no uppercase letters
		if strings.ToLower(id) != id {
			t.Errorf("ID contains uppercase letters: %s", id)
		}

		// Check no special characters that need encoding
		specialChars := []string{"/", "?", "#", "&", "=", "+", " ", "%"}
		for _, char := range specialChars {
			if strings.Contains(id, char) {
				t.Errorf("ID contains special character %q: %s", char, id)
			}
		}
	}
}

// Test detectReferenceType function
func TestDetectReferenceType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ref      string
		expected ReferenceType
	}{
		{
			name:     "HTTPS URL",
			ref:      "https://example.com/document.pdf",
			expected: ReferenceTypeURL,
		},
		{
			name:     "HTTP URL",
			ref:      "http://example.com/doc",
			expected: ReferenceTypeURL,
		},
		{
			name:     "Unix path",
			ref:      "/home/user/documents/file.pdf",
			expected: ReferenceTypePath,
		},
		{
			name:     "Windows path",
			ref:      "C:\\Users\\Documents\\file.pdf",
			expected: ReferenceTypePath,
		},
		{
			name:     "Relative path with forward slash",
			ref:      "docs/file.pdf",
			expected: ReferenceTypePath,
		},
		{
			name:     "Relative path with backslash",
			ref:      "docs\\file.pdf",
			expected: ReferenceTypePath,
		},
		{
			name:     "Plain reference",
			ref:      "policy-2024",
			expected: ReferenceTypeReference,
		},
		{
			name:     "Plain reference with dashes",
			ref:      "company-doc-v2",
			expected: ReferenceTypeReference,
		},
		{
			name:     "Plain reference with underscores",
			ref:      "employee_handbook_2024",
			expected: ReferenceTypeReference,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := detectReferenceType(tt.ref)
			if result != tt.expected {
				t.Errorf("detectReferenceType(%q) = %q, want %q", tt.ref, result, tt.expected)
			}
		})
	}
}

// Test FindByReference success
func TestDocumentService_FindByReference_Success(t *testing.T) {
	t.Parallel()

	expectedDoc := &models.Document{
		DocID: "test123",
		Title: "Test Document",
		URL:   "https://example.com/test.pdf",
	}

	mockRepo := &mockDocumentRepository{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			if ref == "https://example.com/test.pdf" && refType == "url" {
				return expectedDoc, nil
			}
			return nil, nil
		},
	}

	service := NewDocumentService(mockRepo, nil)
	ctx := context.Background()

	doc, err := service.FindByReference(ctx, "https://example.com/test.pdf", "url")

	if err != nil {
		t.Fatalf("FindByReference failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be found, got nil")
	}

	if doc.DocID != expectedDoc.DocID {
		t.Errorf("Expected DocID %q, got %q", expectedDoc.DocID, doc.DocID)
	}
}

// Test FindByReference not found
func TestDocumentService_FindByReference_NotFound(t *testing.T) {
	t.Parallel()

	mockRepo := &mockDocumentRepository{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			return nil, nil
		},
	}

	service := NewDocumentService(mockRepo, nil)
	ctx := context.Background()

	doc, err := service.FindByReference(ctx, "nonexistent", "reference")

	if err != nil {
		t.Fatalf("FindByReference should not error when document not found: %v", err)
	}

	if doc != nil {
		t.Errorf("Expected nil document, got %+v", doc)
	}
}

// Test FindOrCreateDocument - found existing document
func TestDocumentService_FindOrCreateDocument_Found(t *testing.T) {
	t.Parallel()

	existingDoc := &models.Document{
		DocID: "existing123",
		Title: "Existing Document",
		URL:   "https://example.com/existing.pdf",
	}

	mockRepo := &mockDocumentRepository{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			if ref == "https://example.com/existing.pdf" {
				return existingDoc, nil
			}
			return nil, nil
		},
	}

	service := NewDocumentService(mockRepo, nil)
	ctx := context.Background()

	doc, created, err := service.FindOrCreateDocument(ctx, "https://example.com/existing.pdf")

	if err != nil {
		t.Fatalf("FindOrCreateDocument failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be returned, got nil")
	}

	if created {
		t.Error("Expected created to be false for existing document")
	}

	if doc.DocID != existingDoc.DocID {
		t.Errorf("Expected DocID %q, got %q", existingDoc.DocID, doc.DocID)
	}
}

// Test FindOrCreateDocument - create new document with URL
func TestDocumentService_FindOrCreateDocument_CreateWithURL(t *testing.T) {
	t.Parallel()

	mockRepo := &mockDocumentRepository{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			return nil, nil // Not found
		},
	}

	service := NewDocumentService(mockRepo, nil)
	ctx := context.Background()

	doc, created, err := service.FindOrCreateDocument(ctx, "https://example.com/new-doc.pdf")

	if err != nil {
		t.Fatalf("FindOrCreateDocument failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created, got nil")
	}

	if !created {
		t.Error("Expected created to be true for new document")
	}

	if doc.URL != "https://example.com/new-doc.pdf" {
		t.Errorf("Expected URL %q, got %q", "https://example.com/new-doc.pdf", doc.URL)
	}

	if doc.Title != "new-doc" {
		t.Errorf("Expected title %q, got %q", "new-doc", doc.Title)
	}
}

// Test FindOrCreateDocument - create new document with path
func TestDocumentService_FindOrCreateDocument_CreateWithPath(t *testing.T) {
	t.Parallel()

	mockRepo := &mockDocumentRepository{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			return nil, nil // Not found
		},
	}

	service := NewDocumentService(mockRepo, nil)
	ctx := context.Background()

	doc, created, err := service.FindOrCreateDocument(ctx, "/home/user/important-file.pdf")

	if err != nil {
		t.Fatalf("FindOrCreateDocument failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created, got nil")
	}

	if !created {
		t.Error("Expected created to be true for new document")
	}

	// Path is extracted as title (like extractTitleFromURL does for paths)
	if doc.Title != "important-file" {
		t.Errorf("Expected title %q, got %q", "important-file", doc.Title)
	}

	// URL should be empty for paths (they're not http/https)
	if doc.URL != "" {
		t.Errorf("Expected URL to be empty for path, got %q", doc.URL)
	}
}

// Test FindOrCreateDocument - create new document with plain reference
func TestDocumentService_FindOrCreateDocument_CreateWithReference(t *testing.T) {
	t.Parallel()

	mockRepo := &mockDocumentRepository{
		findByReferenceFunc: func(ctx context.Context, ref string, refType string) (*models.Document, error) {
			return nil, nil // Not found
		},
		createFunc: func(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
			return &models.Document{
				DocID:     docID,
				Title:     input.Title,
				URL:       input.URL,
				CreatedBy: createdBy,
			}, nil
		},
	}

	service := NewDocumentService(mockRepo, nil)
	ctx := context.Background()

	doc, created, err := service.FindOrCreateDocument(ctx, "company-policy-2024")

	if err != nil {
		t.Fatalf("FindOrCreateDocument failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created, got nil")
	}

	if !created {
		t.Error("Expected created to be true for new document")
	}

	// For plain reference, doc_id should be the reference itself
	if doc.DocID != "company-policy-2024" {
		t.Errorf("Expected DocID to be the reference %q, got %q", "company-policy-2024", doc.DocID)
	}

	if doc.Title != "company-policy-2024" {
		t.Errorf("Expected title %q, got %q", "company-policy-2024", doc.Title)
	}

	if doc.URL != "" {
		t.Errorf("Expected URL to be empty for plain reference, got %q", doc.URL)
	}
}
