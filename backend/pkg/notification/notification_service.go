package notification

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
)

// queueSize bounds the number of notification batches waiting to be written.
// When the queue is full the batch is dropped (and logged) instead of blocking
// the request or spawning an unbounded number of goroutines.
const queueSize = 512

// Params carries one notification event destined for a set of recipients.
type Params struct {
	UserIDs []uuid.UUID
	Title   string
	Message string
	Type    string
	Link    *string
}

// NotificationService writes in-app notifications for a batch of recipients.
// Writes go through a single bounded worker (same idiom as pkg/audit) so a
// burst of workflow events can never leak goroutines or pile up unbounded
// in-memory work.
type NotificationService struct {
	repo         domainRepo.NotificationRepository
	queue        chan Params
	wg           sync.WaitGroup
	shutdownOnce sync.Once
}

func NewNotificationService(repo domainRepo.NotificationRepository) *NotificationService {
	s := &NotificationService{
		repo:  repo,
		queue: make(chan Params, queueSize),
	}
	s.wg.Add(1)
	go s.worker()
	return s
}

// worker drains the queue and persists each notification batch. It runs for the
// lifetime of the service and is drained by Shutdown.
func (s *NotificationService) worker() {
	defer s.wg.Done()
	for params := range s.queue {
		if err := s.repo.CreateBatch(context.Background(), dedupe(params.UserIDs), params.Title, params.Message, params.Type, params.Link); err != nil {
			logger.Get().Error("failed to create notifications",
				slog.String("type", params.Type),
				slog.String("error", err.Error()),
			)
		}
	}
}

// Shutdown closes the queue channel and blocks until the worker has persisted
// every pending notification batch. It is safe to call multiple times.
func (s *NotificationService) Shutdown() {
	if s == nil {
		return
	}
	s.shutdownOnce.Do(func() {
		close(s.queue)
	})
	s.wg.Wait()
}

// Notify queues a notification batch to be written asynchronously so it never
// delays the response. If the queue is full the batch is dropped and logged.
// A nil service (e.g. in unit tests) is a no-op.
func (s *NotificationService) Notify(params Params) {
	if s == nil || s.repo == nil || len(params.UserIDs) == 0 {
		return
	}
	select {
	case s.queue <- params:
	default:
		logger.Get().Warn("notification queue full, dropping batch",
			slog.String("type", params.Type),
			slog.Int("recipients", len(params.UserIDs)),
		)
	}
}

func dedupe(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// Path builds an optional deep-link for a notification. The path is relative to
// the frontend root; the SPA resolves it with its own base URL.
func Path(format string, args ...interface{}) *string {
	s := fmt.Sprintf(format, args...)
	return &s
}
