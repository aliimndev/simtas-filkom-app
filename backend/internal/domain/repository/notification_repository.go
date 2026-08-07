package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// NotificationRepository defines persistence for in-app notifications.
type NotificationRepository interface {
	// CreateBatch inserts one notification row per recipient in a single statement.
	CreateBatch(ctx context.Context, userIDs []uuid.UUID, title, message, ntype string, link *string) error
	// ListByUser returns the newest notifications for a user (newest first).
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.Notification, error)
	// UnreadCount returns how many unread notifications a user has.
	UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkRead marks a single notification as read, scoped to its owner.
	// Returns gorm.ErrRecordNotFound when the notification does not exist.
	MarkRead(ctx context.Context, userID, id uuid.UUID) error
	// MarkAllRead marks every unread notification of a user as read.
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
}
