package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
)

func TestIsPasswordComplex(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"uppercase and digit", "Password1", true},
		{"uppercase only", "PASSWORD", false},
		{"digit only", "password1", false},
		{"neither", "password", false},
		{"digit at start uppercase at end", "1passwordA", true},
		{"unicode uppercase treated as upper", "Pässwörd1", true},
		{"symbols no digit", "P@ssword!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPasswordComplex(tt.password); got != tt.want {
				t.Errorf("isPasswordComplex(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

type fakeAuthRepo struct{}

func (f *fakeAuthRepo) FindUserByEmail(_ context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (f *fakeAuthRepo) FindUserByID(_ context.Context, _ uuid.UUID) (*entity.User, error) {
	return nil, nil
}
func (f *fakeAuthRepo) IncrementLoginAttempt(_ context.Context, _ uuid.UUID, _ int, _ time.Duration) (int, bool, error) {
	return 0, false, nil
}
func (f *fakeAuthRepo) ResetLoginAttempts(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeAuthRepo) UpdateLastLogin(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeAuthRepo) BlacklistToken(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (f *fakeAuthRepo) IsTokenBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (f *fakeAuthRepo) CreatePasswordResetToken(_ context.Context, _ *entity.PasswordResetToken) error {
	return nil
}
func (f *fakeAuthRepo) FindPasswordResetTokenByToken(_ context.Context, _ string) (*entity.PasswordResetToken, error) {
	return nil, nil
}
func (f *fakeAuthRepo) MarkPasswordResetTokenUsed(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (f *fakeAuthRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (f *fakeAuthRepo) ClearMustChangePassword(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (f *fakeAuthRepo) GetUserTokenVersion(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (f *fakeAuthRepo) CreateRefreshTokenFamily(_ context.Context, _ *entity.RefreshTokenFamily) error {
	return nil
}
func (f *fakeAuthRepo) FindRefreshTokenFamilyByJTI(_ context.Context, _ string) (*entity.RefreshTokenFamily, error) {
	return nil, nil
}
func (f *fakeAuthRepo) RotateRefreshTokenFamily(_ context.Context, _, _ string, _ time.Time) (bool, error) {
	return false, nil
}
func (f *fakeAuthRepo) RevokeRefreshTokenFamiliesByUser(_ context.Context, _ uuid.UUID) error {
	return nil
}

// chanAuditRepo records the actions written by AuditService.Log (which runs in
// a goroutine) so tests can assert the auth audit trail deterministically.
type chanAuditRepo struct {
	actions chan string
}

func (c *chanAuditRepo) Create(_ context.Context, l *entity.AuditLog) error {
	c.actions <- l.Action
	return nil
}
func (c *chanAuditRepo) FindAll(_ context.Context, _ domainRepo.AuditFilter) ([]*entity.AuditLog, int64, error) {
	return nil, 0, nil
}
func (c *chanAuditRepo) FindByEntity(_ context.Context, _ string, _ uuid.UUID) ([]*entity.AuditLog, error) {
	return nil, nil
}

// authUserRepo returns a real user for login/audit tests.
type authUserRepo struct {
	user *entity.User
}

func (f *authUserRepo) FindUserByEmail(_ context.Context, email string) (*entity.User, error) {
	if f.user != nil && f.user.Email == email {
		return f.user, nil
	}
	return nil, errors.New("not found")
}
func (f *authUserRepo) FindUserByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	if f.user != nil && f.user.ID == id {
		return f.user, nil
	}
	return nil, errors.New("not found")
}
func (f *authUserRepo) IncrementLoginAttempt(_ context.Context, _ uuid.UUID, _ int, _ time.Duration) (int, bool, error) {
	return 1, false, nil
}
func (f *authUserRepo) ResetLoginAttempts(_ context.Context, _ uuid.UUID) error { return nil }
func (f *authUserRepo) UpdateLastLogin(_ context.Context, _ uuid.UUID) error { return nil }
func (f *authUserRepo) BlacklistToken(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (f *authUserRepo) IsTokenBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (f *authUserRepo) CreatePasswordResetToken(_ context.Context, _ *entity.PasswordResetToken) error {
	return nil
}
func (f *authUserRepo) FindPasswordResetTokenByToken(_ context.Context, _ string) (*entity.PasswordResetToken, error) {
	return nil, nil
}
func (f *authUserRepo) MarkPasswordResetTokenUsed(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (f *authUserRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (f *authUserRepo) ClearMustChangePassword(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (f *authUserRepo) GetUserTokenVersion(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (f *authUserRepo) CreateRefreshTokenFamily(_ context.Context, _ *entity.RefreshTokenFamily) error {
	return nil
}
func (f *authUserRepo) FindRefreshTokenFamilyByJTI(_ context.Context, _ string) (*entity.RefreshTokenFamily, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *authUserRepo) RotateRefreshTokenFamily(_ context.Context, _, _ string, _ time.Time) (bool, error) {
	return false, nil
}
func (f *authUserRepo) RevokeRefreshTokenFamiliesByUser(_ context.Context, _ uuid.UUID) error {
	return nil
}

func newTestUser(password string) *entity.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return &entity.User{
		ID:           uuid.New(),
		Email:        "budi@example.com",
		PasswordHash: string(hash),
		FullName:     "Budi Santoso",
		IsActive:     true,
		Role:         entity.Role{Name: "mahasiswa"},
	}
}

func waitForAudit(t *testing.T, ch chan string, want string) {
	t.Helper()
	select {
	case got := <-ch:
		if got != want {
			t.Errorf("audit action = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for audit action %q", want)
	}
}

func TestLoginSuccessAudits(t *testing.T) {
	user := newTestUser("Password123")
	repo := &authUserRepo{user: user}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	resp, err := uc.Login(context.Background(), LoginRequest{Email: user.Email, Password: "Password123"}, Actor{})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected an access token")
	}
	waitForAudit(t, auditRepo.actions, audit.ActionUserLogin)
}

// TestLoginUnknownEmail — an unknown email must return the same generic
// credentials error (after a dummy bcrypt compare) so response timing does not
// leak which emails are registered.
func TestLoginUnknownEmail(t *testing.T) {
	user := newTestUser("Password123")
	repo := &authUserRepo{user: user}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	_, err := uc.Login(context.Background(), LoginRequest{Email: "nobody@example.com", Password: "Password123"}, Actor{})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginFailedAudit(t *testing.T) {
	user := newTestUser("Password123")
	repo := &authUserRepo{user: user}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	_, err := uc.Login(context.Background(), LoginRequest{Email: user.Email, Password: "WrongPass1"}, Actor{})
	if !errors.Is(err, ErrInvalidCredentials) && !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("Login error = %v", err)
	}
	waitForAudit(t, auditRepo.actions, audit.ActionUserLoginFailed)
}

func TestLogoutAudit(t *testing.T) {
	user := newTestUser("Password123")
	repo := &authUserRepo{user: user}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	token, _, err := jwtMgr.GenerateAccessToken(user.ID, user.Role.Name, user.Email, 0)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if err := uc.Logout(context.Background(), token, Actor{}); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	waitForAudit(t, auditRepo.actions, audit.ActionUserLogout)
}

func TestResetPasswordValidation(t *testing.T) {
	uc := NewAuthUseCase(&fakeAuthRepo{}, nil, nil)

	tests := []struct {
		name string
		req  ResetPasswordRequest
		want error
	}{
		{"password mismatch", ResetPasswordRequest{Token: "t", NewPassword: "StrongPass1", ConfirmPassword: "Different1"}, ErrPasswordMismatch},
		{"too short", ResetPasswordRequest{Token: "t", NewPassword: "Short1", ConfirmPassword: "Short1"}, ErrPasswordTooShort},
		{"missing uppercase", ResetPasswordRequest{Token: "t", NewPassword: "weakpass1", ConfirmPassword: "weakpass1"}, ErrPasswordNotComplex},
		{"missing digit", ResetPasswordRequest{Token: "t", NewPassword: "WeakPassw", ConfirmPassword: "WeakPassw"}, ErrPasswordNotComplex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.ResetPassword(context.Background(), tt.req)
			if err != tt.want {
				t.Errorf("ResetPassword error = %v, want %v", err, tt.want)
			}
		})
	}
}
