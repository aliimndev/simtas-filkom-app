package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

var ErrNotificationNotFound = errors.New("notifikasi tidak ditemukan")

// NotificationDetail is the API response shape for an in-app notification.
type NotificationDetail struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Type      string     `json:"type"`
	Link      *string    `json:"link,omitempty"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// NotificationUseCase exposes read + read-state operations for the current user.
type NotificationUseCase struct {
	repo domainRepo.NotificationRepository
}

func NewNotificationUseCase(repo domainRepo.NotificationRepository) *NotificationUseCase {
	return &NotificationUseCase{repo: repo}
}

// List returns the user's newest notifications (newest first).
func (uc *NotificationUseCase) List(ctx context.Context, userID uuid.UUID, limit int) ([]*NotificationDetail, error) {
	notifs, err := uc.repo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*NotificationDetail, 0, len(notifs))
	for _, n := range notifs {
		out = append(out, toNotificationDetail(n))
	}
	return out, nil
}

// UnreadCount returns how many notifications the user has not read yet.
func (uc *NotificationUseCase) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return uc.repo.UnreadCount(ctx, userID)
}

// MarkRead marks one of the user's notifications as read.
func (uc *NotificationUseCase) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	if err := uc.repo.MarkRead(ctx, userID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return err
	}
	return nil
}

// MarkAllRead marks every unread notification of the user as read.
func (uc *NotificationUseCase) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return uc.repo.MarkAllRead(ctx, userID)
}

func toNotificationDetail(n *entity.Notification) *NotificationDetail {
	return &NotificationDetail{
		ID:        n.ID,
		Title:     n.Title,
		Message:   n.Message,
		Type:      n.Type,
		Link:      n.Link,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}
