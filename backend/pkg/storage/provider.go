// SPDX-License-Identifier: AGPL-3.0-or-later
package storage

import (
	"context"
	"io"
	"mime"
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
	"text/x-markdown":    true, // Alternative MIME type for markdown
	"application/msword": true, // .doc
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
	"application/vnd.ms-excel": true, // .xls
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx
	"application/vnd.oasis.opendocument.text":                           true, // .odt
	"application/vnd.oasis.opendocument.spreadsheet":                    true, // .ods
}

func IsAllowedMIMEType(mimeType string) bool {
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return AllowedMIMETypes[mimeType]
	}
	return AllowedMIMETypes[mediaType]
}
