package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) domainRepo.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindAll returns paginated audit logs filtered by the given criteria, newest
// first (Job 13).
func (r *auditRepository) FindAll(ctx context.Context, filter domainRepo.AuditFilter) ([]*entity.AuditLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.AuditLog{})

	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.EntityType != "" {
		q = q.Where("entity_type = ?", filter.EntityType)
	}
	if filter.EntityID != nil {
		q = q.Where("entity_id = ?", *filter.EntityID)
	}
	if filter.DateFrom != nil {
		q = q.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		q = q.Where("created_at <= ?", *filter.DateTo)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}

	var logs []*entity.AuditLog
	err := q.Preload("User").
		Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// FindByEntity returns the complete audit history of one entity, oldest first
// (useful for tracing the full lifecycle of a thesis).
func (r *auditRepository) FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*entity.AuditLog, error) {
	var logs []*entity.AuditLog
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at ASC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}
