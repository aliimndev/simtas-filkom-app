package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type archiveRepository struct {
	db *gorm.DB
}

func NewArchiveRepository(db *gorm.DB) domainRepo.ArchiveRepository {
	return &archiveRepository{db: db}
}

// preloadArchive applies the association preloads used for archive reads.
func preloadArchive(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Thesis.Student.Role").
		Preload("Thesis.AcademicYear").
		Preload("Thesis.Supervisors.Role").
		Preload("Archiver.Role")
}

func (r *archiveRepository) Create(ctx context.Context, archive *entity.ThesisArchive) error {
	return r.db.WithContext(ctx).Create(archive).Error
}

func (r *archiveRepository) FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.ThesisArchive, error) {
	var archive entity.ThesisArchive
	err := preloadArchive(r.db.WithContext(ctx)).
		Where("thesis_id = ?", thesisID).
		First(&archive).Error
	if err != nil {
		return nil, err
	}
	return &archive, nil
}

func (r *archiveRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ThesisArchive, error) {
	var archive entity.ThesisArchive
	err := preloadArchive(r.db.WithContext(ctx)).
		Where("id = ?", id).
		First(&archive).Error
	if err != nil {
		return nil, err
	}
	return &archive, nil
}

// Search lists archives with optional full-text search and filters.
// When Query is set, results are ranked by ts_rank over search_vector.
func (r *archiveRepository) Search(ctx context.Context, filter domainRepo.ArchiveFilter) ([]*entity.ThesisArchive, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&entity.ThesisArchive{}).
		Joins("JOIN theses ON theses.id = thesis_archives.thesis_id").
		Joins("JOIN users ON users.id = theses.student_id")

	if filter.Query != "" {
		q = q.Where("thesis_archives.search_vector @@ plainto_tsquery('simple', ?)", filter.Query)
	}
	if filter.GraduationYear > 0 {
		q = q.Where("thesis_archives.graduation_year = ?", filter.GraduationYear)
	}
	if filter.FieldOfStudy != "" {
		q = q.Where("theses.field_of_study = ?", filter.FieldOfStudy)
	}
	if filter.StudyProgram != "" {
		q = q.Where("users.study_program = ?", filter.StudyProgram)
	}
	if filter.SupervisorID != nil {
		q = q.Where(`EXISTS (
			SELECT 1 FROM thesis_supervisors ts
			WHERE ts.thesis_id = thesis_archives.thesis_id AND ts.supervisor_id = ?)`, *filter.SupervisorID)
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

	// Ranked order when searching; newest first otherwise.
	if filter.Query != "" {
		q = q.Order(gorm.Expr("ts_rank(thesis_archives.search_vector, plainto_tsquery('simple', ?)) DESC", filter.Query))
	} else {
		q = q.Order("thesis_archives.archived_at DESC")
	}

	var archives []*entity.ThesisArchive
	if err := preloadArchive(q).
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&archives).Error; err != nil {
		return nil, 0, err
	}
	return archives, total, nil
}

// Stats aggregates archive counts by year, field of study, and study program
// (used by the operational dashboard).
func (r *archiveRepository) Stats(ctx context.Context) (*domainRepo.ArchiveStats, error) {
	stats := &domainRepo.ArchiveStats{
		ByYear:         []domainRepo.ArchiveCountByYear{},
		ByField:        []domainRepo.ArchiveCountByField{},
		ByStudyProgram: []domainRepo.ArchiveCountByProgram{},
	}

	if err := r.db.WithContext(ctx).Model(&entity.ThesisArchive{}).Count(&stats.TotalArchives).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&entity.ThesisArchive{}).
		Select("graduation_year AS year, COUNT(*) AS count").
		Group("graduation_year").
		Order("graduation_year DESC").
		Scan(&stats.ByYear).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&entity.ThesisArchive{}).
		Joins("JOIN theses ON theses.id = thesis_archives.thesis_id").
		Select("theses.field_of_study AS field, COUNT(*) AS count").
		Group("theses.field_of_study").
		Order("count DESC").
		Scan(&stats.ByField).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&entity.ThesisArchive{}).
		Joins("JOIN theses ON theses.id = thesis_archives.thesis_id").
		Joins("JOIN users ON users.id = theses.student_id").
		Select("users.study_program AS program, COUNT(*) AS count").
		Group("users.study_program").
		Order("count DESC").
		Scan(&stats.ByStudyProgram).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
