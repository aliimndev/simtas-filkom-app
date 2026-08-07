package entity

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	Token     string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }

type TokenBlacklist struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TokenJTI  string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (TokenBlacklist) TableName() string { return "token_blacklist" }

// RefreshTokenFamily tracks a family of refresh tokens for a user. Exactly one
// token JTI per family is "current" at a time (token_jti); rotating on every
// refresh lets us detect replay of a stolen token and revoke the whole family.
type RefreshTokenFamily struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	FamilyID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"family_id"`
	TokenJTI  string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	RotatedAt *time.Time `json:"rotated_at,omitempty"`
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

func (RefreshTokenFamily) TableName() string { return "refresh_token_families" }
