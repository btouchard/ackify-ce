// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/backend/pkg/checksum"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

type documentRepository interface {
	Create(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error)
	GetByDocID(ctx context.Context, docID string) (*models.Document, error)
	FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error)
}

// DocumentService handles document metadata operations and unique ID generation
type DocumentService struct {
	repo           documentRepository
	checksumConfig *config.ChecksumConfig
}

// NewDocumentService initializes the document service with its repository dependency
func NewDocumentService(repo documentRepository, checksumConfig *config.ChecksumConfig) *DocumentService {
	return &DocumentService{
		repo:           repo,
		checksumConfig: checksumConfig,
	}
}

// CreateDocumentRequest represents the request to create a document
type CreateDocumentRequest struct {
	Reference string `json:"reference" validate:"required,min=1"`
	Title     string `json:"title"`
}

// CreateDocument generates a collision-resistant base36 identifier and persists document metadata
func (s *DocumentService) CreateDocument(ctx context.Context, req CreateDocumentRequest) (*models.Document, error) {
	logger.Logger.Info("Document creation attempt", "reference", req.Reference)

	var docID string
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		docID = generateDocID()

		existing, err := s.repo.GetByDocID(ctx, docID)
		if err != nil {
			return nil, fmt.Errorf("failed to check doc_id uniqueness: %w", err)
		}

		if existing == nil {
			break
		}

		logger.Logger.Debug("Generated doc_id already exists, retrying",
			"doc_id", docID, "attempt", i+1)
	}

	var url, title string
	if strings.HasPrefix(req.Reference, "http://") || strings.HasPrefix(req.Reference, "https://") {
		url = req.Reference

		if req.Title == "" {
			title = extractTitleFromURL(req.Reference)
		} else {
			title = req.Title
		}
	} else {
		url = ""
		if req.Title == "" {
			title = req.Reference
		} else {
			title = req.Title
		}
	}

	input := models.DocumentInput{
		Title: title,
		URL:   url,
	}

	// Automatically compute checksum for remote URLs if enabled
	if url != "" && s.checksumConfig != nil {
		checksumResult := s.computeChecksumForURL(url)
		if checksumResult != nil {
			input.Checksum = checksumResult.ChecksumHex
			input.ChecksumAlgorithm = checksumResult.Algorithm
			logger.Logger.Info("Automatically computed checksum for document",
				"doc_id", docID,
				"checksum", checksumResult.ChecksumHex,
				"algorithm", checksumResult.Algorithm)
		}
	}

	doc, err := s.repo.Create(ctx, docID, input, "")
	if err != nil {
		logger.Logger.Error("Failed to create document",
			"doc_id", docID,
			"error", err.Error())
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	logger.Logger.Info("Document created successfully",
		"doc_id", docID,
		"url", url,
		"title", title)

	return doc, nil
}

func generateDocID() string {
	timestamp := time.Now().Unix()
	timestampB36 := strconv.FormatInt(timestamp, 36)

	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const suffixLen = 4

	suffix := make([]byte, suffixLen)
	for i := range suffix {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			suffix[i] = charset[(int(timestamp)+i)%len(charset)]
		} else {
			suffix[i] = charset[n.Int64()]
		}
	}

	return timestampB36 + string(suffix)
}

func extractTitleFromURL(urlStr string) string {
	urlStr = strings.TrimRight(urlStr, "/")

	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	parts := strings.Split(urlStr, "/")

	if len(parts) == 0 {
		return urlStr
	}

	var lastSegment string
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			lastSegment = parts[i]
			break
		}
	}

	if lastSegment == "" {
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
		return urlStr
	}

	// Remove query parameters
	if idx := strings.Index(lastSegment, "?"); idx >= 0 {
		lastSegment = lastSegment[:idx]
	}

	// Remove fragment
	if idx := strings.Index(lastSegment, "#"); idx >= 0 {
		lastSegment = lastSegment[:idx]
	}

	// Remove file extension
	if idx := strings.LastIndex(lastSegment, "."); idx > 0 {
		lastSegment = lastSegment[:idx]
	}

	// Clean up hash/ID suffixes (Notion, GitHub, GitLab, etc.)
	lastSegment = cleanHashSuffix(lastSegment)

	return lastSegment
}

