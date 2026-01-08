// SPDX-License-Identifier: AGPL-3.0-or-later
package storage

import (
	"context"
	"io"
)

type Provider interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, int64, string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Type() string
}

type FileInfo struct {
	Key         string
	Size        int64
	ContentType string
}

var AllowedMIMETypes = map[string]bool{
	"application/pdf":    true,
	"image/png":          true,
	"image/jpeg":         true,
	"image/gif":          true,
	"image/webp":         true,
	"image/svg+xml":      true,
	"text/plain":         true,
	"text/html":          true,
	"text/markdown":      true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

func IsAllowedMIMEType(mimeType string) bool {
	return AllowedMIMETypes[mimeType]
}
