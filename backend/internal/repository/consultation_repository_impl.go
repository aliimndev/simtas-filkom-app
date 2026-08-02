package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type consultationRepository struct {
	db *gorm.DB
}

func NewConsultationRepository(db *gorm.DB) domainRepo.ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) Create(ctx context.Context, log *entity.ConsultationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *consultationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ConsultationLog, error) {
	var log entity.ConsultationLog
	err := r.db.WithContext(ctx).
		Preload("Creator.Role").
		Preload("Approver.Role").
		Where("id = ?", id).
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *consultationRepository) FindByThesisID(ctx context.Context, thesisID uuid.UUID, filter domainRepo.ConsultationFilter) ([]*entity.ConsultationLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.ConsultationLog{})

	q = q.Where("consultation_logs.thesis_id = ?", thesisID)
	if filter.Status != "" {
		q = q.Where("consultation_logs.status = ?", filter.Status)
	}
	if filter.DateFrom != "" {
		q = q.Where("consultation_logs.consultation_date >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		q = q.Where("consultation_logs.consultation_date <= ?", filter.DateTo)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// PerPage == 0 means "no pagination" (used by the summary aggregation).
	q = q.Preload("Creator.Role").
		Preload("Approver.Role").
		Order("consultation_logs.consultation_date DESC, consultation_logs.created_at DESC")
	if filter.PerPage > 0 {
		page, perPage := filter.Page, filter.PerPage
		if page < 1 {
			page = 1
		}
		if perPage > 100 {
			perPage = 100
		}
		q = q.Offset((page - 1) * perPage).Limit(perPage)
	}

	var logs []*entity.ConsultationLog
	if err := q.Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *consultationRepository) Update(ctx context.Context, log *entity.ConsultationLog) error {
	return r.db.WithContext(ctx).Model(&entity.ConsultationLog{}).
		Where("id = ?", log.ID).
		Updates(map[string]interface{}{
			"consultation_date": log.ConsultationDate,
			"topics_discussed":  log.TopicsDiscussed,
			"notes":             log.Notes,
			"follow_up":         log.FollowUp,
			"attachment_url":    log.AttachmentURL,
			"updated_at":        time.Now(),
		}).Error
}

func (r *consultationRepository) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.ConsultationLog{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(map[string]interface{}{
			"status":      "approved",
			"approved_by": approvedBy,
			"approved_at": time.Now(),
		}).Error
}

func (r *consultationRepository) CountApprovedByThesisID(ctx context.Context, thesisID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ConsultationLog{}).
		Where("thesis_id = ? AND status = ?", thesisID, "approved").
		Count(&count).Error
	return int(count), err
}

func (r *consultationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.ConsultationLog{}, "id = ?", id).Error
}
