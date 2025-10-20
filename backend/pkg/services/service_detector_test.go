// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TESTS - DetectServiceFromReferrer
// ============================================================================

func TestDetectServiceFromReferrer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		referrer     string
		expectedName string
		expectedIcon string
		expectedType string
		expectNil    bool
	}{
		// Empty/nil cases
		{
			name:      "empty referrer",
			referrer:  "",
			expectNil: true,
		},

		// Google services
		{
			name:         "Google Docs",
			referrer:     "google-docs",
			expectedName: "Google Docs",
			expectedIcon: "https://cdn.simpleicons.org/googledocs",
			expectedType: "docs",
		},
		{
			name:         "Google Sheets",
			referrer:     "google-sheets",
			expectedName: "Google Sheets",
			expectedIcon: "https://cdn.simpleicons.org/googlesheets",
			expectedType: "sheets",
		},
		{
			name:         "Google Slides",
			referrer:     "google-slides",
			expectedName: "Google Slides",
			expectedIcon: "https://cdn.simpleicons.org/googleslides",
			expectedType: "presentation",
		},
		{
			name:         "Google Drive",
			referrer:     "google-drive",
			expectedName: "Google Drive",
			expectedIcon: "https://cdn.simpleicons.org/googledrive",
			expectedType: "storage",
		},
		{
			name:         "Google (generic)",
			referrer:     "google",
			expectedName: "Google",
			expectedIcon: "https://cdn.simpleicons.org/google",
			expectedType: "google",
		},

		// Code platforms
		{
			name:         "GitHub",
			referrer:     "github",
			expectedName: "GitHub",
			expectedIcon: "https://cdn.simpleicons.org/github",
			expectedType: "code",
		},
		{
			name:         "GitLab",
			referrer:     "gitlab",
			expectedName: "GitLab",
			expectedIcon: "https://cdn.simpleicons.org/gitlab",
			expectedType: "code",
		},

		// Collaboration tools
		{
			name:         "Notion",
			referrer:     "notion",
			expectedName: "Notion",
			expectedIcon: "https://cdn.simpleicons.org/notion",
			expectedType: "notes",
		},
		{
			name:         "Confluence",
			referrer:     "confluence",
			expectedName: "Confluence",
			expectedIcon: "https://cdn.simpleicons.org/confluence",
			expectedType: "wiki",
		},
		{
			name:         "Outline",
			referrer:     "outline",
			expectedName: "Outline",
			expectedIcon: "https://cdn.simpleicons.org/outline",
			expectedType: "wiki",
		},

		// Microsoft
		{
			name:         "Microsoft Office",
			referrer:     "microsoft",
			expectedName: "Microsoft Office",
			expectedIcon: "https://cdn.simpleicons.org/microsoft",
			expectedType: "office",
		},

		// Communication platforms
		{
			name:         "Slack",
			referrer:     "slack",
			expectedName: "Slack",
			expectedIcon: "https://cdn.simpleicons.org/slack",
			expectedType: "chat",
		},
		{
			name:         "Discord",
			referrer:     "discord",
			expectedName: "Discord",
			expectedIcon: "https://cdn.simpleicons.org/discord",
			expectedType: "chat",
		},

		// Project management
		{
			name:         "Trello",
			referrer:     "trello",
			expectedName: "Trello",
			expectedIcon: "https://cdn.simpleicons.org/trello",
			expectedType: "boards",
		},
		{
			name:         "Asana",
			referrer:     "asana",
			expectedName: "Asana",
			expectedIcon: "https://cdn.simpleicons.org/asana",
			expectedType: "tasks",
		},
		{
			name:         "Monday.com",
			referrer:     "monday",
			expectedName: "Monday.com",
			expectedIcon: "https://cdn.simpleicons.org/monday",
			expectedType: "project",
		},

		// Design tools
		{
			name:         "Figma",
			referrer:     "figma",
			expectedName: "Figma",
			expectedIcon: "https://cdn.simpleicons.org/figma",
			expectedType: "design",
		},
		{
			name:         "Miro",
			referrer:     "miro",
			expectedName: "Miro",
			expectedIcon: "https://cdn.simpleicons.org/miro",
			expectedType: "whiteboard",
		},

		// Storage
		{
			name:         "Dropbox",
			referrer:     "dropbox",
			expectedName: "Dropbox",
			expectedIcon: "https://cdn.simpleicons.org/dropbox",
			expectedType: "storage",
		},

		// Unknown/custom service
		{
			name:         "unknown service",
			referrer:     "my-custom-service",
			expectedName: "my-custom-service",
			expectedIcon: "https://cdn.simpleicons.org/link",
			expectedType: "custom",
		},
		{
			name:         "custom URL",
			referrer:     "https://example.com",
			expectedName: "https://example.com",
			expectedIcon: "https://cdn.simpleicons.org/link",
			expectedType: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := DetectServiceFromReferrer(tt.referrer)

			if tt.expectNil {
				assert.Nil(t, result, "Expected nil for empty referrer")
				return
			}

			require.NotNil(t, result, "Should return ServiceInfo")
			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedIcon, result.Icon)
			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.referrer, result.Referrer)
		})
	}
}

