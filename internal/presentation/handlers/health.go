package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse represents a health check response
type HealthResponse struct {
	OK   bool      `json:"ok"`
	Time time.Time `json:"time"`
}

// HandleHealth returns the application health status
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	response := HealthResponse{
		OK:   true,
		Time: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
