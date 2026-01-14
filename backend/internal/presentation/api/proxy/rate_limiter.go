// SPDX-License-Identifier: AGPL-3.0-or-later
package proxy

import (
	"sync"
	"time"
)

// RateLimiter provides multi-tier rate limiting for the proxy
type RateLimiter struct {
	mu sync.RWMutex

	// Per-IP tracking
	perIP map[string]*rateBucket

	// Per-IP+Doc tracking
	perIPDoc map[string]*rateBucket

	// Per-Doc tracking
	perDoc map[string]*rateBucket

	// Configuration
	limitPerIP    int
	limitPerIPDoc int
	limitPerDoc   int
	window        time.Duration

	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// rateBucket tracks request counts within a time window
type rateBucket struct {
	timestamps []time.Time
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool
	RetryAfter time.Duration
	LimitType  string // "ip", "ip_doc", or "doc"
}

// NewRateLimiter creates a new multi-tier rate limiter
func NewRateLimiter(limitPerIP, limitPerIPDoc, limitPerDoc int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		perIP:         make(map[string]*rateBucket),
		perIPDoc:      make(map[string]*rateBucket),
		perDoc:        make(map[string]*rateBucket),
		limitPerIP:    limitPerIP,
		limitPerIPDoc: limitPerIPDoc,
		limitPerDoc:   limitPerDoc,
		window:        window,
		cleanupTicker: time.NewTicker(window),
		stopCleanup:   make(chan struct{}),
	}

	// Start background cleanup
	go rl.cleanupLoop()

	return rl
}

// Check checks if a request is allowed and records it if so
func (rl *RateLimiter) Check(ip, docID string) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	ipDocKey := ip + ":" + docID

	// Check per-IP limit
	if !rl.checkAndRecord(rl.perIP, ip, rl.limitPerIP, now) {
		return RateLimitResult{
			Allowed:    false,
			RetryAfter: rl.window,
			LimitType:  "ip",
		}
	}

	// Check per-IP+Doc limit
	if !rl.checkAndRecord(rl.perIPDoc, ipDocKey, rl.limitPerIPDoc, now) {
		// Rollback IP counter
		rl.rollback(rl.perIP, ip)
		return RateLimitResult{
			Allowed:    false,
			RetryAfter: rl.window,
			LimitType:  "ip_doc",
		}
	}

	// Check per-Doc limit
	if !rl.checkAndRecord(rl.perDoc, docID, rl.limitPerDoc, now) {
		// Rollback IP and IP+Doc counters
		rl.rollback(rl.perIP, ip)
		rl.rollback(rl.perIPDoc, ipDocKey)
		return RateLimitResult{
			Allowed:    false,
			RetryAfter: rl.window,
			LimitType:  "doc",
		}
	}

	return RateLimitResult{
		Allowed: true,
	}
}

// checkAndRecord checks if adding a request would exceed the limit and records it if not
func (rl *RateLimiter) checkAndRecord(buckets map[string]*rateBucket, key string, limit int, now time.Time) bool {
	bucket, exists := buckets[key]
	if !exists {
		bucket = &rateBucket{timestamps: make([]time.Time, 0, limit)}
		buckets[key] = bucket
	}

	// Clean old timestamps
	bucket.timestamps = rl.filterRecent(bucket.timestamps, now)

	// Check limit
	if len(bucket.timestamps) >= limit {
		return false
	}

	// Record request
	bucket.timestamps = append(bucket.timestamps, now)
	return true
}

// rollback removes the most recent timestamp from a bucket
func (rl *RateLimiter) rollback(buckets map[string]*rateBucket, key string) {
	if bucket, exists := buckets[key]; exists && len(bucket.timestamps) > 0 {
		bucket.timestamps = bucket.timestamps[:len(bucket.timestamps)-1]
	}
}

// filterRecent filters timestamps to only include those within the window
func (rl *RateLimiter) filterRecent(timestamps []time.Time, now time.Time) []time.Time {
	cutoff := now.Add(-rl.window)
	result := make([]time.Time, 0, len(timestamps))
	for _, t := range timestamps {
		if t.After(cutoff) {
			result = append(result, t)
		}
	}
	return result
}

// cleanupLoop periodically cleans up old entries
func (rl *RateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup removes empty buckets
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	rl.cleanupBuckets(rl.perIP, now)
	rl.cleanupBuckets(rl.perIPDoc, now)
	rl.cleanupBuckets(rl.perDoc, now)
}

// cleanupBuckets removes empty buckets from a map
func (rl *RateLimiter) cleanupBuckets(buckets map[string]*rateBucket, now time.Time) {
	for key, bucket := range buckets {
		bucket.timestamps = rl.filterRecent(bucket.timestamps, now)
		if len(bucket.timestamps) == 0 {
			delete(buckets, key)
		}
	}
}

// Stop stops the rate limiter's cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}
