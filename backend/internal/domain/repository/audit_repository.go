package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// AuditFilter narrows audit log queries (Job 13).
type AuditFilter struct {
	UserID     *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	PerPage    int
}

// AuditRepository defines persistence for audit log entries (Job 13 full module).
type AuditRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	// FindAll returns paginated audit logs matching the filter, newest first.
	FindAll(ctx context.Context, filter AuditFilter) ([]*entity.AuditLog, int64, error)
	// FindByEntity returns the full history of one entity, oldest first.
	FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*entity.AuditLog, error)
}
