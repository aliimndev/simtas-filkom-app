package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email              string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash       string         `gorm:"type:varchar(255);not null" json:"-"`
	FullName           string         `gorm:"type:varchar(255);not null" json:"full_name"`
	NimNidn            *string        `gorm:"type:varchar(50)" json:"nim_nidn,omitempty"`
	RoleID             int            `gorm:"not null" json:"role_id"`
	Role               Role           `gorm:"foreignKey:RoleID" json:"role"`
	StudyProgram       *string        `gorm:"type:varchar(100)" json:"study_program,omitempty"`
	ProfilePhotoURL    *string        `gorm:"type:text" json:"profile_photo_url,omitempty"`
	IsActive           bool           `gorm:"not null;default:true" json:"is_active"`
	MustChangePassword bool           `gorm:"not null;default:true" json:"must_change_password"`
	LoginAttemptCount  int            `gorm:"not null;default:0" json:"-"`
	LockedUntil        *time.Time     `json:"-"`
	LastLoginAt        *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt          time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }
