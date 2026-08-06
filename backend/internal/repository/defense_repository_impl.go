package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/grading"
)

type defenseRepository struct {
	db *gorm.DB
}

func NewDefenseRepository(db *gorm.DB) domainRepo.DefenseRepository {
	return &defenseRepository{db: db}
}

// preloadDefense applies the association preloads used for defense reads.
func preloadDefense(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Thesis.Student.Role").
		Preload("Thesis.AcademicYear").
		Preload("Thesis.Supervisors.Role").
		Preload("Examiners.Role").
		Preload("Scores.Examiner.Role")
}

func (r *defenseRepository) Create(ctx context.Context, defense *entity.ThesisDefense) error {
	return r.db.WithContext(ctx).Create(defense).Error
}

func (r *defenseRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ThesisDefense, error) {
	var defense entity.ThesisDefense
	err := preloadDefense(r.db.WithContext(ctx)).
		Where("id = ?", id).
		First(&defense).Error
	if err != nil {
		return nil, err
	}
	return &defense, nil
}

func (r *defenseRepository) FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.ThesisDefense, error) {
	var defense entity.ThesisDefense
	err := preloadDefense(r.db.WithContext(ctx)).
		Where("thesis_id = ?", thesisID).
		Order("created_at DESC").
		First(&defense).Error
	if err != nil {
		return nil, err
	}
	return &defense, nil
}

func (r *defenseRepository) FindAll(ctx context.Context, filter domainRepo.DefenseFilter) ([]*entity.ThesisDefense, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.ThesisDefense{})

	if filter.Status != "" {
		q = q.Where("thesis_defenses.status = ?", filter.Status)
	}
	if filter.ThesisID != uuid.Nil {
		q = q.Where("thesis_defenses.thesis_id = ?", filter.ThesisID)
	}
	if filter.DateFrom != nil {
		q = q.Where("thesis_defenses.scheduled_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		q = q.Where("thesis_defenses.scheduled_at <= ?", *filter.DateTo)
	}
	if filter.StudentID != uuid.Nil {
		q = q.Joins("JOIN theses t ON t.id = thesis_defenses.thesis_id").
			Where("t.student_id = ?", filter.StudentID)
	}
	if filter.SupervisorID != uuid.Nil {
		q = q.Where(`EXISTS (
			SELECT 1 FROM thesis_supervisors ts
			JOIN theses t2 ON t2.id = ts.thesis_id
			WHERE t2.id = thesis_defenses.thesis_id AND ts.supervisor_id = ?)`, filter.SupervisorID)
	}
	if filter.ExaminerID != uuid.Nil {
		q = q.Where(`EXISTS (
			SELECT 1 FROM defense_examiners de
			WHERE de.defense_id = thesis_defenses.id AND de.examiner_id = ?)`, filter.ExaminerID)
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

	var defenses []*entity.ThesisDefense
	if err := preloadDefense(q).
		Order("thesis_defenses.created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&defenses).Error; err != nil {
		return nil, 0, err
	}
	return defenses, total, nil
}

func (r *defenseRepository) UpdateSchedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time, room string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ThesisDefense{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"scheduled_at": scheduledAt,
			"room":         room,
		}).Error
}

func (r *defenseRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ThesisDefense{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *defenseRepository) UpdateFinalScore(ctx context.Context, id uuid.UUID, score float64) error {
	return r.db.WithContext(ctx).
		Model(&entity.ThesisDefense{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"final_score": score}).Error
}

func (r *defenseRepository) SetRevisionNotes(ctx context.Context, id uuid.UUID, notes string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ThesisDefense{}).
		Where("id = ?", id).
		Update("revision_notes", notes).Error
}

func (r *defenseRepository) AssignExaminer(ctx context.Context, defenseID, examinerID, assignedBy uuid.UUID) error {
	de := &entity.DefenseExaminer{
		DefenseID:  defenseID,
		ExaminerID: examinerID,
		AssignedBy: assignedBy,
	}
	return r.db.WithContext(ctx).Create(de).Error
}

func (r *defenseRepository) RemoveAllExaminers(ctx context.Context, defenseID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("defense_id = ?", defenseID).
		Delete(&entity.DefenseExaminer{}).Error
}

func (r *defenseRepository) GetExaminers(ctx context.Context, defenseID uuid.UUID) ([]*entity.User, error) {
	var examiners []*entity.User
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Joins("JOIN defense_examiners de ON de.examiner_id = users.id").
		Where("de.defense_id = ?", defenseID).
		Preload("Role").
		Find(&examiners).Error
	return examiners, err
}

func (r *defenseRepository) AddScore(ctx context.Context, score *entity.DefenseScore) error {
	return r.db.WithContext(ctx).Create(score).Error
}

func (r *defenseRepository) GetAllScores(ctx context.Context, defenseID uuid.UUID) ([]*entity.DefenseScore, error) {
	var scores []*entity.DefenseScore
	err := r.db.WithContext(ctx).
		Preload("Examiner.Role").
		Where("defense_id = ?", defenseID).
		Find(&scores).Error
	return scores, err
}

