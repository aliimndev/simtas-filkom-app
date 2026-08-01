package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/api/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/api/internal/domain/repository"
)

var (
	ErrAcademicYearNotFound = errors.New("tahun akademik tidak ditemukan")
	ErrInvalidSemester      = errors.New("semester harus ganjil atau genap")
	ErrInvalidDateRange     = errors.New("tanggal akhir harus setelah tanggal mulai")
	ErrAcademicYearInUse    = errors.New("tahun akademik aktif dengan skripsi berjalan tidak dapat diubah")
)

// AcademicYearRequest is the payload for create/update academic year.
type AcademicYearRequest struct {
	Name      string `json:"name" binding:"required"`
	Semester  string `json:"semester" binding:"required"`
	StartDate string `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate   string `json:"end_date" binding:"required"`   // YYYY-MM-DD
}

// AcademicYearUseCase contains business logic for academic year management.
type AcademicYearUseCase struct {
	repo domainRepo.AcademicYearRepository
}

func NewAcademicYearUseCase(repo domainRepo.AcademicYearRepository) *AcademicYearUseCase {
	return &AcademicYearUseCase{repo: repo}
}

// List returns all academic years, newest first (sorted by start_date DESC).
func (uc *AcademicYearUseCase) List(ctx context.Context) ([]*entity.AcademicYear, error) {
	return uc.repo.FindAll(ctx)
}

// Create validates and creates a new academic year.
func (uc *AcademicYearUseCase) Create(ctx context.Context, req AcademicYearRequest) (*entity.AcademicYear, error) {
	year, err := buildAcademicYear(req)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, year); err != nil {
		return nil, err
	}
	return year, nil
}

// Update updates an academic year unless it is active with ongoing theses.
func (uc *AcademicYearUseCase) Update(ctx context.Context, id uuid.UUID, req AcademicYearRequest) (*entity.AcademicYear, error) {
	existing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAcademicYearNotFound
		}
		return nil, err
	}

	if existing.IsActive {
		count, err := uc.repo.CountActiveTheses(ctx, id)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrAcademicYearInUse
		}
	}

	year, err := buildAcademicYear(req)
	if err != nil {
		return nil, err
	}
	year.ID = existing.ID
	year.IsActive = existing.IsActive
	year.CreatedAt = existing.CreatedAt

	if err := uc.repo.Update(ctx, year); err != nil {
		return nil, err
	}
	return year, nil
}

// Activate sets this year active and deactivates all others (only 1 active).
func (uc *AcademicYearUseCase) Activate(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAcademicYearNotFound
		}
		return err
	}
	return uc.repo.Activate(ctx, id)
}

func buildAcademicYear(req AcademicYearRequest) (*entity.AcademicYear, error) {
	if req.Semester != "ganjil" && req.Semester != "genap" {
		return nil, ErrInvalidSemester
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("format tanggal mulai tidak valid (YYYY-MM-DD)")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("format tanggal akhir tidak valid (YYYY-MM-DD)")
	}
	if !endDate.After(startDate) {
		return nil, ErrInvalidDateRange
	}

	return &entity.AcademicYear{
		Name:      req.Name,
		Semester:  req.Semester,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}
