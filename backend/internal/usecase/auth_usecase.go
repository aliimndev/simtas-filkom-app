package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
)

var (
	ErrInvalidCredentials  = errors.New("email atau password salah")
	ErrAccountLocked       = errors.New("akun terkunci")
	ErrUserNotFound        = errors.New("user tidak ditemukan")
	ErrInvalidResetToken   = errors.New("token reset tidak valid atau sudah kadaluarsa")
	ErrPasswordMismatch    = errors.New("password tidak cocok")
	ErrPasswordTooShort    = errors.New("password minimal 8 karakter")
	ErrPasswordNotComplex  = errors.New("password harus mengandung minimal 1 huruf kapital dan 1 angka")
	ErrTokenBlacklisted    = errors.New("token sudah tidak valid")
	ErrRefreshTokenInvalid = errors.New("refresh token tidak valid")
)

const (
	MaxLoginAttempts = 5
	LockDuration     = 30 * time.Minute
	PasswordMinLen   = 8
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token,omitempty"`
	ExpiresIn    int     `json:"expires_in"`
	User         UserDTO `json:"user"`
}

type UserDTO struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	FullName           string    `json:"full_name"`
	Role               string    `json:"role"`
	MustChangePassword bool      `json:"must_change_password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type AuthUseCase struct {
	authRepo   domainRepo.AuthRepository
	jwtManager *jwt.JWTManager
	auditSvc   *audit.AuditService
}

func NewAuthUseCase(authRepo domainRepo.AuthRepository, jwtManager *jwt.JWTManager, auditSvc *audit.AuditService) *AuthUseCase {
	return &AuthUseCase{
		authRepo:   authRepo,
		jwtManager: jwtManager,
		auditSvc:   auditSvc,
	}
}

// Login authenticates user and returns tokens
func (uc *AuthUseCase) Login(ctx context.Context, req LoginRequest, actor Actor) (*LoginResponse, error) {
	user, err := uc.authRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check account lock
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		remaining := time.Until(*user.LockedUntil).Round(time.Minute)
		// Audit: blocked login attempt on a locked account (Job 13)
		uc.logAuthAudit(ctx, user.ID, audit.ActionUserLoginFailed, actor)
		return nil, fmt.Errorf("%w. Coba lagi dalam %v", ErrAccountLocked, remaining)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		newCount := user.LoginAttemptCount + 1
		var lockedUntil *time.Time

		if newCount >= MaxLoginAttempts {
			t := time.Now().Add(LockDuration)
			lockedUntil = &t
		}

		_ = uc.authRepo.UpdateLoginAttempt(ctx, user.ID, newCount, lockedUntil)

		// Audit: failed login attempt (Job 13). Note: attempts against
		// unknown emails are intentionally NOT logged (anti-enumeration).
		uc.logAuthAudit(ctx, user.ID, audit.ActionUserLoginFailed, actor)

		if lockedUntil != nil {
			return nil, fmt.Errorf("%w. Coba lagi dalam %v menit", ErrAccountLocked, LockDuration.Minutes())
		}
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("akun tidak aktif")
	}

	// Reset login attempts on successful login
	_ = uc.authRepo.UpdateLoginAttempt(ctx, user.ID, 0, nil)
	_ = uc.authRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	accessToken, _, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Role.Name, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Audit: successful login (Job 13), recorded only after tokens are minted
	// so a USER_LOGIN is never written for a login that did not complete.
	uc.logAuthAudit(ctx, user.ID, audit.ActionUserLogin, actor)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(uc.jwtManager.AccessTokenExpiry().Seconds()),
		User: UserDTO{
			ID:                 user.ID,
			Email:              user.Email,
			FullName:           user.FullName,
			Role:               user.Role.Name,
			MustChangePassword: user.MustChangePassword,
		},
	}, nil
}

// Logout blacklists the access token JTI
func (uc *AuthUseCase) Logout(ctx context.Context, accessToken string, actor Actor) error {
	claims, err := uc.jwtManager.ValidateToken(accessToken)
	if err != nil {
		return err
	}

	blacklisted, err := uc.authRepo.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return err
	}
	// Idempotent logout: clients routinely retry with an already-blacklisted
	// token, so we skip the USER_LOGOUT audit on that path to avoid noise.
	if blacklisted {
		return nil
	}

	expiresAt := claims.ExpiresAt.Time
	if err := uc.authRepo.BlacklistToken(ctx, claims.JTI, expiresAt); err != nil {
		return err
	}

	// Audit: logout (Job 13)
	if userID, err := uuid.Parse(claims.UserID); err == nil {
		uc.logAuthAudit(ctx, userID, audit.ActionUserLogout, actor)
	}
	return nil
}

// logAuthAudit writes an auth-related audit entry (login, failed login, or
// logout) scoped to a user. Nil-safe: the audit service may be absent in tests.
func (uc *AuthUseCase) logAuthAudit(ctx context.Context, userID uuid.UUID, action string, actor Actor) {
	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &userID,
		Action:     action,
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
}

// RefreshToken validates refresh token and issues new access token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	claims, err := uc.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}

	blacklisted, err := uc.authRepo.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, ErrTokenBlacklisted
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}

	user, err := uc.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, errors.New("akun tidak aktif")
	}

	accessToken, _, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Role.Name, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(uc.jwtManager.AccessTokenExpiry().Seconds()),
	}, nil
}

// GetMe returns current user profile
func (uc *AuthUseCase) GetMe(ctx context.Context, userID uuid.UUID) (*UserDTO, error) {
	user, err := uc.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &UserDTO{
		ID:                 user.ID,
		Email:              user.Email,
		FullName:           user.FullName,
		Role:               user.Role.Name,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// ForgotPassword generates and stores a reset token; returns token for email sending
func (uc *AuthUseCase) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := uc.authRepo.FindUserByEmail(ctx, email)
	if err != nil {
		// Prevent email enumeration — silently succeed
		return "", nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	resetToken := hex.EncodeToString(tokenBytes)

	prt := &entity.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := uc.authRepo.CreatePasswordResetToken(ctx, prt); err != nil {
		return "", err
	}

	return resetToken, nil
}

// ResetPassword validates token, updates password, and marks token used
func (uc *AuthUseCase) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return ErrPasswordMismatch
	}
	if len(req.NewPassword) < PasswordMinLen {
		return ErrPasswordTooShort
	}
	if !isPasswordComplex(req.NewPassword) {
		return ErrPasswordNotComplex
	}

	prt, err := uc.authRepo.FindPasswordResetTokenByToken(ctx, req.Token)
	if err != nil {
		return ErrInvalidResetToken
	}
	if prt.UsedAt != nil {
		return ErrInvalidResetToken
	}
	if prt.ExpiresAt.Before(time.Now()) {
		return ErrInvalidResetToken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return err
	}

	if err := uc.authRepo.UpdatePassword(ctx, prt.UserID, string(hashedPassword)); err != nil {
		return err
	}
	if err := uc.authRepo.MarkPasswordResetTokenUsed(ctx, prt.ID); err != nil {
		return err
	}
	_ = uc.authRepo.ClearMustChangePassword(ctx, prt.UserID)

	return nil
}

// isPasswordComplex checks password for at least one uppercase letter and one digit
func isPasswordComplex(password string) bool {
	hasUpper := false
	hasDigit := false
	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
		if hasUpper && hasDigit {
			return true
		}
	}
	return false
}