// defense statuses (mirrors usecase.DefenseStatus* to avoid an import cycle).
const (
	defStatusScheduled        = "scheduled"
	defStatusPassed           = "passed"
	defStatusFailed           = "failed"
	defStatusRevisionRequired = "revision_required"
)

// FinalizeDefense computes the final score and updates defense + thesis statuses
// atomically. A row lock (FOR UPDATE) guarantees that only one concurrent
// submission finalizes; late submitters observe status != "scheduled" and no-op.
func (r *defenseRepository) FinalizeDefense(ctx context.Context, defenseID uuid.UUID) (float64, string, uuid.UUID, error) {
	var finalScore float64
	var status string
	var thesisID uuid.UUID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var d entity.ThesisDefense
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", defenseID).
			First(&d).Error; err != nil {
			return err
		}
		// Already finalized (or not yet scheduled) → nothing to do.
		if d.Status != defStatusScheduled {
			return nil
		}

		var examiners []*entity.User
		if err := tx.Model(&entity.User{}).
			Joins("JOIN defense_examiners de ON de.examiner_id = users.id").
			Where("de.defense_id = ?", defenseID).
			Find(&examiners).Error; err != nil {
			return err
		}

		var scored int64
		if err := tx.Model(&entity.DefenseScore{}).
			Distinct("examiner_id").
			Where("defense_id = ?", defenseID).
			Count(&scored).Error; err != nil {
			return err
		}
		if int(scored) < len(examiners) {
			return nil
		}

		var scores []*entity.DefenseScore
		if err := tx.Where("defense_id = ?", defenseID).Find(&scores).Error; err != nil {
			return err
		}

		// Guard against a partial score set being finalized while another
		// examiner's inserts are still committing: each examiner must have
		// submitted every grading component, or we wait (no-op) for the rest to
		// land. Otherwise a racing submitter could be counted once with only a
		// subset of their components, skewing the final average.
		perExaminerComponents := map[string]map[string]bool{}
		for _, s := range scores {
			if _, ok := perExaminerComponents[s.ExaminerID.String()]; !ok {
				perExaminerComponents[s.ExaminerID.String()] = map[string]bool{}
			}
			perExaminerComponents[s.ExaminerID.String()][s.ComponentName] = true
		}
		for _, e := range examiners {
			components := perExaminerComponents[e.ID.String()]
			if len(components) < len(entity.DefenseGradingComponents) {
				return nil // examiner's score set is incomplete — wait
			}
		}

		order := []string{}
		perExaminer := map[string]float64{}
		for _, s := range scores {
			key := s.ExaminerID.String()
			if _, ok := perExaminer[key]; !ok {
				order = append(order, key)
			}
			perExaminer[key] += s.Score * s.ComponentWeight / 100.0
		}
		examinerScores := make([]float64, 0, len(order))
		for _, key := range order {
			examinerScores = append(examinerScores, perExaminer[key])
		}
		fs := grading.CalculateFinalScore(examinerScores)

		st := defStatusPassed
		switch {
		case fs < 60:
			st = defStatusFailed
		case fs < 75:
			st = defStatusRevisionRequired
		}

		if err := tx.Model(&entity.ThesisDefense{}).
			Where("id = ?", defenseID).
			Updates(map[string]interface{}{"final_score": fs, "status": st}).Error; err != nil {
			return err
		}

		finalScore = fs
		status = st
		thesisID = d.ThesisID
		return nil
	})
	if err != nil {
		return 0, "", uuid.Nil, err
	}
	return finalScore, status, thesisID, nil
}

func (r *defenseRepository) HasExaminerScored(ctx context.Context, defenseID, examinerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.DefenseScore{}).
		Where("defense_id = ? AND examiner_id = ?", defenseID, examinerID).
		Count(&count).Error
	return count > 0, err
}

func (r *defenseRepository) CountDistinctScoredExaminers(ctx context.Context, defenseID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.DefenseScore{}).
		Distinct("examiner_id").
		Where("defense_id = ?", defenseID).
		Count(&count).Error
	return int(count), err
}

func (r *defenseRepository) CheckScheduleConflict(ctx context.Context, room string, scheduledAt time.Time, examinerIDs []uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	from := scheduledAt.Add(-2 * time.Hour)
	to := scheduledAt.Add(2 * time.Hour)

	var roomCount int64
	roomQ := r.db.WithContext(ctx).Model(&entity.ThesisDefense{}).
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

	if len(examinerIDs) > 0 {
		examinerQ := r.db.WithContext(ctx).Model(&entity.DefenseExaminer{}).
			Joins("JOIN thesis_defenses d ON d.id = defense_examiners.defense_id").
			Where("defense_examiners.examiner_id IN ? AND d.scheduled_at >= ? AND d.scheduled_at <= ?",
				examinerIDs, from, to)
		if excludeID != nil {
			examinerQ = examinerQ.Where("d.id <> ?", *excludeID)
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
