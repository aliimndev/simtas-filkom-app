package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

type AuthRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)

	// IncrementLoginAttempt atomically bumps the failed-attempt counter (and
	// sets locked_until once it reaches maxAttempts) in a single UPDATE, so
	// concurrent wrong-password requests cannot each read a stale count and
	// bypass the account lockout. Returns the new count and whether the
	// account became locked.
	IncrementLoginAttempt(ctx context.Context, userID uuid.UUID, maxAttempts int, lockDuration time.Duration) (count int, locked bool, err error)
	// ResetLoginAttempts clears the failed-attempt counter on successful login.
	ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error
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

	// ── Refresh token rotation ──────────────────────────────────────────
	// CreateRefreshTokenFamily stores a new refresh-token family for a user.
	CreateRefreshTokenFamily(ctx context.Context, family *entity.RefreshTokenFamily) error
	// FindRefreshTokenFamilyByJTI returns the family row whose current token JTI
	// matches jti, or gorm.ErrRecordNotFound when the token was already rotated.
	FindRefreshTokenFamilyByJTI(ctx context.Context, jti string) (*entity.RefreshTokenFamily, error)
	// RotateRefreshTokenFamily atomically swaps the family's current JTI from
	// oldJTI to newJTI. It returns false (and no error) when oldJTI no longer
	// matches — i.e. the token was already rotated by a concurrent request.
	RotateRefreshTokenFamily(ctx context.Context, oldJTI, newJTI string, newExpiresAt time.Time) (bool, error)
	// RevokeRefreshTokenFamiliesByUser deletes every refresh-token family of a
	// user (used on token-reuse detection and logout).
	RevokeRefreshTokenFamiliesByUser(ctx context.Context, userID uuid.UUID) error
}
