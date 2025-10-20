// SPDX-License-Identifier: AGPL-3.0-or-later
package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TESTS - ParseLevel
// ============================================================================

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{
			name:     "debug lowercase",
			input:    "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "debug uppercase",
			input:    "DEBUG",
			expected: slog.LevelDebug,
		},
		{
			name:     "debug mixed case",
			input:    "DeBuG",
			expected: slog.LevelDebug,
		},
		{
			name:     "info lowercase",
			input:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "info uppercase",
			input:    "INFO",
			expected: slog.LevelInfo,
		},
		{
			name:     "warn lowercase",
			input:    "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "warn uppercase",
			input:    "WARN",
			expected: slog.LevelWarn,
		},
		{
			name:     "warning lowercase",
			input:    "warning",
			expected: slog.LevelWarn,
		},
		{
			name:     "warning uppercase",
			input:    "WARNING",
			expected: slog.LevelWarn,
		},
		{
			name:     "error lowercase",
			input:    "error",
			expected: slog.LevelError,
		},
		{
			name:     "error uppercase",
			input:    "ERROR",
			expected: slog.LevelError,
		},
		{
			name:     "unknown level defaults to info",
			input:    "unknown",
			expected: slog.LevelInfo,
		},
		{
			name:     "empty string defaults to info",
			input:    "",
			expected: slog.LevelInfo,
		},
		{
			name:     "whitespace only defaults to info",
			input:    "   ",
			expected: slog.LevelInfo,
		},
		{
			name:     "input with leading whitespace",
			input:    "  debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "input with trailing whitespace",
			input:    "error  ",
			expected: slog.LevelError,
		},
		{
			name:     "input with surrounding whitespace",
			input:    "  warn  ",
			expected: slog.LevelWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ParseLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// TESTS - SetLevel
// ============================================================================

func TestSetLevel(t *testing.T) {
	// Cannot run in parallel as it modifies global Logger state
	// t.Parallel()

	tests := []struct {
		name  string
		level slog.Level
	}{
		{
			name:  "set debug level",
			level: slog.LevelDebug,
		},
		{
			name:  "set info level",
			level: slog.LevelInfo,
		},
		{
			name:  "set warn level",
			level: slog.LevelWarn,
		},
		{
			name:  "set error level",
			level: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot run in parallel as it modifies global Logger
			// t.Parallel()

			SetLevel(tt.level)

			require.NotNil(t, Logger, "Logger should be initialized")
			assert.True(t, Logger.Enabled(nil, tt.level), "Logger should be enabled for the set level")
		})
	}
}

// ============================================================================
// TESTS - Init
// ============================================================================

func TestLogger_Initialization(t *testing.T) {
	// Test that the logger is initialized on package import
	// The init() function should have set Logger to some level
	// We just verify it's not nil since other tests may have changed the level

	require.NotNil(t, Logger, "Logger should be initialized by init()")
	// Note: We don't test the specific level here as other tests may modify it
}

// ============================================================================
// TESTS - Integration
// ============================================================================

func TestParseLevel_Integration(t *testing.T) {
	// Cannot run in parallel as SetLevel modifies global state
	// t.Parallel()

	// Test that ParseLevel output can be used with SetLevel
	tests := []struct {
		levelStr string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
	}

	for _, tt := range tests {
		t.Run("parse_and_set_"+tt.levelStr, func(t *testing.T) {
			// Cannot run in parallel as it modifies global state
			// t.Parallel()

			level := ParseLevel(tt.levelStr)
			assert.Equal(t, tt.expected, level)

			SetLevel(level)
			require.NotNil(t, Logger)
			assert.True(t, Logger.Enabled(nil, level))
		})
	}
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkParseLevel(b *testing.B) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ParseLevel(levels[i%len(levels)])
	}
}

func BenchmarkParseLevel_Parallel(b *testing.B) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ParseLevel(levels[i%len(levels)])
			i++
		}
	})
}

func BenchmarkSetLevel(b *testing.B) {
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SetLevel(levels[i%len(levels)])
	}
}
