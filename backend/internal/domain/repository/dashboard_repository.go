package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DashboardFilter narrows admin/kaprodi dashboard queries.
type DashboardFilter struct {
	AcademicYearID *uuid.UUID
	Semester       string // ganjil / genap
	StudyProgram   string
}

// AcademicSummary aggregates high-level thesis statistics (Job 12).
type AcademicSummary struct {
	TotalActive         int
	TotalGraduated      int
	AvgCompletionMonths float64
}

// StatusCount is a single thesis-status bucket (Job 12).
type StatusCount struct {
	Status string
	Count  int
}

// MonthlyCount is a single month bucket, e.g. "2026-10" (Job 12).
type MonthlyCount struct {
	Month string
	Count int
}

// LecturerWorkload describes one lecturer's supervision/examination load (Job 12).
type LecturerWorkload struct {
	LecturerID       uuid.UUID
	FullName         string
	NIDN             string
	SupervisionCount int
	SeminarCount     int
	DefenseCount     int
}

// PendingActions counts items awaiting a human decision (Job 12).
type PendingActions struct {
	PendingTitleReviews    int
	PendingDocumentReviews int
	PendingSeminars        int
	PendingDefenses        int
}

// ScheduleItem is one upcoming seminar/defense row (Job 12).
type ScheduleItem struct {
	ID          uuid.UUID
	StudentName string
	ThesisTitle string
	ScheduledAt *time.Time
	Room        string
}

// UpcomingSchedules groups upcoming seminars and defenses (Job 12).
type UpcomingSchedules struct {
	Seminars []ScheduleItem
	Defenses []ScheduleItem
}

// ActivityStats aggregates recent system activity (Job 12).
type ActivityStats struct {
	LoginsToday               int
	DocumentsUploadedThisWeek int
	ConsultationsThisWeek     int
}

// UserSummary is a compact user reference used across dashboard responses.
type UserSummary struct {
	FullName string
	Email    string
	NIM      string
}

// DocStatus describes one document version in the student progress view.
type DocStatus struct {
	Type    string
	Status  string
	Version int
}

// ScheduleInfo describes an upcoming seminar/defense for a student.
type ScheduleInfo struct {
	ID          uuid.UUID
	ScheduledAt *time.Time
	Room        string
	Status      string
}

// StudentProgress is the per-student dashboard payload (Job 12).
type StudentProgress struct {
	ThesisID          uuid.UUID
	Title             string
	Status            string
	Supervisors       []UserSummary
	Documents         []DocStatus
	ConsultationCount int
	LastConsultation  *time.Time
	UpcomingSeminar   *ScheduleInfo
	UpcomingDefense   *ScheduleInfo
}

// SupervisedStudent is one thesis row in a supervisor's dashboard (Job 12).
type SupervisedStudent struct {
	ThesisID                  uuid.UUID
	Student                   UserSummary
	Title                     string
	Status                    string
	PendingDocumentReviews    int
	LastConsultation          *time.Time
	ConsultationCount         int
	DaysSinceLastConsultation int
}

// SupervisorDashboard aggregates a supervisor's students and schedules (Job 12).
type SupervisorDashboard struct {
	TotalStudents          int
	Students               []SupervisedStudent
	PendingDocumentReviews int
	UpcomingSchedules      UpcomingSchedules
}

// ExaminerAssignment is one seminar/defense row in an examiner's dashboard (Job 12).
type ExaminerAssignment struct {
	Type        string // seminar | defense
	ID          uuid.UUID
	ThesisTitle string
	StudentName string
	ScheduledAt *time.Time
	Room        string
	HasScored   bool
}

// ExaminerDashboard aggregates an examiner's assignments (Job 12).
type ExaminerDashboard struct {
	UpcomingAssignments []ExaminerAssignment
	PendingScores       []ExaminerAssignment
	ScoringHistory      []ExaminerAssignment
}

// DashboardRepository provides aggregated statistics for the dashboard module (Job 12).
type DashboardRepository interface {
	// Admin & Kaprodi
	GetAcademicSummary(ctx context.Context, filter DashboardFilter) (*AcademicSummary, error)
	GetThesisByStatus(ctx context.Context, filter DashboardFilter) ([]StatusCount, error)
	GetGraduationTrend(ctx context.Context, filter DashboardFilter) ([]MonthlyCount, error)
	GetLecturerWorkload(ctx context.Context, filter DashboardFilter) ([]LecturerWorkload, error)
	GetPendingActions(ctx context.Context) (*PendingActions, error)
	GetUpcomingSchedules(ctx context.Context, days int) (*UpcomingSchedules, error)
	GetActivityStats(ctx context.Context) (*ActivityStats, error)

	// Mahasiswa
	GetStudentProgress(ctx context.Context, studentID uuid.UUID) (*StudentProgress, error)

	// Dosen Pembimbing
	GetSupervisorDashboard(ctx context.Context, supervisorID uuid.UUID) (*SupervisorDashboard, error)

	// Dosen Penguji
	GetExaminerDashboard(ctx context.Context, examinerID uuid.UUID) (*ExaminerDashboard, error)
}
