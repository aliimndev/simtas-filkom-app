package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// ArchiveFilter carries optional filters for listing archives (Job 10).
type ArchiveFilter struct {
	Query          string // full-text search across title, abstract, keywords
	GraduationYear int
	FieldOfStudy   string
	StudyProgram   string
	SupervisorID   *uuid.UUID
	Page           int
	PerPage        int
}

// ArchiveStats groups the aggregate views used by the dashboard (Job 10).
type ArchiveStats struct {
	TotalArchives  int64                   `json:"total_archives"`
	ByYear         []ArchiveCountByYear    `json:"by_year"`
	ByField        []ArchiveCountByField   `json:"by_field"`
	ByStudyProgram []ArchiveCountByProgram `json:"by_study_program"`
}

// ArchiveCountByYear is one (year, count) row of the archives stats.
type ArchiveCountByYear struct {
	Year  int   `json:"year"`
	Count int64 `json:"count"`
}

// ArchiveCountByField is one (field_of_study, count) row of the archives stats.
type ArchiveCountByField struct {
	Field string `json:"field"`
	Count int64  `json:"count"`
}

// ArchiveCountByProgram is one (study_program, count) row of the archives stats.
type ArchiveCountByProgram struct {
	Program string `json:"program"`
	Count   int64  `json:"count"`
}

// ArchiveRepository defines persistence operations for the archive module (Job 10).
type ArchiveRepository interface {
	Create(ctx context.Context, archive *entity.ThesisArchive) error
	FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.ThesisArchive, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ThesisArchive, error)
	Search(ctx context.Context, filter ArchiveFilter) ([]*entity.ThesisArchive, int64, error)
	Stats(ctx context.Context) (*ArchiveStats, error)
}
