package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// DefenseFilter carries optional filters for listing defenses (Job 09).
type DefenseFilter struct {
	Status     string // filter by defense status
	ThesisID   uuid.UUID
	ExaminerID uuid.UUID // dosen_penguji → defenses they examine
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	PerPage    int

	// Role-based scoping (set by the use case layer):
	StudentID    uuid.UUID // mahasiswa → defenses of their theses
	SupervisorID uuid.UUID // dosen_pembimbing → defenses of theses they supervise
}

// DefenseRepository defines persistence operations for the defense module (Job 09).
type DefenseRepository interface {
	Create(ctx context.Context, defense *entity.ThesisDefense) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ThesisDefense, error)
	FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.ThesisDefense, error)
	FindAll(ctx context.Context, filter DefenseFilter) ([]*entity.ThesisDefense, int64, error)
	UpdateSchedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time, room string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateFinalScore(ctx context.Context, id uuid.UUID, score float64) error
	SetRevisionNotes(ctx context.Context, id uuid.UUID, notes string) error
	AssignExaminer(ctx context.Context, defenseID, examinerID, assignedBy uuid.UUID) error
	RemoveAllExaminers(ctx context.Context, defenseID uuid.UUID) error
	GetExaminers(ctx context.Context, defenseID uuid.UUID) ([]*entity.User, error)
	AddScore(ctx context.Context, score *entity.DefenseScore) error
	GetAllScores(ctx context.Context, defenseID uuid.UUID) ([]*entity.DefenseScore, error)
	HasExaminerScored(ctx context.Context, defenseID, examinerID uuid.UUID) (bool, error)
	// CountDistinctScoredExaminers returns the number of examiners who submitted scores.
	CountDistinctScoredExaminers(ctx context.Context, defenseID uuid.UUID) (int, error)
	// CheckScheduleConflict reports whether the room or one of the examiners is
	// already booked within ±2h on the same day.
	CheckScheduleConflict(ctx context.Context, room string, scheduledAt time.Time, examinerIDs []uuid.UUID, excludeID *uuid.UUID) (bool, error)
}
