package utils

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
)

// File validation errors (Job 07).
var (
	ErrNotPDF          = errors.New("hanya file PDF yang diizinkan")
	ErrFileTooLarge    = errors.New("ukuran file melebihi batas maksimum")
	ErrFileEmpty       = errors.New("file tidak boleh kosong")
	ErrInvalidFileType = errors.New("tipe file tidak sesuai")
)

// ValidatePDF checks that the uploaded file is a PDF within the size limit.
// It validates the extension, the MIME type (sniffed from the content), and
// the declared size.
func ValidatePDF(file multipart.File, header *multipart.FileHeader, maxSize int64) error {
	if header == nil || header.Size <= 0 {
		return ErrFileEmpty
	}
	if header.Size > maxSize {
		return ErrFileTooLarge
	}
	if !strings.EqualFold(filepath.Ext(header.Filename), ".pdf") {
		return ErrNotPDF
	}

	// Sniff the magic bytes to reject renamed files (must start with %PDF-).
	// A genuine PDF always begins with the 4-byte %PDF signature.
	if file != nil {
		buf := make([]byte, 4)
		n, _ := file.Read(buf)
		if n < 4 || string(buf[:4]) != "%PDF" {
			return ErrNotPDF
		}
		// Rewind so the caller can re-read the full content (multipart.File is a Seeker).
		if seeker, ok := file.(interface {
			Seek(offset int64, whence int) (int64, error)
		}); ok {
			_, _ = seeker.Seek(0, 0)
		}
	}
	return nil
}

// GetFileSizeInMB converts a byte count to megabytes (2 decimals).
func GetFileSizeInMB(size int64) float64 {
	return float64(size) / (1024 * 1024)
}
