package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type seminarRepository struct {
	db *gorm.DB
}

func NewSeminarRepository(db *gorm.DB) domainRepo.SeminarRepository {
	return &seminarRepository{db: db}
}

// preloadSeminar applies the association preloads used for seminar reads.
func preloadSeminar(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Thesis.Student.Role").
		Preload("Thesis.AcademicYear").
		Preload("Thesis.Supervisors.Role").
		Preload("Examiners.Role").
		Preload("Scores.Examiner.Role")
}

func (r *seminarRepository) Create(ctx context.Context, seminar *entity.Seminar) error {
	return r.db.WithContext(ctx).Create(seminar).Error
}

func (r *seminarRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Seminar, error) {
	var seminar entity.Seminar
	err := preloadSeminar(r.db.WithContext(ctx)).
		Where("id = ?", id).
		First(&seminar).Error
	if err != nil {
		return nil, err
	}
	return &seminar, nil
}

func (r *seminarRepository) FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.Seminar, error) {
	var seminar entity.Seminar
	err := preloadSeminar(r.db.WithContext(ctx)).
		Where("thesis_id = ?", thesisID).
		Order("created_at DESC").
		First(&seminar).Error
	if err != nil {
		return nil, err
	}
	return &seminar, nil
}

func (r *seminarRepository) FindAll(ctx context.Context, filter domainRepo.SeminarFilter) ([]*entity.Seminar, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.Seminar{})

	if filter.Status != "" {
		q = q.Where("seminars.status = ?", filter.Status)
	}
	if filter.ThesisID != uuid.Nil {
		q = q.Where("seminars.thesis_id = ?", filter.ThesisID)
	}
	if filter.DateFrom != nil {
		q = q.Where("seminars.scheduled_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		q = q.Where("seminars.scheduled_at <= ?", *filter.DateTo)
	}
	if filter.StudentID != uuid.Nil {
		q = q.Joins("JOIN theses t ON t.id = seminars.thesis_id").
			Where("t.student_id = ?", filter.StudentID)
	}
	if filter.SupervisorID != uuid.Nil {
		q = q.Where(`EXISTS (
			SELECT 1 FROM thesis_supervisors ts
			JOIN theses t2 ON t2.id = ts.thesis_id
			WHERE t2.id = seminars.thesis_id AND ts.supervisor_id = ?)`, filter.SupervisorID)
	}
	if filter.ExaminerID != uuid.Nil {
		q = q.Where(`EXISTS (
			SELECT 1 FROM seminar_examiners se
			WHERE se.seminar_id = seminars.id AND se.examiner_id = ?)`, filter.ExaminerID)
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

	var seminars []*entity.Seminar
	if err := preloadSeminar(q).
		Order("seminars.created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&seminars).Error; err != nil {
		return nil, 0, err
	}
	return seminars, total, nil
}

func (r *seminarRepository) UpdateSchedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time, room string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Seminar{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"scheduled_at": scheduledAt,
			"room":         room,
		}).Error
}

func (r *seminarRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Seminar{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *seminarRepository) UpdateFinalScore(ctx context.Context, id uuid.UUID, score float64) error {
	return r.db.WithContext(ctx).
		Model(&entity.Seminar{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"final_score": score}).Error
}

func (r *seminarRepository) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Seminar{}).
		Where("id = ?", id).
		Update("notes", notes).Error
}

func (r *seminarRepository) AssignExaminer(ctx context.Context, seminarID, examinerID, assignedBy uuid.UUID) error {
	se := &entity.SeminarExaminer{
		SeminarID:  seminarID,
		ExaminerID: examinerID,
		AssignedBy: assignedBy,
	}
	return r.db.WithContext(ctx).Create(se).Error
}

func (r *seminarRepository) RemoveAllExaminers(ctx context.Context, seminarID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("seminar_id = ?", seminarID).
		Delete(&entity.SeminarExaminer{}).Error
}

func (r *seminarRepository) GetExaminers(ctx context.Context, seminarID uuid.UUID) ([]*entity.User, error) {
	var examiners []*entity.User
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Joins("JOIN seminar_examiners se ON se.examiner_id = users.id").
		Where("se.seminar_id = ?", seminarID).
		Preload("Role").
		Find(&examiners).Error
	return examiners, err
}

func (r *seminarRepository) AddScore(ctx context.Context, score *entity.SeminarScore) error {
	return r.db.WithContext(ctx).Create(score).Error
}

func (r *seminarRepository) GetAllScores(ctx context.Context, seminarID uuid.UUID) ([]*entity.SeminarScore, error) {
	var scores []*entity.SeminarScore
	err := r.db.WithContext(ctx).
		Preload("Examiner.Role").
		Where("seminar_id = ?", seminarID).
		Find(&scores).Error
	return scores, err
}

func (r *seminarRepository) HasExaminerScored(ctx context.Context, seminarID, examinerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.SeminarScore{}).
		Where("seminar_id = ? AND examiner_id = ?", seminarID, examinerID).
		Count(&count).Error
	return count > 0, err
}

func (r *seminarRepository) CountDistinctScoredExaminers(ctx context.Context, seminarID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.SeminarScore{}).
		Distinct("examiner_id").
		Where("seminar_id = ?", seminarID).
		Count(&count).Error
	return int(count), err
}

func (r *seminarRepository) CheckScheduleConflict(ctx context.Context, room string, scheduledAt time.Time, examinerIDs []uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	from := scheduledAt.Add(-2 * time.Hour)
	to := scheduledAt.Add(2 * time.Hour)

	// Room conflict: another seminar uses the same room within ±2h.
	var roomCount int64
	roomQ := r.db.WithContext(ctx).Model(&entity.Seminar{}).
		Where("room = ? AND scheduled_at >= ? AND scheduled_at <= ?", room, from, to)
	if excludeID != nil {
		roomQ = roomQ.Where("id <> ?", *excludeID)
	}
	if err := roomQ.Count(&roomCount).Error; err != nil {
		return false, err
	}
	if roomCount > 0 {
		return true, nil
	}

	// Examiner conflict: any examiner is assigned to another event in the window.
	if len(examinerIDs) > 0 {
		examinerQ := r.db.WithContext(ctx).Model(&entity.SeminarExaminer{}).
			Joins("JOIN seminars s ON s.id = seminar_examiners.seminar_id").
			Where("seminar_examiners.examiner_id IN ? AND s.scheduled_at >= ? AND s.scheduled_at <= ?",
				examinerIDs, from, to)
		if excludeID != nil {
			examinerQ = examinerQ.Where("s.id <> ?", *excludeID)
		}
		var examinerCount int64
		if err := examinerQ.Count(&examinerCount).Error; err != nil {
			return false, err
		}
		if examinerCount > 0 {
			return true, nil
		}
	}
	return false, nil
}
