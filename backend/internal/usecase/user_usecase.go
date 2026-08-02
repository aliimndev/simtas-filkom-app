package usecase

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

var (
	ErrEmailAlreadyExists   = errors.New("email sudah terdaftar")
	ErrRoleInvalid          = errors.New("role tidak valid")
	ErrCannotDeleteSelf     = errors.New("tidak dapat menghapus akun sendiri")
	ErrCannotDeactivateSelf = errors.New("tidak dapat menonaktifkan akun sendiri")
	ErrFullNameRequired     = errors.New("email dan nama lengkap wajib diisi")
	ErrInvalidEmailFormat   = errors.New("format email tidak valid")
)

// CreateUserRequest is the payload for POST /admin/users
type CreateUserRequest struct {
	Email        string `json:"email" binding:"required,email"`
	FullName     string `json:"full_name" binding:"required"`
	NimNidn      string `json:"nim_nidn"`
	Role         string `json:"role" binding:"required"`
	StudyProgram string `json:"study_program"`
}

// UpdateUserRequest is the payload for PUT /admin/users/:id.
// Email and role cannot be changed via this endpoint.
type UpdateUserRequest struct {
	FullName        *string `json:"full_name"`
	NimNidn         *string `json:"nim_nidn"`
	StudyProgram    *string `json:"study_program"`
	ProfilePhotoURL *string `json:"profile_photo_url"`
}

// UserUseCase contains business logic for admin user management.
type UserUseCase struct {
	userRepo domainRepo.UserRepository
	emailSvc email.EmailService
	auditSvc *audit.AuditService
}

func NewUserUseCase(
	userRepo domainRepo.UserRepository,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		emailSvc: emailSvc,
		auditSvc: auditSvc,
	}
}

// List returns paginated users matching the filter.
func (uc *UserUseCase) List(ctx context.Context, filter domainRepo.UserFilter) ([]*entity.User, int64, error) {
	return uc.userRepo.FindAll(ctx, filter)
}

// GetByID returns a single user (404 via ErrUserNotFound if missing/deleted).
func (uc *UserUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// Create creates a new user with an auto-generated temporary password,
// sends a welcome email, and records an audit log.
func (uc *UserUseCase) Create(ctx context.Context, req CreateUserRequest, actor Actor) (*entity.User, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Email == "" || req.FullName == "" {
		return nil, ErrFullNameRequired
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, ErrInvalidEmailFormat
	}

	role, err := uc.userRepo.FindRoleByName(ctx, req.Role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleInvalid
		}
		return nil, err
	}

	// Email uniqueness
	existing, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	tempPassword := utils.GenerateRandomPassword(12)
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), 12)
	if err != nil {
		return nil, err
	}

	var nimNidn, studyProgram *string
	if req.NimNidn != "" {
		nimNidn = &req.NimNidn
	}
	if req.StudyProgram != "" {
		studyProgram = &req.StudyProgram
	}

	user := &entity.User{
		Email:              req.Email,
		PasswordHash:       string(hash),
		FullName:           req.FullName,
		NimNidn:            nimNidn,
		RoleID:             role.ID,
		Role:               *role,
		StudyProgram:       studyProgram,
		IsActive:           true,
		MustChangePassword: true,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Welcome email (async, non-fatal)
	go func() {
		_ = uc.emailSvc.SendWelcomeEmail(context.Background(), user.Email, user.FullName, tempPassword)
	}()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionUserCreated,
		EntityType: "user",
		EntityID:   &user.ID,
		NewValue:   map[string]interface{}{"email": user.Email, "full_name": user.FullName, "role": req.Role},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	return user, nil
}

// Update updates editable fields of a user and records old/new audit values.
func (uc *UserUseCase) Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest, actor Actor) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	oldValue := userSnapshot(user)

	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.NimNidn != nil {
		user.NimNidn = req.NimNidn
	}
	if req.StudyProgram != nil {
		user.StudyProgram = req.StudyProgram
	}
	if req.ProfilePhotoURL != nil {
		user.ProfilePhotoURL = req.ProfilePhotoURL
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionUserUpdated,
		EntityType: "user",
		EntityID:   &user.ID,
		OldValue:   oldValue,
		NewValue:   userSnapshot(user),
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	return user, nil
}

