package repository

import (
	"context"
	"time"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/google/uuid"
)

type TitleChangeRequestRepository interface {
	Create(ctx context.Context, req *entity.TitleChangeRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.TitleChangeRequest, error)
	FindByThesisID(ctx context.Context, thesisID uuid.UUID) ([]*entity.TitleChangeRequest, error)
	FindPendingByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.TitleChangeRequest, error)
	// FindPendingBySupervisorID returns PENDING requests for theses supervised
	// by supervisorID (the dosen pembimbing review queue).
	FindPendingBySupervisorID(ctx context.Context, supervisorID uuid.UUID) ([]*entity.TitleChangeRequest, error)
	// UpdateStatus transitions a PENDING request to a terminal state (CANCELLED /
	// REJECTED). It returns gorm.ErrRecordNotFound when the request is no longer
	// PENDING (a concurrent approve/reject/cancel won the race).
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewedByID *uuid.UUID, reviewedAt *time.Time, reviewNotes *string, cancelledByID *uuid.UUID, cancelledAt *time.Time) error
	// Approve atomically marks the request APPROVED (with review metadata) and
	// updates the thesis title in a single database transaction.
	//
	// It returns gorm.ErrRecordNotFound when the request is no longer PENDING
	// (a concurrent state change won the race).
	Approve(ctx context.Context, id uuid.UUID, reviewedByID uuid.UUID, reviewedAt time.Time, reviewNotes *string, newTitle string) error
}
