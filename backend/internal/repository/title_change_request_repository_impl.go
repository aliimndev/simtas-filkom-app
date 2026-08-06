package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type titleChangeRequestRepository struct {
	db *gorm.DB
}

func NewTitleChangeRequestRepository(db *gorm.DB) domainRepo.TitleChangeRequestRepository {
	return &titleChangeRequestRepository{db: db}
}

func (r *titleChangeRequestRepository) Create(ctx context.Context, req *entity.TitleChangeRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *titleChangeRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.TitleChangeRequest, error) {
	var req entity.TitleChangeRequest
	err := r.db.WithContext(ctx).
		Preload("Thesis.Student.Role").
		Preload("Thesis.AcademicYear").
		Preload("Thesis.Supervisors.Role").
		Preload("RequestedBy.Role").
		Preload("ReviewedBy.Role").
		Preload("CancelledBy.Role").
		Where("id = ?", id).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *titleChangeRequestRepository) FindByThesisID(ctx context.Context, thesisID uuid.UUID) ([]*entity.TitleChangeRequest, error) {
	var reqs []*entity.TitleChangeRequest
	err := r.db.WithContext(ctx).
		Where("thesis_id = ?", thesisID).
		Order("created_at DESC").
		Find(&reqs).Error
	if err != nil {
		return nil, err
	}
	return reqs, nil
}

func (r *titleChangeRequestRepository) FindPendingByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.TitleChangeRequest, error) {
	var req entity.TitleChangeRequest
	err := r.db.WithContext(ctx).
		Where("thesis_id = ? AND status = ?", thesisID, "PENDING").
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *titleChangeRequestRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status string,
	reviewedByID *uuid.UUID,
	reviewedAt *time.Time,
	reviewNotes *string,
	cancelledByID *uuid.UUID,
	cancelledAt *time.Time,
) error {
	return r.db.WithContext(ctx).Model(&entity.TitleChangeRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         status,
			"reviewed_by":    reviewedByID,
			"reviewed_at":    reviewedAt,
			"review_notes":   reviewNotes,
			"cancelled_by":   cancelledByID,
			"cancelled_at":   cancelledAt,
		}).Error
}
