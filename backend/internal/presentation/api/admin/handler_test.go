//go:build integration
// +build integration

// SPDX-License-Identifier: AGPL-3.0-or-later
package admin_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/admin"
	"github.com/btouchard/ackify-ce/backend/pkg/crypto"
	"github.com/go-chi/chi/v5"
)

func setupTestDB(t *testing.T) *database.TestDB {
	// Use database.SetupTestDB which applies all migrations automatically
	// This ensures test schema matches production schema
	testDB := database.SetupTestDB(t)
	return testDB
}

func TestAdminHandler_GetDocumentStatus_WithUnexpectedSignatures(t *testing.T) {
	testDB := setupTestDB(t)

	ctx := context.Background()

	// Setup repositories and services
	docRepo := database.NewDocumentRepository(testDB.DB)
	sigRepo := database.NewSignatureRepository(testDB.DB)
	expectedSignerRepo := database.NewExpectedSignerRepository(testDB.DB)
	signer, _ := crypto.NewEd25519Signer()
	sigService := services.NewSignatureService(sigRepo, docRepo, signer)

	// Create test document
	docID := "test-doc-001"
	_, err := docRepo.CreateOrUpdate(ctx, docID, models.DocumentInput{
		Title:             "Test Document",
		URL:               "https://example.com/doc.pdf",
		Checksum:          "abc123",
		ChecksumAlgorithm: "SHA-256",
		Description:       "Test description",
	}, "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add expected signer
	err = expectedSignerRepo.AddExpected(ctx, docID, []models.ContactInfo{
		{Email: "expected@example.com", Name: "Expected User"},
	}, "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to add expected signer: %v", err)
	}

	// Create signature from expected user
	expectedUser := &models.User{
		Sub:   "expected-sub",
		Email: "expected@example.com",
		Name:  "Expected User",
	}
	err = sigService.CreateSignature(ctx, &models.SignatureRequest{
		DocID: docID,
		User:  expectedUser,
	})
	if err != nil {
		t.Fatalf("Failed to create expected signature: %v", err)
	}

	// Create signature from unexpected user
	unexpectedUser := &models.User{
		Sub:   "unexpected-sub",
		Email: "unexpected@example.com",
		Name:  "Unexpected User",
	}
	err = sigService.CreateSignature(ctx, &models.SignatureRequest{
		DocID: docID,
		User:  unexpectedUser,
	})
	if err != nil {
		t.Fatalf("Failed to create unexpected signature: %v", err)
	}

	// Create admin handler
	handler := admin.NewHandler(docRepo, expectedSignerRepo, nil, sigService, "https://example.com")

	// Create HTTP request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/"+docID+"/status", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("docId", docID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGetDocumentStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response struct {
		DocID           string `json:"docId"`
		ExpectedSigners []struct {
			Email     string `json:"email"`
			HasSigned bool   `json:"hasSigned"`
		} `json:"expectedSigners"`
		UnexpectedSignatures []struct {
			UserEmail   string  `json:"userEmail"`
			UserName    *string `json:"userName,omitempty"`
			SignedAtUTC string  `json:"signedAtUTC"`
		} `json:"unexpectedSignatures"`
		Stats struct {
			ExpectedCount  int     `json:"expectedCount"`
			SignedCount    int     `json:"signedCount"`
			PendingCount   int     `json:"pendingCount"`
			CompletionRate float64 `json:"completionRate"`
		} `json:"stats"`
		ShareLink string `json:"shareLink"`
	}

	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if response.DocID != docID {
		t.Errorf("Expected docId %s, got %s", docID, response.DocID)
	}

	// Check expected signers
	if len(response.ExpectedSigners) != 1 {
		t.Errorf("Expected 1 expected signer, got %d", len(response.ExpectedSigners))
	} else {
		if response.ExpectedSigners[0].Email != "expected@example.com" {
			t.Errorf("Expected email 'expected@example.com', got '%s'", response.ExpectedSigners[0].Email)
		}
		if !response.ExpectedSigners[0].HasSigned {
			t.Error("Expected signer should have signed")
		}
	}

	// Check unexpected signatures
	if len(response.UnexpectedSignatures) != 1 {
		t.Fatalf("Expected 1 unexpected signature, got %d", len(response.UnexpectedSignatures))
	}
	if response.UnexpectedSignatures[0].UserEmail != "unexpected@example.com" {
		t.Errorf("Expected unexpected email 'unexpected@example.com', got '%s'", response.UnexpectedSignatures[0].UserEmail)
	}
	if response.UnexpectedSignatures[0].UserName == nil || *response.UnexpectedSignatures[0].UserName != "Unexpected User" {
		t.Error("Expected unexpected userName to be 'Unexpected User'")
	}

	// Check stats
	if response.Stats.ExpectedCount != 1 {
		t.Errorf("Expected expectedCount 1, got %d", response.Stats.ExpectedCount)
	}
	if response.Stats.SignedCount != 1 {
		t.Errorf("Expected signedCount 1, got %d", response.Stats.SignedCount)
	}
	if response.Stats.CompletionRate != 100.0 {
		t.Errorf("Expected completionRate 100.0, got %f", response.Stats.CompletionRate)
	}

	// Check share link
	expectedShareLink := "https://example.com/?doc=" + docID
	if response.ShareLink != expectedShareLink {
		t.Errorf("Expected shareLink '%s', got '%s'", expectedShareLink, response.ShareLink)
	}
}

func TestAdminHandler_GetDocumentStatus_NoExpectedSigners(t *testing.T) {
	testDB := setupTestDB(t)

	ctx := context.Background()

	// Setup repositories and services
	docRepo := database.NewDocumentRepository(testDB.DB)
	sigRepo := database.NewSignatureRepository(testDB.DB)
	expectedSignerRepo := database.NewExpectedSignerRepository(testDB.DB)
	signer, _ := crypto.NewEd25519Signer()
	sigService := services.NewSignatureService(sigRepo, docRepo, signer)

	// Create test document
	docID := "test-doc-002"

	// Create signature from user (no expected signers)
	user := &models.User{
		Sub:   "user-sub",
		Email: "user@example.com",
		Name:  "Test User",
	}
	err := sigService.CreateSignature(ctx, &models.SignatureRequest{
		DocID: docID,
		User:  user,
	})
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Create admin handler
	handler := admin.NewHandler(docRepo, expectedSignerRepo, nil, sigService, "https://example.com")

	// Create HTTP request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/documents/"+docID+"/status", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("docId", docID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGetDocumentStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response struct {
		ExpectedSigners      []interface{} `json:"expectedSigners"`
		UnexpectedSignatures []struct {
			UserEmail string `json:"userEmail"`
		} `json:"unexpectedSignatures"`
	}

	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if len(response.ExpectedSigners) != 0 {
		t.Errorf("Expected 0 expected signers, got %d", len(response.ExpectedSigners))
	}

	// All signatures should be unexpected since there are no expected signers
	if len(response.UnexpectedSignatures) != 1 {
		t.Fatalf("Expected 1 unexpected signature, got %d", len(response.UnexpectedSignatures))
	}
	if response.UnexpectedSignatures[0].UserEmail != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", response.UnexpectedSignatures[0].UserEmail)
	}
}
