package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type academicYearRepository struct {
	db *gorm.DB
}

func NewAcademicYearRepository(db *gorm.DB) domainRepo.AcademicYearRepository {
	return &academicYearRepository{db: db}
}

func (r *academicYearRepository) FindAll(ctx context.Context) ([]*entity.AcademicYear, error) {
	var years []*entity.AcademicYear
	err := r.db.WithContext(ctx).
		Order("start_date DESC").
		Find(&years).Error
	return years, err
}

func (r *academicYearRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.AcademicYear, error) {
	var year entity.AcademicYear
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&year).Error
	if err != nil {
		return nil, err
	}
	return &year, nil
}

func (r *academicYearRepository) FindActive(ctx context.Context) (*entity.AcademicYear, error) {
	var year entity.AcademicYear
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		First(&year).Error
	if err != nil {
		return nil, err
	}
	return &year, nil
}

func (r *academicYearRepository) Create(ctx context.Context, year *entity.AcademicYear) error {
	return r.db.WithContext(ctx).Create(year).Error
}

func (r *academicYearRepository) Update(ctx context.Context, year *entity.AcademicYear) error {
	return r.db.WithContext(ctx).Save(year).Error
}

// Activate sets the given year active and deactivates all others (single active).
func (r *academicYearRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.AcademicYear{}).
			Where("is_active = ?", true).
			Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Model(&entity.AcademicYear{}).
			Where("id = ?", id).
			Update("is_active", true).Error
	})
}

func (r *academicYearRepository) CountActiveTheses(ctx context.Context, academicYearID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Thesis{}).
		Where("academic_year_id = ? AND status <> ?", academicYearID, "graduated").
		Count(&count).Error
	return count, err
}
