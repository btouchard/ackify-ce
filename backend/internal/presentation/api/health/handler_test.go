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
