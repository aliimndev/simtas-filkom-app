package service

import (
	"context"
	"io"
)

// StorageService abstracts file storage (Job 07). Job 21 replaces the local
// stub with a real Supabase implementation without changing this interface.
type StorageService interface {
	// Upload stores the file content at the given path and returns the
	// relative storage path (used as file_url in the database).
	Upload(ctx context.Context, path string, file io.Reader, size int64, contentType string) (string, error)
	// GeneratePresignedURL returns a temporary URL to download the file.
	GeneratePresignedURL(ctx context.Context, path string, expirySeconds int) (string, error)
	// Delete removes the file at the given path.
	Delete(ctx context.Context, path string) error
}
