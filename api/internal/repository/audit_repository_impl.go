package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/api/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/api/internal/domain/repository"
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
