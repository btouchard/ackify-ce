// SPDX-License-Identifier: AGPL-3.0-or-later
package logger

import (
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger

func init() {
	SetLevelAndFormat(slog.LevelInfo, "classic")
}

func SetLevel(level slog.Level) {
	SetLevelAndFormat(level, "classic")
}

func SetLevelAndFormat(level slog.Level, format string) {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "classic", "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		// Default to classic (text) format
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler)
}

func ParseLevel(levelStr string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
