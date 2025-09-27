// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
    return &HealthHandler{}
}

type HealthResponse struct {
    OK   bool      `json:"ok"`
    Time time.Time `json:"time"`
}

func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		OK:   true,
		Time: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
