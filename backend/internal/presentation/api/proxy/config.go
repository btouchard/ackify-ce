// SPDX-License-Identifier: AGPL-3.0-or-later
package proxy

import "time"

// Configuration constants for the proxy
const (
	// MaxDocumentSize is the maximum document size allowed (50 MB)
	MaxDocumentSize = 50 * 1024 * 1024

	// DialTimeout is the timeout for establishing connection
	DialTimeout = 10 * time.Second

	// HeaderTimeout is the timeout for receiving response headers
	HeaderTimeout = 30 * time.Second

	// TotalTimeout is the total timeout for the entire request
	TotalTimeout = 5 * time.Minute

	// RateLimitPerIP is the rate limit per IP address (60/min)
	RateLimitPerIP = 60

	// RateLimitPerIPDoc is the rate limit per IP+document combination (20/min)
	RateLimitPerIPDoc = 20

	// RateLimitPerDoc is the rate limit per document (300/min)
	RateLimitPerDoc = 300

	// RateLimitWindow is the time window for rate limiting
	RateLimitWindow = time.Minute
)

// AllowedMIMETypes is the whitelist of allowed MIME types
var AllowedMIMETypes = map[string]bool{
	// PDF
	"application/pdf": true,

	// Images
	"image/png":     true,
	"image/jpeg":    true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,

	// Text
	"text/html":     true,
	"text/markdown": true,
	"text/plain":    true,
}

// IsAllowedMIMEType checks if a MIME type is allowed
func IsAllowedMIMEType(mimeType string) bool {
	return AllowedMIMETypes[mimeType]
}
