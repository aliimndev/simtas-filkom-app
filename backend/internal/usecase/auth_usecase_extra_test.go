package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
)

// resetAwareAuthRepo extends authUserRepo with configurable reset-token lookup
// so the ResetPassword success path can be exercised.
type resetAwareAuthRepo struct {
	*authUserRepo
	resetToken *entity.PasswordResetToken
}

func (f *resetAwareAuthRepo) FindPasswordResetTokenByToken(_ context.Context, _ string) (*entity.PasswordResetToken, error) {
	if f.resetToken == nil {
		return nil, errors.New("token not found")
	}
	return f.resetToken, nil
}

// blacklistAwareAuthRepo reports blacklist state for refresh-token tests.
type blacklistAwareAuthRepo struct {
	*authUserRepo
	blacklisted bool
}

func (f *blacklistAwareAuthRepo) IsTokenBlacklisted(_ context.Context, _ string) (bool, error) {
	return f.blacklisted, nil
}

func newAuthUC(user *entity.User) (*AuthUseCase, *jwt.JWTManager) {
	repo := &authUserRepo{user: user}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))
	return uc, jwtMgr
}

func TestRefreshTokenSuccess(t *testing.T) {
	user := newTestUser("Password123")
	uc, jwtMgr, repo := newRotationAuthUC(user)

	refresh, jti, err := jwtMgr.GenerateRefreshToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	// Seed the family row the way Login would.
	if err := repo.CreateRefreshTokenFamily(context.Background(), &entity.RefreshTokenFamily{
		UserID:    user.ID,
		FamilyID:  uuid.New(),
		TokenJTI:  jti,
		ExpiresAt: time.Now().Add(jwtMgr.RefreshTokenExpiry()),
	}); err != nil {
		t.Fatalf("seed family: %v", err)
	}

	resp, err := uc.RefreshToken(context.Background(), refresh)
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected a new access token")
	}
	if resp.RefreshToken == "" {
		t.Error("expected a rotated refresh token")
	}
	if resp.ExpiresIn != 3600 {
		t.Errorf("expires_in = %d, want 3600", resp.ExpiresIn)
	}
}

func TestRefreshTokenInvalid(t *testing.T) {
	user := newTestUser("Password123")
	uc, _ := newAuthUC(user)

	_, err := uc.RefreshToken(context.Background(), "not-a-jwt")
	if !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Errorf("expected ErrRefreshTokenInvalid, got %v", err)
	}
}

// blacklistAwareRotationRepo combines family tracking with a forced blacklist
// so the blacklisted-token path can be tested after the family check.
type blacklistAwareRotationRepo struct {
	*rotationAuthRepo
	blacklisted bool
}

func (f *blacklistAwareRotationRepo) IsTokenBlacklisted(_ context.Context, _ string) (bool, error) {
	return f.blacklisted, nil
}

func TestRefreshTokenBlacklisted(t *testing.T) {
	user := newTestUser("Password123")
	repo := &blacklistAwareRotationRepo{
		rotationAuthRepo: newRotationAuthRepo(user),
		blacklisted:      true,
	}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	refresh, jti, _ := jwtMgr.GenerateRefreshToken(user.ID)
	if err := repo.CreateRefreshTokenFamily(context.Background(), &entity.RefreshTokenFamily{
		UserID:    user.ID,
		FamilyID:  uuid.New(),
		TokenJTI:  jti,
		ExpiresAt: time.Now().Add(jwtMgr.RefreshTokenExpiry()),
	}); err != nil {
		t.Fatalf("seed family: %v", err)
	}
	_, err := uc.RefreshToken(context.Background(), refresh)
	if !errors.Is(err, ErrTokenBlacklisted) {
		t.Errorf("expected ErrTokenBlacklisted, got %v", err)
	}
}

func TestRefreshTokenInactiveUser(t *testing.T) {
	user := newTestUser("Password123")
	user.IsActive = false
	uc, jwtMgr, repo := newRotationAuthUC(user)

	refresh, jti, _ := jwtMgr.GenerateRefreshToken(user.ID)
	if err := repo.CreateRefreshTokenFamily(context.Background(), &entity.RefreshTokenFamily{
		UserID:    user.ID,
		FamilyID:  uuid.New(),
		TokenJTI:  jti,
		ExpiresAt: time.Now().Add(jwtMgr.RefreshTokenExpiry()),
	}); err != nil {
		t.Fatalf("seed family: %v", err)
	}
	_, err := uc.RefreshToken(context.Background(), refresh)
	if err == nil {
		t.Fatal("expected error for inactive user")
	}
}

func TestRefreshTokenUserNotFound(t *testing.T) {
	user := newTestUser("Password123")
	uc, _ := newAuthUC(user)

	// Refresh token for a user that is not the one stored in the repo.
	other := uuid.New()
	mgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	refresh, _, _ := mgr.GenerateRefreshToken(other)
	_, err := uc.RefreshToken(context.Background(), refresh)
	if !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Errorf("expected ErrRefreshTokenInvalid (no family row for unknown user), got %v", err)
	}
}