// cleanHashSuffix removes common hash/ID patterns appended by various platforms
// Examples:
//   - "Introduction-to-Cybersecurity-26b2915834718093a062f54c798d63c5" -> "Introduction-to-Cybersecurity"
//   - "My-Document-abc123def456" -> "My-Document"
//   - "Report-2024-1a2b3c4d5e6f" -> "Report-2024"
func cleanHashSuffix(title string) string {
	// Pattern 1: Remove UUID-like suffixes (with dashes) - check this first before splitting
	// Example: "Title-a1b2c3d4-e5f6-7890-abcd-ef1234567890" -> "Title"
	// UUID format: 8-4-4-4-12 = 36 chars total with dashes
	parts := strings.Split(title, "-")
	if len(parts) >= 6 {
		// Check if last 5 segments form a UUID pattern
		potentialUUID := strings.Join(parts[len(parts)-5:], "-")
		cleanUUID := strings.ReplaceAll(potentialUUID, "-", "")
		if len(cleanUUID) == 32 && isHexString(cleanUUID) {
			return strings.Join(parts[:len(parts)-5], "-")
		}
	}

	// Pattern 2: Remove long hexadecimal suffixes (24+ chars) - Notion style
	// Example: "Title-26b2915834718093a062f54c798d63c5" -> "Title"
	if idx := strings.LastIndex(title, "-"); idx > 0 {
		suffix := title[idx+1:]
		if len(suffix) >= 24 && isHexString(suffix) {
			return title[:idx]
		}
	}

	// Pattern 3: Remove short hash suffixes (8-16 chars) only if alphanumeric
	// Example: "Document-abc123def" -> "Document"
	if idx := strings.LastIndex(title, "-"); idx > 0 {
		suffix := title[idx+1:]
		if len(suffix) >= 8 && len(suffix) <= 16 && isAlphanumeric(suffix) && hasLettersAndNumbers(suffix) {
			return title[:idx]
		}
	}

	// Pattern 4: Remove numeric-only suffixes (timestamps, IDs) 8+ digits
	// Example: "Article-1234567890" -> "Article"
	if idx := strings.LastIndex(title, "-"); idx > 0 {
		suffix := title[idx+1:]
		if len(suffix) >= 8 && isNumericString(suffix) {
			return title[:idx]
		}
	}

	// Pattern 5: Remove base64-like suffixes (URL-safe base64)
	// Example: "Page-aGVsbG93b3JsZA" -> "Page"
	if idx := strings.LastIndex(title, "-"); idx > 0 {
		suffix := title[idx+1:]
		if len(suffix) >= 12 && isBase64Like(suffix) {
			return title[:idx]
		}
	}

	return title
}

// isHexString checks if a string contains only hexadecimal characters (0-9, a-f, A-F)
func isHexString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return false
		}
	}
	return true
}

// isAlphanumeric checks if string contains only letters and numbers
func isAlphanumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			return false
		}
	}
	return true
}

// hasLettersAndNumbers checks if string contains both letters AND numbers (likely a hash)
func hasLettersAndNumbers(s string) bool {
	hasLetter := false
	hasNumber := false
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			hasLetter = true
		}
		if ch >= '0' && ch <= '9' {
			hasNumber = true
		}
	}
	return hasLetter && hasNumber
}

// isNumericString checks if string contains only digits
func isNumericString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// isBase64Like checks if string looks like base64 encoding
func isBase64Like(s string) bool {
	if len(s) == 0 {
		return false
	}
	base64Chars := 0
	for _, ch := range s {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '-' {
			base64Chars++
		}
	}
	// If 90%+ of chars are base64-compatible, likely base64
	return float64(base64Chars)/float64(len(s)) >= 0.9
}

