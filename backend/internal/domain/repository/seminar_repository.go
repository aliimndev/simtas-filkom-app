package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// SeminarFilter carries optional filters for listing seminars (Job 08).
type SeminarFilter struct {
	Status     string // filter by seminar status
	ThesisID   uuid.UUID
	ExaminerID uuid.UUID // dosen_penguji → seminars they examine
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	PerPage    int

	// Role-based scoping (set by the use case layer):
	StudentID    uuid.UUID // mahasiswa → seminars of their theses
	SupervisorID uuid.UUID // dosen_pembimbing → seminars of theses they supervise
}

// SeminarRepository defines persistence operations for the seminar module (Job 08).
type SeminarRepository interface {
	Create(ctx context.Context, seminar *entity.Seminar) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Seminar, error)
	FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.Seminar, error)
	FindAll(ctx context.Context, filter SeminarFilter) ([]*entity.Seminar, int64, error)
	UpdateSchedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time, room string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateFinalScore(ctx context.Context, id uuid.UUID, score float64) error
	// UpdateNotes persists the revision notes recorded after a passed seminar.
	UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error
	AssignExaminer(ctx context.Context, seminarID, examinerID, assignedBy uuid.UUID) error
	RemoveAllExaminers(ctx context.Context, seminarID uuid.UUID) error
	GetExaminers(ctx context.Context, seminarID uuid.UUID) ([]*entity.User, error)
	AddScore(ctx context.Context, score *entity.SeminarScore) error
	GetAllScores(ctx context.Context, seminarID uuid.UUID) ([]*entity.SeminarScore, error)
	HasExaminerScored(ctx context.Context, seminarID, examinerID uuid.UUID) (bool, error)
	// CountDistinctScoredExaminers returns the number of examiners who submitted scores.
	CountDistinctScoredExaminers(ctx context.Context, seminarID uuid.UUID) (int, error)
	// CheckScheduleConflict reports whether the room or one of the examiners is
	// already booked within ±2h on the same day.
	CheckScheduleConflict(ctx context.Context, room string, scheduledAt time.Time, examinerIDs []uuid.UUID, excludeID *uuid.UUID) (bool, error)
}
