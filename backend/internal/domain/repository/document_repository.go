package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// DocumentFilter carries optional filters for listing documents (Job 07).
type DocumentFilter struct {
	DocumentType string // filter by document type
	Status       string // filter by review status
	Page         int
	PerPage      int
}

// DocumentRepository defines persistence operations for document management (Job 07).
type DocumentRepository interface {
	Create(ctx context.Context, doc *entity.Document) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Document, error)
	// FindByThesisID returns the active (latest version per type) documents of a thesis.
	FindByThesisID(ctx context.Context, thesisID uuid.UUID, filter DocumentFilter) ([]*entity.Document, int64, error)
	// FindLatestByType returns the latest version of a specific document type
	// (optionally for a specific chapter). Returns gorm.ErrRecordNotFound if none.
	FindLatestByType(ctx context.Context, thesisID uuid.UUID, docType string, chapterNum *int) (*entity.Document, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewerID uuid.UUID, notes string) error
	// GetVersionHistory returns all versions of a document type, newest first.
	GetVersionHistory(ctx context.Context, thesisID uuid.UUID, docType string, chapterNum *int) ([]*entity.Document, error)
	// IsDocumentApproved reports whether the latest version of a document type is approved.
	IsDocumentApproved(ctx context.Context, thesisID uuid.UUID, docType string) (bool, error)
}