// ============================================================================
// TESTS - ServiceInfo Structure
// ============================================================================

func TestServiceInfo_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		info    *ServiceInfo
		checkFn func(t *testing.T, info *ServiceInfo)
	}{
		{
			name: "all fields populated",
			info: &ServiceInfo{
				Name:     "Test Service",
				Icon:     "https://cdn.example.com/icon.svg",
				Type:     "test",
				Referrer: "test-service",
			},
			checkFn: func(t *testing.T, info *ServiceInfo) {
				assert.Equal(t, "Test Service", info.Name)
				assert.Equal(t, "https://cdn.example.com/icon.svg", info.Icon)
				assert.Equal(t, "test", info.Type)
				assert.Equal(t, "test-service", info.Referrer)
			},
		},
		{
			name: "minimal info",
			info: &ServiceInfo{
				Name:     "Minimal",
				Referrer: "minimal",
			},
			checkFn: func(t *testing.T, info *ServiceInfo) {
				assert.Equal(t, "Minimal", info.Name)
				assert.Empty(t, info.Icon)
				assert.Empty(t, info.Type)
				assert.Equal(t, "minimal", info.Referrer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.checkFn(t, tt.info)
		})
	}
}

// ============================================================================
// TESTS - Edge Cases
// ============================================================================

func TestDetectServiceFromReferrer_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		referrer  string
		expectNil bool
	}{
		{
			name:      "whitespace only",
			referrer:  "   ",
			expectNil: false, // Returns custom service
		},
		{
			name:      "very long referrer",
			referrer:  string(make([]byte, 10000)),
			expectNil: false,
		},
		{
			name:      "special characters",
			referrer:  "service-with-special-chars!@#$%",
			expectNil: false,
		},
		{
			name:      "unicode characters",
			referrer:  "服务-サービス",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := DetectServiceFromReferrer(tt.referrer)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.referrer, result.Referrer)
				// For unknown services, should get custom type
				if tt.referrer != "" {
					assert.Equal(t, "custom", result.Type)
				}
			}
		})
	}
}

// ============================================================================
// TESTS - Case Sensitivity
// ============================================================================

func TestDetectServiceFromReferrer_CaseSensitivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		referrer     string
		expectedType string
	}{
		{
			name:         "lowercase github",
			referrer:     "github",
			expectedType: "code",
		},
		{
			name:         "uppercase GITHUB (should be custom)",
			referrer:     "GITHUB",
			expectedType: "custom",
		},
		{
			name:         "mixed case GitHub (should be custom)",
			referrer:     "GitHub",
			expectedType: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := DetectServiceFromReferrer(tt.referrer)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType, result.Type)
		})
	}
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkDetectServiceFromReferrer(b *testing.B) {
	referrers := []string{
		"github",
		"google-docs",
		"notion",
		"slack",
		"unknown-service",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DetectServiceFromReferrer(referrers[i%len(referrers)])
	}
}

func BenchmarkDetectServiceFromReferrer_Parallel(b *testing.B) {
	referrers := []string{
		"github",
		"google-docs",
		"notion",
		"slack",
		"unknown-service",
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			DetectServiceFromReferrer(referrers[i%len(referrers)])
			i++
		}
	})
}

func BenchmarkDetectServiceFromReferrer_Known(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DetectServiceFromReferrer("github")
	}
}

func BenchmarkDetectServiceFromReferrer_Unknown(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DetectServiceFromReferrer("unknown-custom-service")
	}
}

func BenchmarkDetectServiceFromReferrer_Empty(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DetectServiceFromReferrer("")
	}
}
