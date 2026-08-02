package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

// AuditLogDetail is the API response shape for an audit log entry (Job 13).
type AuditLogDetail struct {
	ID         uuid.UUID       `json:"id"`
	User       *AuditUserBrief `json:"user,omitempty"`
	Action     string          `json:"action"`
	EntityType *string         `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID      `json:"entity_id,omitempty"`
	OldValue   interface{}     `json:"old_value,omitempty"`
	NewValue   interface{}     `json:"new_value,omitempty"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	UserAgent  *string         `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AuditUserBrief struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
}

// AuditUseCase exposes audit log queries for admin monitoring (Job 13).
type AuditUseCase struct {
	repo domainRepo.AuditRepository
}

func NewAuditUseCase(repo domainRepo.AuditRepository) *AuditUseCase {
	return &AuditUseCase{repo: repo}
}

// List returns paginated audit logs matching the filter, newest first.
func (uc *AuditUseCase) List(ctx context.Context, filter domainRepo.AuditFilter) ([]*AuditLogDetail, int64, error) {
	logs, total, err := uc.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	details := make([]*AuditLogDetail, 0, len(logs))
	for _, l := range logs {
		details = append(details, toAuditLogDetail(l))
	}
	return details, total, nil
}

// ByEntity returns the complete history of one entity, oldest first.
func (uc *AuditUseCase) ByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*AuditLogDetail, error) {
	logs, err := uc.repo.FindByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}
	details := make([]*AuditLogDetail, 0, len(logs))
	for _, l := range logs {
		details = append(details, toAuditLogDetail(l))
	}
	return details, nil
}

func toAuditLogDetail(l *entity.AuditLog) *AuditLogDetail {
	d := &AuditLogDetail{
		ID:         l.ID,
		Action:     l.Action,
		EntityType: l.EntityType,
		EntityID:   l.EntityID,
		OldValue:   l.OldValue,
		NewValue:   l.NewValue,
		IPAddress:  l.IPAddress,
		UserAgent:  l.UserAgent,
		CreatedAt:  l.CreatedAt,
	}
	if l.User != nil {
		d.User = &AuditUserBrief{ID: l.User.ID, FullName: l.User.FullName, Email: l.User.Email}
	}
	return d
}
