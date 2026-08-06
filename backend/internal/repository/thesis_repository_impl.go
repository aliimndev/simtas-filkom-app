package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type thesisRepository struct {
	db *gorm.DB
}

func NewThesisRepository(db *gorm.DB) domainRepo.ThesisRepository {
	return &thesisRepository{db: db}
}

// preloadThesis applies the standard association preloads used for thesis reads.
func preloadThesis(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Student.Role").
		Preload("AcademicYear").
		Preload("Supervisors.Role")
}

func (r *thesisRepository) Create(ctx context.Context, thesis *entity.Thesis) error {
	return r.db.WithContext(ctx).Create(thesis).Error
}

func (r *thesisRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Thesis, error) {
	var thesis entity.Thesis
	err := preloadThesis(r.db.WithContext(ctx)).
		Where("id = ?", id).
		First(&thesis).Error
	if err != nil {
		return nil, err
	}
	return &thesis, nil
}

func (r *thesisRepository) FindAll(ctx context.Context, filter domainRepo.ThesisFilter) ([]*entity.Thesis, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.Thesis{})

	if filter.Status != "" {
		q = q.Where("theses.status = ?", filter.Status)
	}
	if filter.AcademicYearID != uuid.Nil {
		q = q.Where("theses.academic_year_id = ?", filter.AcademicYearID)
	}
	if filter.FieldOfStudy != "" {
		q = q.Where("theses.field_of_study = ?", filter.FieldOfStudy)
	}
	if filter.StudentID != uuid.Nil {
		q = q.Where("theses.student_id = ?", filter.StudentID)
	}
	if filter.SupervisorID != uuid.Nil {
		q = q.Where(`EXISTS (
			SELECT 1 FROM thesis_supervisors ts
			WHERE ts.thesis_id = theses.id AND ts.supervisor_id = ?)`, filter.SupervisorID)
	}
	if filter.ExaminerID != uuid.Nil {
		q = q.Where(`(EXISTS (
				SELECT 1 FROM defense_examiners de
				JOIN thesis_defenses d ON d.id = de.defense_id
				WHERE d.thesis_id = theses.id AND de.examiner_id = ?)
			OR EXISTS (
				SELECT 1 FROM seminar_examiners se
				JOIN seminars s ON s.id = se.seminar_id
				WHERE s.thesis_id = theses.id AND se.examiner_id = ?))`,
			filter.ExaminerID, filter.ExaminerID)
	}
	if filter.StudyProgram != "" || filter.Search != "" {
		q = q.Joins("JOIN users su ON su.id = theses.student_id")
		if filter.StudyProgram != "" {
			q = q.Where("su.study_program = ?", filter.StudyProgram)
		}
		if filter.Search != "" {
			like := "%" + filter.Search + "%"
			q = q.Where("(theses.title ILIKE ? OR su.full_name ILIKE ? OR su.nim_nidn ILIKE ?)", like, like, like)
		}
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

	var theses []*entity.Thesis
	if err := preloadThesis(q).
		Order("theses.submitted_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&theses).Error; err != nil {
		return nil, 0, err
	}
	return theses, total, nil
}

func (r *thesisRepository) FindByStudentID(ctx context.Context, studentID uuid.UUID) (*entity.Thesis, error) {
	var thesis entity.Thesis
	err := preloadThesis(r.db.WithContext(ctx)).
		Where("student_id = ?", studentID).
		Order("submitted_at DESC").
		First(&thesis).Error
	if err != nil {
		return nil, err
	}
	return &thesis, nil
}

func (r *thesisRepository) FindActiveByStudentID(ctx context.Context, studentID uuid.UUID) (*entity.Thesis, error) {
	var thesis entity.Thesis
	err := r.db.WithContext(ctx).
		Where("student_id = ? AND status NOT IN ?", studentID, []string{"cancelled", "graduated"}).
		Order("submitted_at DESC").
		First(&thesis).Error
	if err != nil {
		return nil, err
	}
	return &thesis, nil
}

func (r *thesisRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string) error {
	updates := map[string]interface{}{"status": status}
	if notes != "" {
		updates["kaprodi_notes"] = notes
	}
	return r.db.WithContext(ctx).
		Model(&entity.Thesis{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *thesisRepository) Update(ctx context.Context, thesis *entity.Thesis) error {
	return r.db.WithContext(ctx).
		Omit("Student", "AcademicYear", "Supervisors").
		Save(thesis).Error
}

func (r *thesisRepository) AssignSupervisor(ctx context.Context, thesisID, supervisorID, assignedBy uuid.UUID) error {
	ts := &entity.ThesisSupervisor{
		ThesisID:     thesisID,
		SupervisorID: supervisorID,
		AssignedBy:   assignedBy,
		AssignedAt:   time.Now(),
	}
	return r.db.WithContext(ctx).Create(ts).Error
}

// AssignSupervisors assigns multiple supervisors and flips the thesis to
// in_progress atomically. It is the transactional counterpart to the
// single-supervisor AssignSupervisor used elsewhere.
func (r *thesisRepository) AssignSupervisors(ctx context.Context, thesisID uuid.UUID, supervisorIDs []uuid.UUID, assignedBy uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sid := range supervisorIDs {
			ts := &entity.ThesisSupervisor{
				ThesisID:     thesisID,
				SupervisorID: sid,
				AssignedBy:   assignedBy,
				AssignedAt:   time.Now(),
			}
			if err := tx.Create(ts).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&entity.Thesis{}).
			Where("id = ?", thesisID).
			Update("status", "in_progress").Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *thesisRepository) GetSupervisors(ctx context.Context, thesisID uuid.UUID) ([]*entity.User, error) {
	var supervisors []*entity.User
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Joins("JOIN thesis_supervisors ts ON ts.supervisor_id = users.id").
		Where("ts.thesis_id = ?", thesisID).
		Preload("Role").
		Find(&supervisors).Error
	return supervisors, err
}

func (r *thesisRepository) CountActiveSupervisions(ctx context.Context, supervisorID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.ThesisSupervisor{}).
		Joins("JOIN theses t ON t.id = thesis_supervisors.thesis_id").
		Where("thesis_supervisors.supervisor_id = ? AND t.status NOT IN ? AND t.deleted_at IS NULL", supervisorID, []string{"cancelled", "graduated"}).
		Count(&count).Error
	return count, err
}

func (r *thesisRepository) IsExaminer(ctx context.Context, thesisID, examinerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ThesisDefense{}).
		Joins("JOIN defense_examiners de ON de.defense_id = thesis_defenses.id").
		Where("thesis_defenses.thesis_id = ? AND de.examiner_id = ?", thesisID, examinerID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	err = r.db.WithContext(ctx).Model(&entity.Seminar{}).
		Joins("JOIN seminar_examiners se ON se.seminar_id = seminars.id").
		Where("seminars.thesis_id = ? AND se.examiner_id = ?", thesisID, examinerID).
		Count(&count).Error
	return count > 0, err
}
