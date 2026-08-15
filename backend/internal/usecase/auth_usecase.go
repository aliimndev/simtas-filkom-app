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
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
)

var (
	ErrInvalidCredentials  = errors.New("email atau password salah")
	ErrAccountLocked       = errors.New("akun terkunci")
	ErrUserInactive        = errors.New("akun tidak aktif")
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

// dummyPasswordHash is compared against when the submitted email does not
// exist, so unknown-email logins take the same bcrypt time as a wrong password
// (anti user-enumeration via response timing).
var dummyPasswordHash = func() string {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing-equalization"), 12)
	if err != nil {
		panic(err)
	}
	return string(h)
}()

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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
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
		// Equalize timing: burn a bcrypt compare so unknown emails respond at
		// the same latency as a wrong password (anti user-enumeration).
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(req.Password))
		return nil, ErrInvalidCredentials
	}

	// Check account lock
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		remaining := time.Until(*user.LockedUntil).Round(time.Minute)
		// Audit: blocked login attempt on a locked account (Job 13)
		uc.logAuthAudit(ctx, user.ID, audit.ActionUserLoginFailed, actor)
		return nil, fmt.Errorf("%w. Coba lagi dalam %v", ErrAccountLocked, remaining)
	}

	// Verify password. Failed attempts are counted atomically in the database
	// (login_attempt_count = login_attempt_count + 1), so concurrent wrong
	// passwords cannot each read the same stale count and bypass the lockout.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		_, locked, _ := uc.authRepo.IncrementLoginAttempt(ctx, user.ID, MaxLoginAttempts, LockDuration)

		// Audit: failed login attempt (Job 13). Note: attempts against
		// unknown emails are intentionally NOT logged (anti-enumeration).
		uc.logAuthAudit(ctx, user.ID, audit.ActionUserLoginFailed, actor)

		if locked {
			return nil, fmt.Errorf("%w. Coba lagi dalam %v menit", ErrAccountLocked, LockDuration.Minutes())
		}
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Reset login attempts on successful login
	_ = uc.authRepo.ResetLoginAttempts(ctx, user.ID)
	_ = uc.authRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	accessToken, _, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Role.Name, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshJTI, err := uc.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Store the new refresh-token family (rotation baseline). A failure here
	// must fail the login: without a family row the token could never be
	// refreshed, stranding the user.
	if err := uc.authRepo.CreateRefreshTokenFamily(ctx, &entity.RefreshTokenFamily{
		UserID:    user.ID,
		FamilyID:  uuid.New(),
		TokenJTI:  refreshJTI,
		ExpiresAt: time.Now().Add(uc.jwtManager.RefreshTokenExpiry()),
	}); err != nil {
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
		// Revoke all refresh-token families so a stolen refresh token can no
		// longer mint new access tokens after logout.
		_ = uc.authRepo.RevokeRefreshTokenFamiliesByUser(ctx, userID)
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

// RefreshToken validates the refresh token, rotates the refresh-token family,
// and issues a new access token. Rotation bounds the token-theft window: the
// presented JTI becomes the *old* one, and replaying it (or any family
// member) revokes the entire family.
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	claims, err := uc.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}

	// The presented token must be the family's *current* JTI. If it is not,
	// it was already rotated (or never issued) → treat as token reuse/theft
	// and revoke the whole family.
	family, err := uc.authRepo.FindRefreshTokenFamilyByJTI(ctx, claims.JTI)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Reuse detection: kill every refresh session for this user.
			_ = uc.authRepo.RevokeRefreshTokenFamiliesByUser(ctx, userID)
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}
	if family.UserID != userID {
		_ = uc.authRepo.RevokeRefreshTokenFamiliesByUser(ctx, userID)
		return nil, ErrRefreshTokenInvalid
	}

	blacklisted, err := uc.authRepo.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, ErrTokenBlacklisted
	}

	user, err := uc.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	accessToken, _, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Role.Name, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	// Rotate: mint a new refresh token and atomically swap the family's
	// current JTI. RotateRefreshTokenFamily is a compare-and-swap: it returns
	// false when another request already rotated the same token, which means
	// the presented token was replayed concurrently — revoke the family.
	newRefreshToken, newJTI, err := uc.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}
	rotated, err := uc.authRepo.RotateRefreshTokenFamily(ctx, claims.JTI, newJTI, time.Now().Add(uc.jwtManager.RefreshTokenExpiry()))
	if err != nil {
		return nil, err
	}
	if !rotated {
		_ = uc.authRepo.RevokeRefreshTokenFamiliesByUser(ctx, userID)
		return nil, ErrRefreshTokenInvalid
	}

	// The old JTI must not be accepted again; blacklist it until it naturally
	// expires (non-fatal — the family row is already the source of truth).
	if exp := claims.ExpiresAt.Time; !exp.IsZero() {
		_ = uc.authRepo.BlacklistToken(ctx, claims.JTI, exp)
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(uc.jwtManager.AccessTokenExpiry().Seconds()),
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

// ForgotPassword generates and stores a reset token; returns the token and the
// user's full name for email sending.
func (uc *AuthUseCase) ForgotPassword(ctx context.Context, email string) (string, string, error) {
	user, err := uc.authRepo.FindUserByEmail(ctx, email)
	if err != nil {
		// Prevent email enumeration — silently succeed
		return "", "", nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	resetToken := hex.EncodeToString(tokenBytes)

	prt := &entity.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := uc.authRepo.CreatePasswordResetToken(ctx, prt); err != nil {
		return "", "", err
	}

	return resetToken, user.FullName, nil
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
