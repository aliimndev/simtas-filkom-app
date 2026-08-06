package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) domainRepo.AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) FindUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) UpdateLoginAttempt(ctx context.Context, userID uuid.UUID, count int, lockedUntil *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"login_attempt_count": count,
			"locked_until":        lockedUntil,
		}).Error
}

func (r *authRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

func (r *authRepository) BlacklistToken(ctx context.Context, jti string, expiresAt time.Time) error {
	blacklist := &entity.TokenBlacklist{
		TokenJTI:  jti,
		ExpiresAt: expiresAt,
	}
	return r.db.WithContext(ctx).Create(blacklist).Error
}

func (r *authRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.TokenBlacklist{}).
		Where("token_jti = ? AND expires_at > ?", jti, time.Now()).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *authRepository) CreatePasswordResetToken(ctx context.Context, token *entity.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepository) FindPasswordResetTokenByToken(ctx context.Context, token string) (*entity.PasswordResetToken, error) {
	var prt entity.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&prt).Error
	if err != nil {
		return nil, err
	}
	return &prt, nil
}

func (r *authRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.PasswordResetToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now).Error
}

func (r *authRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

func (r *authRepository) ClearMustChangePassword(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("must_change_password", false).Error
}

func (r *authRepository) GetUserTokenVersion(ctx context.Context, userID string) (int, error) {
	// First() returns ErrRecordNotFound for missing AND soft-deleted users
	// (gorm auto-appends deleted_at IS NULL), so deleted users' tokens die.
	var user entity.User
	err := r.db.WithContext(ctx).
		Select("token_version").
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.TokenVersion, nil
}

// ── Refresh token rotation ──────────────────────────────────────────────

func (r *authRepository) CreateRefreshTokenFamily(ctx context.Context, family *entity.RefreshTokenFamily) error {
	return r.db.WithContext(ctx).Create(family).Error
}

func (r *authRepository) FindRefreshTokenFamilyByJTI(ctx context.Context, jti string) (*entity.RefreshTokenFamily, error) {
	var family entity.RefreshTokenFamily
	err := r.db.WithContext(ctx).
		Where("token_jti = ?", jti).
		First(&family).Error
	if err != nil {
		return nil, err
	}
	return &family, nil
}

func (r *authRepository) RotateRefreshTokenFamily(ctx context.Context, oldJTI, newJTI string, newExpiresAt time.Time) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&entity.RefreshTokenFamily{}).
		Where("token_jti = ?", oldJTI).
		Updates(map[string]interface{}{
			"token_jti":  newJTI,
			"expires_at": newExpiresAt,
			"rotated_at": time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *authRepository) RevokeRefreshTokenFamiliesByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.RefreshTokenFamily{}).Error
}
