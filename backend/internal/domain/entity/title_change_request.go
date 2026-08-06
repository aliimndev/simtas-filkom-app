package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TitleChangeRequest struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThesisID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"thesis_id"`
	Thesis         Thesis         `gorm:"foreignKey:ThesisID" json:"thesis,omitempty"`
	RequestedByID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"requested_by_id"`
	RequestedBy    User           `gorm:"foreignKey:RequestedByID" json:"requested_by,omitempty"`
	PreviousTitle  string         `gorm:"type:text;not null" json:"previous_title"`
	RequestedTitle string         `gorm:"type:text;not null" json:"requested_title"`
	Status         string         `gorm:"type:varchar(20);not null;default:PENDING" json:"status"`
	ReviewedByID   *uuid.UUID     `gorm:"type:uuid;index" json:"reviewed_by_id,omitempty"`
	ReviewedBy     *User          `gorm:"foreignKey:ReviewedByID" json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time     `json:"reviewed_at,omitempty"`
	ReviewNotes    *string        `gorm:"type:text" json:"review_notes,omitempty"`
	CancelledByID  *uuid.UUID     `gorm:"type:uuid;index" json:"cancelled_by_id,omitempty"`
	CancelledBy    *User          `gorm:"foreignKey:CancelledByID" json:"cancelled_by,omitempty"`
	CancelledAt    *time.Time     `json:"cancelled_at,omitempty"`
	CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TitleChangeRequest) TableName() string { return "title_change_requests" }
