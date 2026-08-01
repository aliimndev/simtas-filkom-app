package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// UserFilter carries optional filters for listing users.
type UserFilter struct {
	Role         string // filter by role name
	IsActive     *bool  // filter by active status
	StudyProgram string // filter by study program
	Search       string // matches full_name, email, or nim_nidn
	Page         int
	PerPage      int
}

// UserRepository defines persistence operations for user management (Job 04).
type UserRepository interface {
	FindAll(ctx context.Context, filter UserFilter) ([]*entity.User, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	// FindByRole returns all active users with the given role name (e.g. kaprodi,
	// dosen_pembimbing). Used for notifications and lecturer listings (Job 05).
	FindByRole(ctx context.Context, role string) ([]*entity.User, error)
	FindRoleByName(ctx context.Context, name string) (*entity.Role, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	BulkCreate(ctx context.Context, users []*entity.User) error
	SetActiveStatus(ctx context.Context, id uuid.UUID, isActive bool) error
	ResetPassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	InvalidateUserSessions(ctx context.Context, id uuid.UUID) error
}
