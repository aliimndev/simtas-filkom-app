package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3StorageService implements StorageService using any S3-compatible backend
// (MinIO, Cloudflare R2, Backblaze B2, etc.).
type S3StorageService struct {
	client           *minio.Client
	documentsBucket  string
	archivesBucket   string
}

// NewS3StorageService builds an S3-backed storage service.
func NewS3StorageService(endpoint, region, accessKey, secretKey, documentsBucket, archivesBucket string, usePathStyle bool) (*S3StorageService, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: strings.HasPrefix(endpoint, "https"),
	}
	if usePathStyle {
		opts.BucketLookup = minio.BucketLookupPath
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Verify buckets exist (create if missing)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, bucket := range []string{documentsBucket, archivesBucket} {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to check bucket %s: %w", bucket, err)
		}
		if !exists {
			if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
				return nil, fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
	}

	return &S3StorageService{
		client:          client,
		documentsBucket: documentsBucket,
		archivesBucket:  archivesBucket,
	}, nil
}

func (s *S3StorageService) bucketFor(path string) string {
	if strings.HasPrefix(path, "archives/") {
		return s.archivesBucket
	}
	return s.documentsBucket
}

func (s *S3StorageService) Upload(ctx context.Context, path string, file io.Reader, size int64, contentType string) (string, error) {
	bucket := s.bucketFor(path)
	_, err := s.client.PutObject(ctx, bucket, path, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload: %w", err)
	}
	return strings.TrimPrefix(path, "/"), nil
}

func (s *S3StorageService) GeneratePresignedURL(ctx context.Context, path string, expirySeconds int) (string, error) {
	bucket := s.bucketFor(path)
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, path, time.Duration(expirySeconds)*time.Second, reqParams)
	if err != nil {
		return "", fmt.Errorf("s3 presign: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *S3StorageService) Delete(ctx context.Context, path string) error {
	bucket := s.bucketFor(path)
	if err := s.client.RemoveObject(ctx, bucket, path, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	return nil
}