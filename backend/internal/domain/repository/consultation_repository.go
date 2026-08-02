package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// ConsultationFilter carries optional filters for listing consultation logs (Job 06).
type ConsultationFilter struct {
	Status   string // filter by log status: pending | approved
	DateFrom string // YYYY-MM-DD, inclusive lower bound on consultation_date
	DateTo   string // YYYY-MM-DD, inclusive upper bound on consultation_date
	Page     int
	PerPage  int
}

// ConsultationRepository defines persistence operations for consultation logs (Job 06).
type ConsultationRepository interface {
	Create(ctx context.Context, log *entity.ConsultationLog) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ConsultationLog, error)
	FindByThesisID(ctx context.Context, thesisID uuid.UUID, filter ConsultationFilter) ([]*entity.ConsultationLog, int64, error)
	Update(ctx context.Context, log *entity.ConsultationLog) error
	Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error
	CountApprovedByThesisID(ctx context.Context, thesisID uuid.UUID) (int, error)
	// Delete hard-deletes a consultation log that has not been approved yet.
	Delete(ctx context.Context, id uuid.UUID) error
}
