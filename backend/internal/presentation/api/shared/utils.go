// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"net/http"
	"strings"
)

// GetClientIP extracts the real client IP address from the request
// It checks X-Forwarded-For, X-Real-IP, and falls back to RemoteAddr
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy/load balancer)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
