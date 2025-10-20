// SPDX-License-Identifier: AGPL-3.0-or-later
package checksum

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestComputeRemoteChecksum_Success(t *testing.T) {
	// Create test HTTP server
	content := "Hello, World!"
	expectedChecksum := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f" // SHA-256 of "Hello, World!"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		if r.Method == "GET" {
			w.Write([]byte(content))
		}
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true      // For testing with httptest
	opts.InsecureSkipVerify = true // Accept self-signed certs
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.ChecksumHex != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, result.ChecksumHex)
	}

	if result.Algorithm != "SHA-256" {
		t.Errorf("Expected algorithm SHA-256, got %s", result.Algorithm)
	}
}

func TestComputeRemoteChecksum_TooLarge(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", "20971520") // 20 MB
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true
	opts.InsecureSkipVerify = true
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for too large file, got result")
	}
}

func TestComputeRemoteChecksum_WrongContentType(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", "100")
		if r.Method == "GET" {
			w.Write([]byte("<html>test</html>"))
		}
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true
	opts.InsecureSkipVerify = true
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for wrong content type, got result")
	}
}

func TestComputeRemoteChecksum_HTTPNotHTTPS(t *testing.T) {
	// Test HTTP (not HTTPS) - should be rejected
	opts := DefaultOptions()
	result, err := ComputeRemoteChecksum("http://example.com/file.pdf", opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for HTTP URL, got result")
	}
}

func TestComputeRemoteChecksum_StreamedGETFallback(t *testing.T) {
	content := "Test content for streaming"
	expectedChecksum := "e9157132b66b4ef7eb0395b483f0dd30364ad356919c59f9f5eeb26087339b64"

	// Server that doesn't support HEAD
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte(content))
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true
	opts.InsecureSkipVerify = true
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.ChecksumHex != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, result.ChecksumHex)
	}
}

func TestComputeRemoteChecksum_ExceedsSizeDuringStreaming(t *testing.T) {
	// Server without Content-Length that returns too much data
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		// Write more than MaxBytes
		largeContent := strings.Repeat("x", 11*1024*1024) // 11 MB
		w.Write([]byte(largeContent))
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true
	opts.InsecureSkipVerify = true
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for oversized stream, got result")
	}
}

func TestComputeRemoteChecksum_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true
	opts.InsecureSkipVerify = true
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for 404 error, got result")
	}
}

func TestComputeRemoteChecksum_TooManyRedirects(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always redirect, creating an infinite loop
		http.Redirect(w, r, server.URL+"/redirect", http.StatusFound)
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true
	opts.InsecureSkipVerify = true
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should fail due to too many redirects
	if result != nil {
		t.Error("Expected nil result for too many redirects, got result")
	}
}

func TestIsAllowedContentType(t *testing.T) {
	allowedTypes := []string{"application/pdf", "image/*", "application/vnd.oasis.opendocument.*"}

	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/pdf", true},
		{"application/pdf; charset=utf-8", true},
		{"image/png", true},
		{"image/jpeg", true},
		{"image/svg+xml", true},
		{"application/vnd.oasis.opendocument.text", true},
		{"application/vnd.oasis.opendocument.spreadsheet", true},
		{"text/html", false},
		{"application/json", false},
		{"video/mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := isAllowedContentType(tt.contentType, allowedTypes)
			if result != tt.expected {
				t.Errorf("isAllowedContentType(%q) = %v, expected %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.1.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"::1", true},
		{"2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := parseIPAddress(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			result := isPrivateIP(ip)
			if result != tt.expected {
				t.Errorf("isPrivateIP(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func parseIPAddress(s string) net.IP {
	return net.ParseIP(s)
}

func TestIsBlockedHost_Localhost(t *testing.T) {
	hosts := []string{"localhost", "127.0.0.1"}

	for _, host := range hosts {
		if !isBlockedHost(host) {
			t.Errorf("Expected %s to be blocked", host)
		}
	}
}

func TestComputeRemoteChecksum_ImageContentType(t *testing.T) {
	content := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	expectedChecksum := "0f4636c78f65d3639ece5a064b5ae753e3408614a14fb18ab4d7540d2c248543"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		if r.Method == "GET" {
			w.Write(content)
		}
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.SkipSSRFCheck = true
	opts.InsecureSkipVerify = true
	result, err := ComputeRemoteChecksum(server.URL, opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.ChecksumHex != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, result.ChecksumHex)
	}
}
