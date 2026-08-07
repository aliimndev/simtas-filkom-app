package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

// fakeDashboardRepo implements DashboardRepository in-memory for tests.
type fakeDashboardRepo struct {
	summary   *domainRepo.AcademicSummary
	byStatus  []domainRepo.StatusCount
	trend     []domainRepo.MonthlyCount
	workload  []domainRepo.LecturerWorkload
	pending   *domainRepo.PendingActions
	schedules *domainRepo.UpcomingSchedules
	activity  *domainRepo.ActivityStats
	student   *domainRepo.StudentProgress
	superv    *domainRepo.SupervisorDashboard
	examiner  *domainRepo.ExaminerDashboard
}

func (f *fakeDashboardRepo) GetAcademicSummary(context.Context, domainRepo.DashboardFilter) (*domainRepo.AcademicSummary, error) {
	return f.summary, nil
}
func (f *fakeDashboardRepo) GetThesisByStatus(context.Context, domainRepo.DashboardFilter) ([]domainRepo.StatusCount, error) {
	return f.byStatus, nil
}
func (f *fakeDashboardRepo) GetGraduationTrend(context.Context, domainRepo.DashboardFilter) ([]domainRepo.MonthlyCount, error) {
	return f.trend, nil
}
func (f *fakeDashboardRepo) GetLecturerWorkload(context.Context, domainRepo.DashboardFilter) ([]domainRepo.LecturerWorkload, error) {
	return f.workload, nil
}
func (f *fakeDashboardRepo) GetPendingActions(context.Context) (*domainRepo.PendingActions, error) {
	return f.pending, nil
}
func (f *fakeDashboardRepo) GetUpcomingSchedules(context.Context, int) (*domainRepo.UpcomingSchedules, error) {
	return f.schedules, nil
}
func (f *fakeDashboardRepo) GetActivityStats(context.Context) (*domainRepo.ActivityStats, error) {
	return f.activity, nil
}
func (f *fakeDashboardRepo) GetStudentProgress(context.Context, uuid.UUID) (*domainRepo.StudentProgress, error) {
	return f.student, nil
}
func (f *fakeDashboardRepo) GetSupervisorDashboard(context.Context, uuid.UUID) (*domainRepo.SupervisorDashboard, error) {
	return f.superv, nil
}
func (f *fakeDashboardRepo) GetExaminerDashboard(context.Context, uuid.UUID) (*domainRepo.ExaminerDashboard, error) {
	return f.examiner, nil
}

func newTestDashboardUseCase(f *fakeDashboardRepo) *DashboardUseCase {
	return NewDashboardUseCase(f)
}

func TestDashboardSummary(t *testing.T) {
	uc := newTestDashboardUseCase(&fakeDashboardRepo{
		summary: &domainRepo.AcademicSummary{TotalActive: 85, TotalGraduated: 42, AvgCompletionMonths: 14.5},
		byStatus: []domainRepo.StatusCount{
			{Status: "submitted", Count: 8},
			{Status: "in_progress", Count: 35},
			{Status: "graduated", Count: 42},
		},
		trend: []domainRepo.MonthlyCount{{Month: "2026-08", Count: 5}, {Month: "2026-09", Count: 8}},
	})

	resp, err := uc.Summary(context.Background(), domainRepo.DashboardFilter{})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if resp.AcademicSummary.TotalActive != 85 {
		t.Errorf("total_active = %d, want 85", resp.AcademicSummary.TotalActive)
	}
	if resp.AcademicSummary.AvgCompletionMonths != 14.5 {
		t.Errorf("avg_completion_months = %v, want 14.5", resp.AcademicSummary.AvgCompletionMonths)
	}
	if len(resp.ByStatus) != 3 {
		t.Fatalf("by_status len = %d, want 3", len(resp.ByStatus))
	}
	// Label mapping applied.
	if resp.ByStatus[0].Label != "Menunggu Review" {
		t.Errorf("label for submitted = %q, want Menunggu Review", resp.ByStatus[0].Label)
	}
	if resp.GraduationTrend[0].Month != "2026-08" {
		t.Errorf("trend[0].month = %q, want 2026-08", resp.GraduationTrend[0].Month)
	}
}

