package repository

import (
	"context"

	"github.com/aliimndev/simtas-filkom-app/api/internal/domain/entity"
)

// AuditRepository defines persistence for audit log entries (Job 04 minimal).
type AuditRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
}