// computeChecksumForURL attempts to compute the checksum for a remote URL
// Returns nil if the checksum cannot be computed (error, too large, etc.)
func (s *DocumentService) computeChecksumForURL(url string) *checksum.Result {
	if s.checksumConfig == nil {
		return nil
	}

	opts := checksum.ComputeOptions{
		MaxBytes:           s.checksumConfig.MaxBytes,
		TimeoutMs:          s.checksumConfig.TimeoutMs,
		MaxRedirects:       s.checksumConfig.MaxRedirects,
		AllowedContentType: s.checksumConfig.AllowedContentType,
		SkipSSRFCheck:      s.checksumConfig.SkipSSRFCheck,
		InsecureSkipVerify: s.checksumConfig.InsecureSkipVerify,
	}

	result, err := checksum.ComputeRemoteChecksum(url, opts)
	if err != nil {
		logger.Logger.Warn("Failed to compute checksum for URL",
			"url", url,
			"error", err.Error())
		return nil
	}

	return result
}

type ReferenceType string

const (
	ReferenceTypeURL       ReferenceType = "url"
	ReferenceTypePath      ReferenceType = "path"
	ReferenceTypeReference ReferenceType = "reference"
)

func detectReferenceType(ref string) ReferenceType {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ReferenceTypeURL
	}

	if strings.Contains(ref, "/") || strings.Contains(ref, "\\") {
		return ReferenceTypePath
	}

	return ReferenceTypeReference
}

// FindByReference finds a document by its reference without creating it
func (s *DocumentService) FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error) {
	doc, err := s.repo.FindByReference(ctx, ref, refType)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// FindOrCreateDocument performs smart lookup by URL/path/reference or creates new document if not found
func (s *DocumentService) FindOrCreateDocument(ctx context.Context, ref string) (*models.Document, bool, error) {
	logger.Logger.Info("Find or create document", "reference", ref)

	refType := detectReferenceType(ref)
	logger.Logger.Debug("Reference type detected", "type", refType, "reference", ref)

	doc, err := s.repo.FindByReference(ctx, ref, string(refType))
	if err != nil {
		logger.Logger.Error("Error searching for document", "reference", ref, "error", err.Error())
		return nil, false, fmt.Errorf("failed to search for document: %w", err)
	}

	if doc != nil {
		logger.Logger.Info("Document found", "doc_id", doc.DocID, "reference", ref)
		return doc, false, nil
	}

	logger.Logger.Info("Document not found, creating new one", "reference", ref)

	var title string
	switch refType {
	case ReferenceTypeURL:
		title = extractTitleFromURL(ref)
	case ReferenceTypePath:
		title = extractTitleFromURL(ref)
	case ReferenceTypeReference:
		title = ref
	}

	createReq := CreateDocumentRequest{
		Reference: ref,
		Title:     title,
	}

	if refType == ReferenceTypeReference {
		input := models.DocumentInput{
			Title: title,
			URL:   "",
		}

		doc, err := s.repo.Create(ctx, ref, input, "")
		if err != nil {
			logger.Logger.Error("Failed to create document with custom doc_id",
				"doc_id", ref,
				"error", err.Error())
			return nil, false, fmt.Errorf("failed to create document: %w", err)
		}

		logger.Logger.Info("Document created with custom doc_id",
			"doc_id", ref,
			"title", title)

		return doc, true, nil
	}

	// For URL references, compute checksum before creating
	if refType == ReferenceTypeURL && s.checksumConfig != nil {
		logger.Logger.Debug("Computing checksum for URL reference", "url", ref)
		checksumResult := s.computeChecksumForURL(ref)
		if checksumResult != nil {
			logger.Logger.Info("Automatically computed checksum for URL reference",
				"url", ref,
				"checksum", checksumResult.ChecksumHex,
				"algorithm", checksumResult.Algorithm)
		}
	}

	doc, err = s.CreateDocument(ctx, createReq)
	if err != nil {
		return nil, false, err
	}

	return doc, true, nil
}
