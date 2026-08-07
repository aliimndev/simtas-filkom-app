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
		Preload("RequestedBy.Role").
		Preload("ReviewedBy.Role").
		Preload("CancelledBy.Role").
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

func (r *titleChangeRequestRepository) FindPendingBySupervisorID(ctx context.Context, supervisorID uuid.UUID) ([]*entity.TitleChangeRequest, error) {
	var reqs []*entity.TitleChangeRequest
	err := r.db.WithContext(ctx).
		Preload("RequestedBy.Role").
		Preload("ReviewedBy.Role").
		Preload("CancelledBy.Role").
		Joins("JOIN thesis_supervisors ts ON ts.thesis_id = title_change_requests.thesis_id").
		Where("ts.supervisor_id = ? AND title_change_requests.status = ?", supervisorID, "PENDING").
		Order("title_change_requests.created_at DESC").
		Find(&reqs).Error
	if err != nil {
		return nil, err
	}
	return reqs, nil
}

// Approve marks the request APPROVED and updates the thesis title atomically.
// The status guard (WHERE status = 'PENDING') keeps the transition safe against
// concurrent approve/reject/cancel: if zero rows are affected the request was
// already transitioned, so gorm.ErrRecordNotFound is returned.
func (r *titleChangeRequestRepository) Approve(ctx context.Context, id uuid.UUID, reviewedByID uuid.UUID, reviewedAt time.Time, reviewNotes *string, newTitle string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&entity.TitleChangeRequest{}).
			Where("id = ? AND status = ?", id, "PENDING").
			Updates(map[string]interface{}{
				"status":         "APPROVED",
				"reviewed_by_id": reviewedByID,
				"reviewed_at":    reviewedAt,
				"review_notes":   reviewNotes,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Model(&entity.Thesis{}).
			Where("id = (SELECT thesis_id FROM title_change_requests WHERE id = ?)", id).
			Update("title", newTitle).Error
	})
}

// UpdateStatus transitions a PENDING request to a terminal state (CANCELLED /
// REJECTED) with the associated review/cancel metadata. The status guard
// (WHERE status = 'PENDING') mirrors Approve so a concurrent approve/reject/
// cancel cannot flip an already-transitioned request: if zero rows are affected
// the request was already transitioned, so gorm.ErrRecordNotFound is returned.
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
	res := r.db.WithContext(ctx).Model(&entity.TitleChangeRequest{}).
		Where("id = ? AND status = ?", id, "PENDING").
		Updates(map[string]interface{}{
			"status":          status,
			"reviewed_by_id":  reviewedByID,
			"reviewed_at":     reviewedAt,
			"review_notes":    reviewNotes,
			"cancelled_by_id": cancelledByID,
			"cancelled_at":    cancelledAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
