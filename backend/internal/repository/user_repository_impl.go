package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainRepo.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(ctx context.Context, filter domainRepo.UserFilter) ([]*entity.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.User{}).Preload("Role")

	if filter.Role != "" {
		q = q.Joins("JOIN roles ON roles.id = users.role_id").
			Where("roles.name = ?", filter.Role)
	}
	if filter.IsActive != nil {
		q = q.Where("users.is_active = ?", *filter.IsActive)
	}
	if filter.StudyProgram != "" {
		q = q.Where("users.study_program = ?", filter.StudyProgram)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("(users.full_name ILIKE ? OR users.email ILIKE ? OR users.nim_nidn ILIKE ?)", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	var users []*entity.User
	if err := q.
		Order("users.created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
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

func (r *userRepository) FindByRole(ctx context.Context, role string) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ? AND users.is_active = ?", role, true).
		Order("users.full_name ASC").
		Find(&users).Error
	return users, err
}

func (r *userRepository) FindRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	// Omit the Role association so updating a user never touches the roles table.
	return r.db.WithContext(ctx).Omit("Role").Save(user).Error
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *userRepository) BulkCreate(ctx context.Context, users []*entity.User) error {
	if len(users) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, u := range users {
			if err := tx.Create(u).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *userRepository) SetActiveStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Update("is_active", isActive).Error
}

func (r *userRepository) ResetPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"password_hash":        passwordHash,
			"must_change_password": true,
		}).Error
}

func (r *userRepository) ChangePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"password_hash":        passwordHash,
			"must_change_password": false,
		}).Error
}

func (r *userRepository) InvalidateUserSessions(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
}