func TestDashboardLecturerAnalytics(t *testing.T) {
	uc := newTestDashboardUseCase(&fakeDashboardRepo{
		workload: []domainRepo.LecturerWorkload{
			{LecturerID: uuid.New(), FullName: "Dr. X", NIDN: "001", SupervisionCount: 8, SeminarCount: 3, DefenseCount: 2},
			{LecturerID: uuid.New(), FullName: "Dr. Y", NIDN: "002", SupervisionCount: 1, SeminarCount: 5, DefenseCount: 1},
		},
	})

	resp, err := uc.LecturerAnalytics(context.Background(), domainRepo.DashboardFilter{})
	if err != nil {
		t.Fatalf("LecturerAnalytics: %v", err)
	}
	if len(resp.Lecturers) != 2 {
		t.Fatalf("lecturers len = %d, want 2", len(resp.Lecturers))
	}
	if resp.WorkloadDistribution.MaxSupervision != 8 {
		t.Errorf("max = %d, want 8", resp.WorkloadDistribution.MaxSupervision)
	}
	if resp.WorkloadDistribution.MinSupervision != 1 {
		t.Errorf("min = %d, want 1", resp.WorkloadDistribution.MinSupervision)
	}
	if resp.WorkloadDistribution.AvgSupervision != 4.5 {
		t.Errorf("avg = %v, want 4.5", resp.WorkloadDistribution.AvgSupervision)
	}
	if resp.WorkloadDistribution.HighestLoad == nil || resp.WorkloadDistribution.HighestLoad.FullName != "Dr. X" {
		t.Errorf("highest_load = %+v, want Dr. X", resp.WorkloadDistribution.HighestLoad)
	}
	if resp.WorkloadDistribution.LowestLoad == nil || resp.WorkloadDistribution.LowestLoad.FullName != "Dr. Y" {
		t.Errorf("lowest_load = %+v, want Dr. Y", resp.WorkloadDistribution.LowestLoad)
	}
}

func TestDashboardOperational(t *testing.T) {
	now := time.Now()
	uc := newTestDashboardUseCase(&fakeDashboardRepo{
		pending: &domainRepo.PendingActions{PendingTitleReviews: 5, PendingDocumentReviews: 12, PendingSeminars: 3, PendingDefenses: 2},
		schedules: &domainRepo.UpcomingSchedules{
			Seminars: []domainRepo.ScheduleItem{{ID: uuid.New(), StudentName: "Budi", ThesisTitle: "Judul", ScheduledAt: &now, Room: "Ruang A"}},
		},
		activity: &domainRepo.ActivityStats{LoginsToday: 15, DocumentsUploadedThisWeek: 8, ConsultationsThisWeek: 23},
	})

	resp, err := uc.Operational(context.Background(), 7)
	if err != nil {
		t.Fatalf("Operational: %v", err)
	}
	if resp.PendingActions.PendingTitleReviews != 5 || resp.PendingActions.PendingDocumentReviews != 12 {
		t.Errorf("pending actions mismatch: %+v", resp.PendingActions)
	}
	if len(resp.UpcomingSchedules.Seminars) != 1 {
		t.Errorf("seminars len = %d, want 1", len(resp.UpcomingSchedules.Seminars))
	}
	if resp.ActivityStats.LoginsToday != 15 {
		t.Errorf("logins_today = %d, want 15", resp.ActivityStats.LoginsToday)
	}
}

func TestDashboardStudent(t *testing.T) {
	lastCons := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	upcoming := time.Now().Add(48 * time.Hour)
	thesisID := uuid.New()
	uc := newTestDashboardUseCase(&fakeDashboardRepo{
		student: &domainRepo.StudentProgress{
			ThesisID: thesisID,
			Title:    "Judul Skripsi",
			Status:   "in_progress",
			Supervisors: []domainRepo.UserSummary{
				{FullName: "Dr. Ahmad", Email: "ahmad@test.local", NIM: "123456"},
			},
			Documents: []domainRepo.DocStatus{
				{Type: "proposal", Status: "approved", Version: 1},
				{Type: "seminar_doc", Status: "pending_review", Version: 1},
			},
			ConsultationCount: 8,
			LastConsultation:  &lastCons,
			UpcomingSeminar:   &domainRepo.ScheduleInfo{ID: uuid.New(), ScheduledAt: &upcoming, Room: "Ruang A", Status: "scheduled"},
		},
	})

	resp, err := uc.Student(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Student: %v", err)
	}
	if resp.ThesisID != thesisID {
		t.Errorf("thesis_id mismatch")
	}
	if resp.CurrentStage != "Proses Bimbingan" {
		t.Errorf("current_stage = %q, want Proses Bimbingan", resp.CurrentStage)
	}
	if resp.ProgressPercentage != 30 {
		t.Errorf("progress_percentage = %d, want 30", resp.ProgressPercentage)
	}
	if resp.ConsultationCount != 8 {
		t.Errorf("consultation_count = %d, want 8", resp.ConsultationCount)
	}
	if resp.UpcomingSeminar == nil {
		t.Fatal("upcoming_seminar should be set")
	}
	// pending_actions: waiting for seminar_doc review.
	found := false
	for _, a := range resp.PendingActions {
		if a == "Menunggu review dokumen seminar_doc" {
			found = true
		}
	}
	if !found {
		t.Errorf("pending_actions missing seminar_doc review: %v", resp.PendingActions)
	}
}

