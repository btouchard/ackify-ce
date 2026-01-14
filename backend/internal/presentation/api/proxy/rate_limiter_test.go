// SPDX-License-Identifier: AGPL-3.0-or-later
package proxy

import (
	"testing"
	"time"
)

func TestRateLimiter_BasicLimit(t *testing.T) {
	rl := NewRateLimiter(3, 2, 5, 100*time.Millisecond)
	defer rl.Stop()

	ip := "192.168.1.1"
	docID := "doc-1"

	// First 2 requests should be allowed (limited by per-IP+Doc)
	for i := 0; i < 2; i++ {
		result := rl.Check(ip, docID)
		if !result.Allowed {
			t.Errorf("Request %d should be allowed, got denied (limit: %s)", i+1, result.LimitType)
		}
	}

	// Third request should be denied (per-IP+Doc limit = 2)
	result := rl.Check(ip, docID)
	if result.Allowed {
		t.Error("Request 3 should be denied due to per-IP+Doc limit")
	}
	if result.LimitType != "ip_doc" {
		t.Errorf("Expected limit type 'ip_doc', got '%s'", result.LimitType)
	}
}

func TestRateLimiter_PerIPLimit(t *testing.T) {
	rl := NewRateLimiter(2, 10, 10, 100*time.Millisecond)
	defer rl.Stop()

	ip := "192.168.1.1"

	// Make requests to different documents to avoid per-IP+Doc limit
	for i := 0; i < 2; i++ {
		result := rl.Check(ip, "doc-"+string(rune('a'+i)))
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Third request should be denied (per-IP limit = 2)
	result := rl.Check(ip, "doc-c")
	if result.Allowed {
		t.Error("Request 3 should be denied due to per-IP limit")
	}
	if result.LimitType != "ip" {
		t.Errorf("Expected limit type 'ip', got '%s'", result.LimitType)
	}
}

func TestRateLimiter_PerDocLimit(t *testing.T) {
	rl := NewRateLimiter(100, 100, 2, 100*time.Millisecond)
	defer rl.Stop()

	docID := "doc-1"

	// Make requests from different IPs
	for i := 0; i < 2; i++ {
		result := rl.Check("192.168.1."+string(rune('1'+i)), docID)
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Third request should be denied (per-Doc limit = 2)
	result := rl.Check("192.168.1.100", docID)
	if result.Allowed {
		t.Error("Request 3 should be denied due to per-Doc limit")
	}
	if result.LimitType != "doc" {
		t.Errorf("Expected limit type 'doc', got '%s'", result.LimitType)
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 2, 2, 50*time.Millisecond)
	defer rl.Stop()

	ip := "192.168.1.1"
	docID := "doc-1"

	// Use up the limit
	for i := 0; i < 2; i++ {
		result := rl.Check(ip, docID)
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Should be denied now
	result := rl.Check(ip, docID)
	if result.Allowed {
		t.Error("Request should be denied after limit reached")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	result = rl.Check(ip, docID)
	if !result.Allowed {
		t.Error("Request should be allowed after window expires")
	}
}

func TestRateLimiter_RetryAfter(t *testing.T) {
	window := 100 * time.Millisecond
	rl := NewRateLimiter(1, 1, 1, window)
	defer rl.Stop()

	ip := "192.168.1.1"
	docID := "doc-1"

	// Use up the limit
	rl.Check(ip, docID)

	// Check RetryAfter
	result := rl.Check(ip, docID)
	if result.Allowed {
		t.Error("Request should be denied")
	}
	if result.RetryAfter != window {
		t.Errorf("Expected RetryAfter %v, got %v", window, result.RetryAfter)
	}
}

func TestRateLimiter_DifferentIPsSameDoc(t *testing.T) {
	rl := NewRateLimiter(100, 100, 100, 100*time.Millisecond)
	defer rl.Stop()

	docID := "doc-1"

	// Different IPs should each have their own limits
	for i := 0; i < 10; i++ {
		ip := "192.168.1." + string(rune('1'+i))
		result := rl.Check(ip, docID)
		if !result.Allowed {
			t.Errorf("Request from IP %s should be allowed", ip)
		}
	}
}

func TestRateLimiter_SameIPDifferentDocs(t *testing.T) {
	rl := NewRateLimiter(100, 100, 100, 100*time.Millisecond)
	defer rl.Stop()

	ip := "192.168.1.1"

	// Same IP to different docs should each have their own IP+Doc limits
	for i := 0; i < 10; i++ {
		docID := "doc-" + string(rune('a'+i))
		result := rl.Check(ip, docID)
		if !result.Allowed {
			t.Errorf("Request for doc %s should be allowed", docID)
		}
	}
}