func TestGetMe(t *testing.T) {
	user := newTestUser("Password123")
	user.MustChangePassword = true
	uc, _ := newAuthUC(user)

	dto, err := uc.GetMe(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if dto.Email != user.Email {
		t.Errorf("email = %q, want %q", dto.Email, user.Email)
	}
	if dto.Role != "mahasiswa" {
		t.Errorf("role = %q, want mahasiswa", dto.Role)
	}
	if !dto.MustChangePassword {
		t.Error("must_change_password should be true for a fresh user")
	}
}

func TestGetMeNotFound(t *testing.T) {
	user := newTestUser("Password123")
	uc, _ := newAuthUC(user)

	_, err := uc.GetMe(context.Background(), uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestForgotPassword(t *testing.T) {
	user := newTestUser("Password123")
	uc, _ := newAuthUC(user)

	token, err := uc.ForgotPassword(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("ForgotPassword: %v", err)
	}
	if len(token) != 64 { // 32 random bytes hex-encoded
		t.Errorf("token length = %d, want 64", len(token))
	}
}

// TestForgotPasswordUnknownEmail — anti-enumeration: silently succeeds.
func TestForgotPasswordUnknownEmail(t *testing.T) {
	user := newTestUser("Password123")
	uc, _ := newAuthUC(user)

	token, err := uc.ForgotPassword(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatalf("ForgotPassword: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token for unknown email, got %q", token)
	}
}

func TestResetPasswordSuccess(t *testing.T) {
	user := newTestUser("Password123")
	repo := &resetAwareAuthRepo{
		authUserRepo: &authUserRepo{user: user},
		resetToken: &entity.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			Token:     "valid-token",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	err := uc.ResetPassword(context.Background(), ResetPasswordRequest{
		Token:           "valid-token",
		NewPassword:     "NewPassword1",
		ConfirmPassword: "NewPassword1",
	})
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
}

func TestResetPasswordInvalidToken(t *testing.T) {
	user := newTestUser("Password123")
	// resetAwareAuthRepo with a nil resetToken returns a lookup error (like a
	// real repository does for a missing row) instead of (nil, nil).
	repo := &resetAwareAuthRepo{authUserRepo: &authUserRepo{user: user}}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	err := uc.ResetPassword(context.Background(), ResetPasswordRequest{
		Token:           "invalid-token",
		NewPassword:     "NewPassword1",
		ConfirmPassword: "NewPassword1",
	})
	if !errors.Is(err, ErrInvalidResetToken) {
		t.Errorf("expected ErrInvalidResetToken, got %v", err)
	}
}

func TestResetPasswordExpiredToken(t *testing.T) {
	user := newTestUser("Password123")
	repo := &resetAwareAuthRepo{
		authUserRepo: &authUserRepo{user: user},
		resetToken: &entity.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			Token:     "expired-token",
			ExpiresAt: time.Now().Add(-time.Hour),
		},
	}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	err := uc.ResetPassword(context.Background(), ResetPasswordRequest{
		Token:           "expired-token",
		NewPassword:     "NewPassword1",
		ConfirmPassword: "NewPassword1",
	})
	if !errors.Is(err, ErrInvalidResetToken) {
		t.Errorf("expected ErrInvalidResetToken for expired token, got %v", err)
	}
}

// TestLoginLockedAfterFiveFailures — Job 23: 5x gagal → akun terkunci.
type lockTrackingAuthRepo struct {
	*authUserRepo
	attemptCount int
	lockedUntil  *time.Time
}

// UpdateLoginAttempt mirrors persistence: keep the in-memory user's counter in
// sync so the next Login() call reads the updated attempt count (like a real DB
// row would after UPDATE).
func (f *lockTrackingAuthRepo) UpdateLoginAttempt(_ context.Context, _ uuid.UUID, count int, locked *time.Time) error {
	f.attemptCount = count
	f.lockedUntil = locked
	if f.user != nil {
		f.user.LoginAttemptCount = count
		f.user.LockedUntil = locked
	}
	return nil
}

func TestLoginLockedAfterFiveFailures(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.MinCost)
	user := &entity.User{
		ID:           uuid.New(),
		Email:        "budi@example.com",
		PasswordHash: string(hash),
		FullName:     "Budi Santoso",
		IsActive:     true,
		Role:         entity.Role{Name: "mahasiswa"},
	}
	repo := &lockTrackingAuthRepo{authUserRepo: &authUserRepo{user: user}}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 8)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	var lockedErr error
	for i := 0; i < MaxLoginAttempts; i++ {
		_, err := uc.Login(context.Background(), LoginRequest{Email: user.Email, Password: "WrongPass1"}, Actor{})
		if errors.Is(err, ErrAccountLocked) {
			lockedErr = err
			break
		}
		if i == MaxLoginAttempts-1 && err == nil {
			t.Fatal("expected account to lock after max attempts")
		}
	}
	if lockedErr == nil {
		t.Fatal("expected ErrAccountLocked after 5 failed attempts")
	}
	if repo.lockedUntil == nil {
		t.Error("expected LockedUntil to be set")
	}
}
