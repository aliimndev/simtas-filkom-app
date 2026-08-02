package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// StubStorageService is a development-only StorageService that writes files to
// a local folder (./tmp/uploads by default) and serves them back as local
// URLs. It is used until Job 21 swaps in the real Supabase implementation.
type StubStorageService struct {
	baseDir string
	baseURL string
}

// NewStubStorageService builds a local storage stub.
//   - baseDir: absolute or relative folder where files are stored (default ./tmp/uploads)
//   - baseURL: public prefix used when generating download URLs (default http://localhost:8080)
func NewStubStorageService(baseDir, baseURL string) *StubStorageService {
	if baseDir == "" {
		baseDir = "./tmp/uploads"
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &StubStorageService{baseDir: baseDir, baseURL: baseURL}
}

// Upload writes the reader to baseDir/path and returns the relative path.
func (s *StubStorageService) Upload(ctx context.Context, path string, file io.Reader, size int64, contentType string) (string, error) {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	// Return the path without a leading slash so it can be stored as file_url.
	return strings.TrimPrefix(filepath.ToSlash(path), "/"), nil
}

// GeneratePresignedURL returns a local URL pointing at the stored file.
// The stub does not expire URLs; the expirySeconds argument is accepted for
// interface compatibility.
func (s *StubStorageService) GeneratePresignedURL(_ context.Context, path string, _ int) (string, error) {
	return s.baseURL + "/tmp/uploads/" + strings.TrimPrefix(filepath.ToSlash(path), "/"), nil
}

// Delete removes the file; missing files are treated as success.
func (s *StubStorageService) Delete(_ context.Context, path string) error {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(path))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}