func TestDashboardStudentNoThesis(t *testing.T) {
	uc := newTestDashboardUseCase(&fakeDashboardRepo{student: nil})
	_, err := uc.Student(context.Background(), uuid.New())
	if err != ErrThesisNotFound {
		t.Errorf("expected ErrThesisNotFound, got %v", err)
	}
}

// TestDashboardStudentEmptyCollections guards against Go marshaling nil slices
// as `null`: the response must carry empty (non-nil) slices so the frontend
// can call len()/map() without crashing (regression: dashboard page threw
// "Cannot read properties of null (reading 'length')").
func TestDashboardStudentEmptyCollections(t *testing.T) {
	uc := newTestDashboardUseCase(&fakeDashboardRepo{
		// graduated → no pending_actions; no supervisors/documents attached.
		student: &domainRepo.StudentProgress{
			ThesisID: uuid.New(),
			Title:    "Judul",
			Status:   "graduated",
		},
	})

	resp, err := uc.Student(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Student: %v", err)
	}
	if resp.Documents == nil {
		t.Error("documents must be an empty slice ([]), not null")
	}
	if resp.Supervisors == nil {
		t.Error("supervisors must be an empty slice ([]), not null")
	}
	if resp.PendingActions == nil {
		t.Error("pending_actions must be an empty slice ([]), not null")
	}
}

func TestDashboardSupervisor(t *testing.T) {
	lastCons := time.Now().AddDate(0, 0, -12)
	uc := newTestDashboardUseCase(&fakeDashboardRepo{
		superv: &domainRepo.SupervisorDashboard{
			TotalStudents: 1,
			Students: []domainRepo.SupervisedStudent{
				{
					ThesisID:                  uuid.New(),
					Student:                   domainRepo.UserSummary{FullName: "Budi", NIM: "123"},
					Title:                     "Judul",
					Status:                    "in_progress",
					PendingDocumentReviews:    2,
					LastConsultation:          &lastCons,
					ConsultationCount:         5,
					DaysSinceLastConsultation: 12,
				},
			},
			PendingDocumentReviews: 2,
		},
	})

	resp, err := uc.Supervisor(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Supervisor: %v", err)
	}
	if resp.TotalStudents != 1 {
		t.Errorf("total_students = %d, want 1", resp.TotalStudents)
	}
	if len(resp.Students) != 1 {
		t.Fatalf("students len = %d, want 1", len(resp.Students))
	}
	if resp.Students[0].DaysSinceLastConsultation != 12 {
		t.Errorf("days_since_last_consultation = %d, want 12", resp.Students[0].DaysSinceLastConsultation)
	}
	if resp.PendingDocumentReviews != 2 {
		t.Errorf("pending_document_reviews = %d, want 2", resp.PendingDocumentReviews)
	}
}

func TestDashboardExaminer(t *testing.T) {
	upcoming := time.Now().Add(24 * time.Hour)
	uc := newTestDashboardUseCase(&fakeDashboardRepo{
		examiner: &domainRepo.ExaminerDashboard{
			UpcomingAssignments: []domainRepo.ExaminerAssignment{
				{Type: "seminar", ID: uuid.New(), ThesisTitle: "Judul", StudentName: "Budi", ScheduledAt: &upcoming, Room: "Ruang A", HasScored: false},
			},
			PendingScores: []domainRepo.ExaminerAssignment{
				{Type: "defense", ID: uuid.New(), ThesisTitle: "Judul 2", StudentName: "Ani", ScheduledAt: &upcoming, Room: "Ruang B", HasScored: false},
			},
			ScoringHistory: []domainRepo.ExaminerAssignment{
				{Type: "seminar", ID: uuid.New(), ThesisTitle: "Judul 3", StudentName: "Citra", HasScored: true},
			},
		},
	})

	resp, err := uc.Examiner(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Examiner: %v", err)
	}
	if len(resp.UpcomingAssignments) != 1 || len(resp.PendingScores) != 1 || len(resp.ScoringHistory) != 1 {
		t.Errorf("assignment buckets wrong: %d/%d/%d",
			len(resp.UpcomingAssignments), len(resp.PendingScores), len(resp.ScoringHistory))
	}
	if resp.UpcomingAssignments[0].Type != "seminar" {
		t.Errorf("upcoming[0].type = %q, want seminar", resp.UpcomingAssignments[0].Type)
	}
}

// ensure fakeDashboardRepo implements the full interface at compile time.
var _ domainRepo.DashboardRepository = (*fakeDashboardRepo)(nil)
