// SPDX-License-Identifier: AGPL-3.0-or-later
package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	handler := NewHandler()

	assert.NotNil(t, handler)
}

func TestHandler_HandleHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET returns 200 OK",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST also works (health check should be method-agnostic)",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HEAD also works",
			method:         http.MethodHead,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			handler := NewHandler()
			req := httptest.NewRequest(tt.method, "/api/v1/health", nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.HandleHealth(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Validate response body for non-HEAD requests
			if tt.method != http.MethodHead {
				// Response is wrapped in {"data": {...}}
				var wrapper struct {
					Data HealthResponse `json:"data"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
				require.NoError(t, err, "Response should be valid JSON")

				assert.Equal(t, "ok", wrapper.Data.Status)
				assert.NotZero(t, wrapper.Data.Timestamp)

				// Timestamp should be recent (within last 5 seconds)
				now := time.Now()
				assert.WithinDuration(t, now, wrapper.Data.Timestamp, 5*time.Second)
			}
		})
	}
}

func TestHandler_HandleHealth_ResponseFormat(t *testing.T) {
	t.Parallel()

	handler := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.HandleHealth(rec, req)

	// Check Content-Type
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Validate JSON structure
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check wrapper structure
	assert.Contains(t, response, "data")

	// Get data object
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")

	// Check required fields in data
	assert.Contains(t, data, "status")
	assert.Contains(t, data, "timestamp")

	// Validate status value
	status, ok := data["status"].(string)
	require.True(t, ok, "status should be a string")
	assert.Equal(t, "ok", status)

	// Validate timestamp format (RFC3339)
	timestampStr, ok := data["timestamp"].(string)
	require.True(t, ok, "timestamp should be a string")

	_, err = time.Parse(time.RFC3339, timestampStr)
	assert.NoError(t, err, "timestamp should be in RFC3339 format")
}

func TestHandler_HandleHealth_Concurrent(t *testing.T) {
	t.Parallel()

	handler := NewHandler()

	const numRequests = 100
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	// Spawn concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			defer func() { done <- true }()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			rec := httptest.NewRecorder()

			handler.HandleHealth(rec, req)

			if rec.Code != http.StatusOK {
				errors <- assert.AnError
			}

			var wrapper struct {
				Data HealthResponse `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &wrapper); err != nil {
				errors <- err
			}
		}()
	}

	// Wait for all requests
	for i := 0; i < numRequests; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errCount++
	}

	assert.Equal(t, 0, errCount, "All concurrent health checks should succeed")
}

func TestHandler_HandleHealth_Idempotency(t *testing.T) {
	t.Parallel()

	handler := NewHandler()

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec1 := httptest.NewRecorder()
	handler.HandleHealth(rec1, req1)

	var wrapper1 struct {
		Data HealthResponse `json:"data"`
	}
	err := json.Unmarshal(rec1.Body.Bytes(), &wrapper1)
	require.NoError(t, err)

	// Small delay
	time.Sleep(10 * time.Millisecond)

	// Second request
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec2 := httptest.NewRecorder()
	handler.HandleHealth(rec2, req2)

	var wrapper2 struct {
		Data HealthResponse `json:"data"`
	}
	err = json.Unmarshal(rec2.Body.Bytes(), &wrapper2)
	require.NoError(t, err)

	// Status should be same
	assert.Equal(t, wrapper1.Data.Status, wrapper2.Data.Status)

	// Timestamps should be different (but close)
	assert.NotEqual(t, wrapper1.Data.Timestamp, wrapper2.Data.Timestamp)
	assert.WithinDuration(t, wrapper1.Data.Timestamp, wrapper2.Data.Timestamp, 1*time.Second)
}

func BenchmarkHandler_HandleHealth(b *testing.B) {
	handler := NewHandler()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()

		handler.HandleHealth(rec, req)
	}
}

func BenchmarkHandler_HandleHealth_Parallel(b *testing.B) {
	handler := NewHandler()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			rec := httptest.NewRecorder()

			handler.HandleHealth(rec, req)
		}
	})
}
