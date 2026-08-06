package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domainRepo.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) CreateBatch(ctx context.Context, userIDs []uuid.UUID, title, message, ntype string, link *string) error {
	if len(userIDs) == 0 {
		return nil
	}
	rows := make([]entity.Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		rows = append(rows, entity.Notification{
			UserID:  uid,
			Title:   title,
			Message: message,
			Type:    ntype,
			Link:    link,
		})
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.Notification, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var notifs []*entity.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifs).Error
	return notifs, err
}

func (r *notificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND id = ?", userID, id).
		Updates(map[string]interface{}{"is_read": true, "read_at": &now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": &now}).Error
}
