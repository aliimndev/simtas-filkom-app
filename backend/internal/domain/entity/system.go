package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User       *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action     string         `gorm:"type:varchar(100);not null" json:"action"`
	EntityType *string        `gorm:"type:varchar(50)" json:"entity_type,omitempty"`
	EntityID   *uuid.UUID     `gorm:"type:uuid" json:"entity_id,omitempty"`
	OldValue   datatypes.JSON `gorm:"type:jsonb" json:"old_value,omitempty"`
	NewValue   datatypes.JSON `gorm:"type:jsonb" json:"new_value,omitempty"`
	IPAddress  *string        `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent  *string        `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt  time.Time      `gorm:"not null;default:now();index" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type EmailLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipientEmail string    `gorm:"type:varchar(255);not null" json:"recipient_email"`
	EventType      string    `gorm:"type:varchar(100);not null" json:"event_type"`
	Subject        *string   `gorm:"type:varchar(500)" json:"subject,omitempty"`
	Status         string    `gorm:"type:varchar(20);not null;default:sent" json:"status"`
	Provider       string    `gorm:"type:varchar(50);not null;default:resend" json:"provider"`
	ErrorMessage   *string   `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (EmailLog) TableName() string { return "email_logs" }
