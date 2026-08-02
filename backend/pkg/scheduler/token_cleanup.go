package scheduler

import (
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
)

const tokenCleanupInterval = 1 * time.Hour

// StartTokenCleanup runs a background goroutine that deletes expired token
// blacklist entries once per hour. It is intended to be launched after
// startup; the goroutine exits when the returned stop channel is closed.
func StartTokenCleanup(db *gorm.DB) (stop func()) {
	stopCh := make(chan struct{})
	ticker := time.NewTicker(tokenCleanupInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				result := db.Where("expires_at < ?", time.Now()).Delete(&entity.TokenBlacklist{})
				if result.Error != nil {
					logger.Get().Error("token blacklist cleanup failed", slog.String("error", result.Error.Error()))
					continue
				}
				logger.Get().Info("token blacklist cleanup done", slog.Int64("deleted", result.RowsAffected))
			case <-stopCh:
				return
			}
		}
	}()

	var once sync.Once
	return func() { once.Do(func() { close(stopCh) }) }
}
