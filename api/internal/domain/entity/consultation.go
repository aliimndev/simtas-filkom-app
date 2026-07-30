package entity

import (
	"time"

	"github.com/google/uuid"
)

type ConsultationLog struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThesisID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"thesis_id"`
	Thesis           Thesis     `gorm:"foreignKey:ThesisID" json:"-"`
	CreatedBy        uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	Creator          User       `gorm:"foreignKey:CreatedBy" json:"creator"`
	ConsultationDate time.Time  `gorm:"type:date;not null" json:"consultation_date"`
	TopicsDiscussed  string     `gorm:"type:text;not null" json:"topics_discussed"`
	Notes            *string    `gorm:"type:text" json:"notes,omitempty"`
	FollowUp         *string    `gorm:"type:text" json:"follow_up,omitempty"`
	AttachmentURL    *string    `gorm:"type:text" json:"attachment_url,omitempty"`
	Status           string     `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	ApprovedBy       *uuid.UUID `gorm:"type:uuid" json:"approved_by,omitempty"`
	Approver         *User      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	CreatedAt        time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

func (ConsultationLog) TableName() string { return "consultation_logs" }
