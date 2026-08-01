package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/api/internal/domain/entity"
)

type AuthRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)

	UpdateLoginAttempt(ctx context.Context, userID uuid.UUID, count int, lockedUntil *time.Time) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error

	BlacklistToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)

	CreatePasswordResetToken(ctx context.Context, token *entity.PasswordResetToken) error
	FindPasswordResetTokenByToken(ctx context.Context, token string) (*entity.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error

	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	ClearMustChangePassword(ctx context.Context, userID uuid.UUID) error

	// GetUserTokenVersion returns the current token_version for session validation.
	GetUserTokenVersion(ctx context.Context, userID string) (int, error)
}
