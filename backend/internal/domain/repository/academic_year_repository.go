package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// AcademicYearRepository defines persistence for academic year management (Job 04).
type AcademicYearRepository interface {
	FindAll(ctx context.Context) ([]*entity.AcademicYear, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.AcademicYear, error)
	// FindActive returns the currently active academic year (Job 05: theses are
	// attached to the active year). Returns gorm.ErrRecordNotFound if none.
	FindActive(ctx context.Context) (*entity.AcademicYear, error)
	Create(ctx context.Context, year *entity.AcademicYear) error
	Update(ctx context.Context, year *entity.AcademicYear) error
	// Activate sets the given year active and deactivates all others (single active).
	Activate(ctx context.Context, id uuid.UUID) error
	// CountActiveTheses counts non-graduated theses attached to an academic year.
	CountActiveTheses(ctx context.Context, academicYearID uuid.UUID) (int64, error)
}
