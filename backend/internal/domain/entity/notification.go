package entity

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Title     string     `gorm:"type:varchar(255);not null" json:"title"`
	Message   string     `gorm:"type:text;not null" json:"message"`
	Type      string     `gorm:"type:varchar(50);not null" json:"type"`
	Link      *string    `gorm:"type:varchar(500)" json:"link,omitempty"`
	IsRead    bool       `gorm:"not null;default:false;index" json:"is_read"`
	ReadAt    *time.Time `gorm:"index" json:"read_at,omitempty"`
	CreatedAt time.Time  `gorm:"not null;default:now();index" json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }
