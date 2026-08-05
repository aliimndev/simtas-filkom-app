package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
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

// auditQueueSize bounds the number of audit entries waiting to be written.
// When the queue is full the entry is dropped (and logged) instead of blocking
// the request or spawning an unbounded number of goroutines.
const auditQueueSize = 256

// AuditService writes audit log entries. Job 13 expands this into the full
// audit module; this minimal version keeps the trail complete from Job 04.
// Writes go through a single bounded worker so a burst of actions can never
// leak goroutines or pile up unbounded in-memory work.
type AuditService struct {
	repo           domainRepo.AuditRepository
	queue          chan AuditParams
	wg             sync.WaitGroup
	shutdownOnce   sync.Once
}

func NewAuditService(repo domainRepo.AuditRepository) *AuditService {
	s := &AuditService{
		repo:  repo,
		queue: make(chan AuditParams, auditQueueSize),
	}
	s.wg.Add(1)
	go s.worker()
	return s
}

// worker drains the queue and persists each entry. It runs for the lifetime of
// the service (same pattern as the rate-limiter cleanup goroutine).
func (s *AuditService) worker() {
	defer s.wg.Done()
	for params := range s.queue {
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
			logger.Get().Error("failed to write audit log",
				slog.String("action", params.Action),
				slog.String("error", err.Error()),
			)
		}
	}
}

// Shutdown closes the queue channel and blocks until the worker goroutine
// has drained every pending audit entry. It is safe to call multiple times.
func (s *AuditService) Shutdown() {
	s.shutdownOnce.Do(func() {
		close(s.queue)
	})
	s.wg.Wait()
}

// Log queues an audit entry to be written asynchronously so it never delays
// the response. If the queue is full the entry is dropped and logged.
func (s *AuditService) Log(ctx context.Context, params AuditParams) {
	if s == nil || s.repo == nil {
		return
	}

	select {
	case s.queue <- params:
	default:
		logger.Get().Warn("audit queue full, dropping entry",
			slog.String("action", params.Action),
		)
	}
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
