package audit

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

// AuditParams carries everything needed to write one audit log entry.
type AuditParams struct {
	UserID     *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	OldValue   interface{}
	NewValue   interface{}
	IPAddress  *string
	UserAgent  *string
}

// AuditService writes audit log entries. Job 13 expands this into the full
// audit module; this minimal version keeps the trail complete from Job 04.
type AuditService struct {
	repo domainRepo.AuditRepository
}

func NewAuditService(repo domainRepo.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log writes an audit entry asynchronously so it never delays the response.
func (s *AuditService) Log(ctx context.Context, params AuditParams) {
	if s == nil || s.repo == nil {
		return
	}

	go func() {
		entry := &entity.AuditLog{
			UserID:     params.UserID,
			Action:     params.Action,
			EntityType: nullableString(params.EntityType),
			EntityID:   params.EntityID,
			OldValue:   toJSON(params.OldValue),
			NewValue:   toJSON(params.NewValue),
			IPAddress:  params.IPAddress,
			UserAgent:  params.UserAgent,
		}
		if err := s.repo.Create(context.Background(), entry); err != nil {
			log.Printf("audit: failed to write log action=%s: %v", params.Action, err)
		}
	}()
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return datatypes.JSON(b)
}
