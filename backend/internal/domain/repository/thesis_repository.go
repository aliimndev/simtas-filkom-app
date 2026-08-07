package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// ThesisFilter carries optional filters for listing theses (Job 05).
type ThesisFilter struct {
	Status         string    // filter by thesis status
	AcademicYearID uuid.UUID // filter by academic year
	StudyProgram   string    // filter by student's study program
	FieldOfStudy   string    // filter by field of study
	SupervisorID   uuid.UUID // filter by supervisor
	Search         string    // matches thesis title, student full_name, or NIM
	Page           int
	PerPage        int

	// Role-based scoping (set by the use case layer):
	StudentID  uuid.UUID // mahasiswa → only own theses
	ExaminerID uuid.UUID // dosen_penguji → theses they examine
}

// ThesisRepository defines persistence operations for thesis submission (Job 05).
type ThesisRepository interface {
	Create(ctx context.Context, thesis *entity.Thesis) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Thesis, error)
	FindAll(ctx context.Context, filter ThesisFilter) ([]*entity.Thesis, int64, error)
	FindByStudentID(ctx context.Context, studentID uuid.UUID) (*entity.Thesis, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, notes string) error
	Update(ctx context.Context, thesis *entity.Thesis) error
	AssignSupervisor(ctx context.Context, thesisID, supervisorID, assignedBy uuid.UUID) error
	// AssignSupervisors atomically assigns 1–2 supervisors and moves the thesis
	// to in_progress within a single database transaction, so a partial failure
	// (e.g. one insert succeeds then UpdateStatus fails) can never leave an
	// inconsistent state.
	AssignSupervisors(ctx context.Context, thesisID uuid.UUID, supervisorIDs []uuid.UUID, assignedBy uuid.UUID) error
	GetSupervisors(ctx context.Context, thesisID uuid.UUID) ([]*entity.User, error)
	FindActiveByStudentID(ctx context.Context, studentID uuid.UUID) (*entity.Thesis, error)
	// CountActiveSupervisions returns the number of non-finished theses a lecturer supervises.
	CountActiveSupervisions(ctx context.Context, supervisorID uuid.UUID) (int64, error)
	// IsExaminer reports whether the user is assigned as an examiner for the thesis.
	IsExaminer(ctx context.Context, thesisID, examinerID uuid.UUID) (bool, error)
}