// Delete soft-deletes a user. An admin cannot delete their own account.
func (uc *UserUseCase) Delete(ctx context.Context, id uuid.UUID, actor Actor) error {
	if id == actor.UserID {
		return ErrCannotDeleteSelf
	}
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := uc.userRepo.SoftDelete(ctx, id); err != nil {
		return err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionUserDeleted,
		EntityType: "user",
		EntityID:   &user.ID,
		OldValue:   map[string]interface{}{"email": user.Email},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return nil
}

// Activate sets is_active = true.
func (uc *UserUseCase) Activate(ctx context.Context, id uuid.UUID, actor Actor) error {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := uc.userRepo.SetActiveStatus(ctx, id, true); err != nil {
		return err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionUserActivated,
		EntityType: "user",
		EntityID:   &user.ID,
		NewValue:   map[string]interface{}{"is_active": true},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return nil
}

// Deactivate sets is_active = false and invalidates all active sessions.
// An admin cannot deactivate their own account.
func (uc *UserUseCase) Deactivate(ctx context.Context, id uuid.UUID, actor Actor) error {
	if id == actor.UserID {
		return ErrCannotDeactivateSelf
	}
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := uc.userRepo.SetActiveStatus(ctx, id, false); err != nil {
		return err
	}
	// Bump token_version so all previously issued tokens are rejected.
	if err := uc.userRepo.InvalidateUserSessions(ctx, id); err != nil {
		return err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionUserDeactivated,
		EntityType: "user",
		EntityID:   &user.ID,
		NewValue:   map[string]interface{}{"is_active": false},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return nil
}

// ResetPassword generates a new temporary password, stores its hash,
// marks must_change_password, emails the user, and records an audit log.
func (uc *UserUseCase) ResetPassword(ctx context.Context, id uuid.UUID, actor Actor) error {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	newPassword := utils.GenerateRandomPassword(12)
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	if err := uc.userRepo.ResetPassword(ctx, id, string(hash)); err != nil {
		return err
	}
	// Invalidate existing sessions so old tokens die after the reset
	if err := uc.userRepo.InvalidateUserSessions(ctx, id); err != nil {
		return err
	}

	go func() {
		_ = uc.emailSvc.SendPasswordReset(context.Background(), user.Email, user.FullName, newPassword)
	}()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionUserPasswordReset,
		EntityType: "user",
		EntityID:   &user.ID,
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return nil
}

// ChangePasswordRequest is the payload for PUT /users/me/password
// (user changing their own password).
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// ChangeMyPassword lets a user change their own password after verifying the
// current one. On success it clears must_change_password (Job 04/15).
//
// Note: we deliberately do NOT bump token_version here — the user is changing
// their own password with an active session, and invalidating tokens would log
// them straight back out (breaking the post-login change flow). Admin-initiated
// resets (ResetPassword) still invalidate sessions because that user is not
// logged in at that point.
func (uc *UserUseCase) ChangeMyPassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest, actor Actor) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrPasswordMismatch
	}
	if len(req.NewPassword) < PasswordMinLen {
		return ErrPasswordTooShort
	}
	if !isPasswordComplex(req.NewPassword) {
		return ErrPasswordNotComplex
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return err
	}
	if err := uc.userRepo.ChangePassword(ctx, userID, string(hash)); err != nil {
		return err
	}

	// Audit: self-initiated password change (Job 13)
	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &userID,
		Action:     audit.ActionUserPasswordReset,
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return nil
}

// userSnapshot builds a safe map of editable fields for audit diffing.
func userSnapshot(u *entity.User) map[string]interface{} {
	snap := map[string]interface{}{
		"full_name":     u.FullName,
		"study_program": derefStr(u.StudyProgram),
	}
	if u.NimNidn != nil {
		snap["nim_nidn"] = *u.NimNidn
	}
	if u.ProfilePhotoURL != nil {
		snap["profile_photo_url"] = *u.ProfilePhotoURL
	}
	return snap
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Actor describes who performed an admin action (for audit logs).
type Actor struct {
	UserID    uuid.UUID
	IPAddress *string
	UserAgent *string
}
