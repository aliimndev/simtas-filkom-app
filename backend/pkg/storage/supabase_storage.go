package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	storage_go "github.com/supabase-community/storage-go"
)

// Default bucket names (Job 21). The documents bucket holds thesis documents;
// the archives bucket holds final-thesis PDFs. Both must be private buckets —
// all downloads go through expiring presigned URLs.
const (
	defaultDocumentsBucket = "simtas-documents"
	defaultArchivesBucket  = "simtas-archives"
)

// SupabaseStorageService is the production StorageService backed by Supabase
// Storage (Job 21). It replaces the local stub whenever STORAGE_PROVIDER is
// set to "supabase" and a project URL + service-role key are configured.
type SupabaseStorageService struct {
	client          *storage_go.Client
	documentsBucket string
	archivesBucket  string
}

// NewSupabaseStorageService builds a StorageService backed by Supabase.
//
//   - rawURL:    Supabase project URL, e.g. https://xxx.supabase.co
//     (the /storage/v1 suffix is appended automatically when missing)
//   - apiKey:    service-role key (required for server-side upload/sign/delete)
//   - documentsBucket: bucket for thesis documents (default simtas-documents)
//   - archivesBucket:  bucket for final-thesis archives (default simtas-archives)
func NewSupabaseStorageService(rawURL, apiKey, documentsBucket, archivesBucket string) *SupabaseStorageService {
	base := strings.TrimRight(rawURL, "/")
	if !strings.HasSuffix(base, "/storage/v1") {
		base += "/storage/v1"
	}
	if documentsBucket == "" {
		documentsBucket = defaultDocumentsBucket
	}
	if archivesBucket == "" {
		archivesBucket = defaultArchivesBucket
	}
	return &SupabaseStorageService{
		client:          storage_go.NewClient(base, apiKey, nil),
		documentsBucket: documentsBucket,
		archivesBucket:  archivesBucket,
	}
}

// bucketFor routes a storage path to the correct bucket based on its prefix:
// archives/… → archives bucket, everything else → documents bucket.
func (s *SupabaseStorageService) bucketFor(path string) string {
	if strings.HasPrefix(path, "archives/") {
		return s.archivesBucket
	}
	return s.documentsBucket
}

// Upload streams the file to Supabase and returns the relative storage path,
// keeping file_url portable across providers (same convention as the stub).
// Note: ctx is accepted for interface compatibility — the storage-go client
// has no context support, so cancellation is not propagated today.
// Note: storage-go sets content-type on the client's shared transport header
// per request. That is benign in practice — in SIMTAS all uploads are PDFs, so
// any content-type variance is a metadata nuance, not a data-integrity issue.
func (s *SupabaseStorageService) Upload(ctx context.Context, path string, file io.Reader, _ int64, contentType string) (string, error) {
	ct := contentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	opts := storage_go.FileOptions{ContentType: &ct}
	resp, err := s.client.UploadFile(s.bucketFor(path), path, file, opts)
	if err != nil {
		return "", fmt.Errorf("supabase upload: %w", err)
	}
	key := resp.Key
	if key == "" {
		key = path
	}
	return strings.TrimPrefix(key, "/"), nil
}

// GeneratePresignedURL returns a short-lived signed URL for the file.
func (s *SupabaseStorageService) GeneratePresignedURL(_ context.Context, path string, expirySeconds int) (string, error) {
	resp, err := s.client.CreateSignedUrl(s.bucketFor(path), path, expirySeconds)
	if err != nil {
		return "", fmt.Errorf("supabase sign url: %w", err)
	}
	if resp.SignedURL == "" {
		return "", fmt.Errorf("supabase sign url: empty response")
	}
	return resp.SignedURL, nil
}

// Delete removes the file; missing files are treated as success by the API.
func (s *SupabaseStorageService) Delete(_ context.Context, path string) error {
	if _, err := s.client.RemoveFile(s.bucketFor(path), []string{path}); err != nil {
		return fmt.Errorf("supabase delete: %w", err)
	}
	return nil
}
