package scheduler

import (
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
)

const (
	emailRetryInterval    = 2 * time.Minute
	emailRetryThreshold   = 2 * time.Minute // how long a "queued" row may be stuck before re-enqueueing
	emailRetryBatchSize   = 200
	emailMaxSchedulerRuns = 5 // dead-letter cap: rows failed more than this many cycles stay failed
)

// StartEmailRetry runs a background goroutine that re-enqueues undelivered
// emails into the worker pool. It covers two cases that in-process retries
// cannot: rows left "queued" by a crash mid-delivery, and "failed" rows that
// may succeed once the provider recovers. Rows are bounded by an attempt cap so
// a permanently failing address is not retried forever. The goroutine exits
// when the returned stop function is called.
func StartEmailRetry(db *gorm.DB, svc *email.ResendEmailService) (stop func()) {
	stopCh := make(chan struct{})
	ticker := time.NewTicker(emailRetryInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				retryPendingEmails(db, svc)
			case <-stopCh:
				return
			}
		}
	}()

	var once sync.Once
	return func() { once.Do(func() { close(stopCh) }) }
}

// retryPendingEmails finds email_logs rows that still need delivery and
// re-enqueues them one at a time. Only rows with a persisted body can be retried.
func retryPendingEmails(db *gorm.DB, svc *email.ResendEmailService) {
	threshold := time.Now().Add(-emailRetryThreshold)
	var logs []entity.EmailLog
	err := db.Where(
		"(status = 'queued' AND created_at < ?) OR (status = 'failed' AND attempts < ?)",
		threshold, emailMaxSchedulerRuns,
	).
		Order("created_at ASC").
		Limit(emailRetryBatchSize).
		Find(&logs).Error
	if err != nil {
		logger.Get().Error("email retry query failed", slog.String("error", err.Error()))
		return
	}
	if len(logs) == 0 {
		return
	}

	retried := 0
	for _, l := range logs {
		if l.Body == nil || *l.Body == "" {
			continue
		}
		subject := ""
		if l.Subject != nil {
			subject = *l.Subject
		}
		svc.Retry(l.ID, l.RecipientEmail, subject, *l.Body, l.EventType)
		retried++
	}
	logger.Get().Info("email retry cycle done", slog.Int("re-enqueued", retried))
}
