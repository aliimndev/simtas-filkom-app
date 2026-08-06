package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

type TitleChangeRequestRepository interface {
	Create(ctx context.Context, req *entity.TitleChangeRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.TitleChangeRequest, error)
	FindByThesisID(ctx context.Context, thesisID uuid.UUID) ([]*entity.TitleChangeRequest, error)
	FindPendingByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.TitleChangeRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewedByID *uuid.UUID, reviewedAt *time.Time, reviewNotes *string, cancelledByID *uuid.UUID, cancelledAt *time.Time) error
}
