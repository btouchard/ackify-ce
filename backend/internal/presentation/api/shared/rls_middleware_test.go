// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// mockTenantProvider is a test implementation of tenant.Provider
type mockTenantProvider struct {
	tenantID uuid.UUID
	err      error
}

func (m *mockTenantProvider) CurrentTenant(ctx context.Context) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return m.tenantID, nil
}

func TestNewRLSMiddleware(t *testing.T) {
	tenantID := uuid.New()
	provider := &mockTenantProvider{tenantID: tenantID}

	// Test with nil db (should still create middleware)
	m := NewRLSMiddleware(nil, provider)
	if m == nil {
		t.Error("NewRLSMiddleware returned nil")
	}
	if m.db != nil {
		t.Error("db should be nil")
	}
	if m.tenants != provider {
		t.Error("tenants should be the provided provider")
	}
}

func TestStatusCapturingResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	wrapped := &statusCapturingResponseWriter{ResponseWriter: rr, status: http.StatusOK}

	// First WriteHeader should set status
	wrapped.WriteHeader(http.StatusNotFound)
	if wrapped.status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, wrapped.status)
	}
	if !wrapped.wroteHeader {
		t.Error("wroteHeader should be true after WriteHeader")
	}

	// Second WriteHeader should be ignored
	wrapped.WriteHeader(http.StatusOK)
	if wrapped.status != http.StatusNotFound {
		t.Errorf("Expected status %d after second WriteHeader, got %d", http.StatusNotFound, wrapped.status)
	}
}

func TestStatusCapturingResponseWriter_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	wrapped := &statusCapturingResponseWriter{ResponseWriter: rr, status: http.StatusOK}

	// Write without explicit WriteHeader should trigger implicit 200
	_, err := wrapped.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	if wrapped.status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, wrapped.status)
	}
	if !wrapped.wroteHeader {
		t.Error("wroteHeader should be true after Write")
	}
}

func TestRLSMiddleware_TenantError(t *testing.T) {
	provider := &mockTenantProvider{err: sql.ErrNoRows}
	m := NewRLSMiddleware(nil, provider)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	m.Handler(handler).ServeHTTP(rr, req)

	if called {
		t.Error("Handler should not be called when tenant lookup fails")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}
