package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

// StageProgress maps thesis status → progress percentage (Job 12 spec).
var StageProgress = map[string]int{
	"submitted":     10,
	"approved":      15,
	"in_progress":   30,
	"seminar_ready": 45,
	"seminar_done":  60,
	"defense_ready": 75,
	"defense_done":  90,
	"graduated":     100,
}

// stageLabel maps thesis status → human-readable current stage label.
var stageLabel = map[string]string{
	"submitted":     "Menunggu Review Judul",
	"approved":      "Judul Disetujui",
	"in_progress":   "Proses Bimbingan",
	"seminar_ready": "Siap Seminar",
	"seminar_done":  "Pasca Seminar",
	"defense_ready": "Siap Sidang",
	"defense_done":  "Pasca Sidang",
	"graduated":     "Lulus",
}

// StatusLabel maps thesis status → display label for chart buckets.
var StatusLabel = map[string]string{
	"submitted":     "Menunggu Review",
	"approved":      "Judul Disetujui",
	"in_progress":   "Bimbingan",
	"seminar_ready": "Siap Seminar",
	"seminar_done":  "Pasca Seminar",
	"defense_ready": "Siap Sidang",
	"defense_done":  "Pasca Sidang",
	"graduated":     "Lulus",
	"cancelled":     "Dibatalkan",
}

// ── Response shapes ───────────────────────────────────────────────────────

type DashboardSummaryResponse struct {
	AcademicSummary AcademicSummaryView `json:"academic_summary"`
	ByStatus        []StatusCountView   `json:"by_status"`
	GraduationTrend []MonthlyCountView  `json:"graduation_trend"`
}

type AcademicSummaryView struct {
	TotalActive         int     `json:"total_active"`
	TotalGraduated      int     `json:"total_graduated"`
	AvgCompletionMonths float64 `json:"avg_completion_months"`
}

type StatusCountView struct {
	Status string `json:"status"`
	Label  string `json:"label"`
	Count  int    `json:"count"`
}

type MonthlyCountView struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type LecturerAnalyticsResponse struct {
	Lecturers            []LecturerWorkloadView `json:"lecturers"`
	WorkloadDistribution WorkloadDistribution   `json:"workload_distribution"`
}

type LecturerWorkloadView struct {
	LecturerID       uuid.UUID `json:"lecturer_id"`
	FullName         string    `json:"full_name"`
	NIDN             string    `json:"nidn"`
	SupervisionCount int       `json:"supervision_count"`
	SeminarCount     int       `json:"seminar_count"`
	DefenseCount     int       `json:"defense_count"`
}

type WorkloadDistribution struct {
	MaxSupervision int                `json:"max_supervision"`
	MinSupervision int                `json:"min_supervision"`
	AvgSupervision float64            `json:"avg_supervision"`
	HighestLoad    *LecturerLoadBrief `json:"highest_load"`
	LowestLoad     *LecturerLoadBrief `json:"lowest_load"`
}

type LecturerLoadBrief struct {
	FullName         string `json:"full_name"`
	SupervisionCount int    `json:"supervision_count"`
}

type OperationalResponse struct {
	PendingActions    PendingActionsView    `json:"pending_actions"`
	UpcomingSchedules UpcomingSchedulesView `json:"upcoming_schedules"`
	ActivityStats     ActivityStatsView     `json:"activity_stats"`
}

type PendingActionsView struct {
	PendingTitleReviews    int `json:"pending_title_reviews"`
	PendingDocumentReviews int `json:"pending_document_reviews"`
	PendingSeminars        int `json:"pending_seminars"`
	PendingDefenses        int `json:"pending_defenses"`
}

type UpcomingSchedulesView struct {
	Seminars []ScheduleItemView `json:"seminars"`
	Defenses []ScheduleItemView `json:"defenses"`
}

type ScheduleItemView struct {
	ID          uuid.UUID `json:"id"`
	StudentName string    `json:"student_name"`
	ThesisTitle string    `json:"thesis_title"`
	ScheduledAt *string   `json:"scheduled_at,omitempty"`
	Room        string    `json:"room"`
}

type ActivityStatsView struct {
	LoginsToday               int `json:"logins_today"`
	DocumentsUploadedThisWeek int `json:"documents_uploaded_this_week"`
	ConsultationsThisWeek     int `json:"consultations_this_week"`
}

type StudentProgressResponse struct {
	ThesisID           uuid.UUID         `json:"thesis_id"`
	Title              string            `json:"title"`
	Status             string            `json:"status"`
	CurrentStage       string            `json:"current_stage"`
	ProgressPercentage int               `json:"progress_percentage"`
	Supervisors        []UserSummaryView `json:"supervisors"`
	Documents          []DocStatusView   `json:"documents"`
	ConsultationCount  int               `json:"consultation_count"`
	LastConsultation   *string           `json:"last_consultation,omitempty"`
	UpcomingSeminar    *ScheduleInfoView `json:"upcoming_seminar"`
	UpcomingDefense    *ScheduleInfoView `json:"upcoming_defense"`
	PendingActions     []string          `json:"pending_actions"`
}

type UserSummaryView struct {
	FullName string `json:"full_name"`
	Email    string `json:"email,omitempty"`
	NIM      string `json:"nim,omitempty"`
}

type DocStatusView struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Version int    `json:"version"`
}

type ScheduleInfoView struct {
	ID          uuid.UUID `json:"id"`
	ScheduledAt *string   `json:"scheduled_at,omitempty"`
	Room        string    `json:"room"`
	Status      string    `json:"status"`
}

type SupervisorDashboardResponse struct {
	TotalStudents          int                     `json:"total_students"`
	Students               []SupervisedStudentView `json:"students"`
	PendingDocumentReviews int                     `json:"pending_document_reviews"`
	UpcomingSchedules      UpcomingSchedulesView   `json:"upcoming_schedules"`
}

type SupervisedStudentView struct {
	ThesisID                  uuid.UUID       `json:"thesis_id"`
	Student                   UserSummaryView `json:"student"`
	Title                     string          `json:"title"`
	Status                    string          `json:"status"`
	CurrentStage              string          `json:"current_stage"`
	PendingDocumentReviews    int             `json:"pending_document_reviews"`
	LastConsultation          *string         `json:"last_consultation,omitempty"`
	ConsultationCount         int             `json:"consultation_count"`
	DaysSinceLastConsultation int             `json:"days_since_last_consultation"`
}

type ExaminerDashboardResponse struct {
	UpcomingAssignments []ExaminerAssignmentView `json:"upcoming_assignments"`
	PendingScores       []ExaminerAssignmentView `json:"pending_scores"`
	ScoringHistory      []ExaminerAssignmentView `json:"scoring_history"`
}

type ExaminerAssignmentView struct {
	Type        string    `json:"type"`
	ID          uuid.UUID `json:"id"`
	ThesisTitle string    `json:"thesis_title"`
	StudentName string    `json:"student_name"`
	ScheduledAt *string   `json:"scheduled_at,omitempty"`
	Room        string    `json:"room"`
	HasScored   bool      `json:"has_scored"`
}

// ── Use case ──────────────────────────────────────────────────────────────

type DashboardUseCase struct {
	repo domainRepo.DashboardRepository
}

func NewDashboardUseCase(repo domainRepo.DashboardRepository) *DashboardUseCase {
	return &DashboardUseCase{repo: repo}
}

// Summary returns the admin/kaprodi academic summary (Job 12).
func (uc *DashboardUseCase) Summary(ctx context.Context, filter domainRepo.DashboardFilter) (*DashboardSummaryResponse, error) {
	academic, err := uc.repo.GetAcademicSummary(ctx, filter)
	if err != nil {
		return nil, err
	}
	byStatus, err := uc.repo.GetThesisByStatus(ctx, filter)
	if err != nil {
		return nil, err
	}
	trend, err := uc.repo.GetGraduationTrend(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := &DashboardSummaryResponse{
		AcademicSummary: AcademicSummaryView{
			TotalActive:         academic.TotalActive,
			TotalGraduated:      academic.TotalGraduated,
			AvgCompletionMonths: academic.AvgCompletionMonths,
		},
	}
	for _, s := range byStatus {
		label := StatusLabel[s.Status]
		if label == "" {
			label = s.Status
		}
		resp.ByStatus = append(resp.ByStatus, StatusCountView{Status: s.Status, Label: label, Count: s.Count})
	}
	for _, m := range trend {
		resp.GraduationTrend = append(resp.GraduationTrend, MonthlyCountView{Month: m.Month, Count: m.Count})
	}
	return resp, nil
}

// LecturerAnalytics returns per-lecturer workload plus distribution stats (Job 12).
func (uc *DashboardUseCase) LecturerAnalytics(ctx context.Context, filter domainRepo.DashboardFilter) (*LecturerAnalyticsResponse, error) {
	rows, err := uc.repo.GetLecturerWorkload(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := &LecturerAnalyticsResponse{
		WorkloadDistribution: WorkloadDistribution{
			MinSupervision: -1,
		},
	}
	total := 0
	for _, row := range rows {
		view := LecturerWorkloadView{
			LecturerID:       row.LecturerID,
			FullName:         row.FullName,
			NIDN:             row.NIDN,
			SupervisionCount: row.SupervisionCount,
			SeminarCount:     row.SeminarCount,
			DefenseCount:     row.DefenseCount,
		}
		resp.Lecturers = append(resp.Lecturers, view)
		total += row.SupervisionCount

		if resp.WorkloadDistribution.MaxSupervision < row.SupervisionCount {
			resp.WorkloadDistribution.MaxSupervision = row.SupervisionCount
			resp.WorkloadDistribution.HighestLoad = &LecturerLoadBrief{FullName: row.FullName, SupervisionCount: row.SupervisionCount}
		}
		if resp.WorkloadDistribution.MinSupervision < 0 || resp.WorkloadDistribution.MinSupervision > row.SupervisionCount {
			resp.WorkloadDistribution.MinSupervision = row.SupervisionCount
			resp.WorkloadDistribution.LowestLoad = &LecturerLoadBrief{FullName: row.FullName, SupervisionCount: row.SupervisionCount}
		}
	}
	if len(rows) > 0 {
		resp.WorkloadDistribution.AvgSupervision = float64(total) / float64(len(rows))
	} else {
		resp.WorkloadDistribution.MinSupervision = 0
	}
	return resp, nil
}

// Operational returns pending actions, upcoming schedules, and activity (Job 12).
func (uc *DashboardUseCase) Operational(ctx context.Context, days int) (*OperationalResponse, error) {
	pending, err := uc.repo.GetPendingActions(ctx)
	if err != nil {
		return nil, err
	}
	schedules, err := uc.repo.GetUpcomingSchedules(ctx, days)
	if err != nil {
		return nil, err
	}
	activity, err := uc.repo.GetActivityStats(ctx)
	if err != nil {
		return nil, err
	}
	return &OperationalResponse{
		PendingActions: PendingActionsView{
			PendingTitleReviews:    pending.PendingTitleReviews,
			PendingDocumentReviews: pending.PendingDocumentReviews,
			PendingSeminars:        pending.PendingSeminars,
			PendingDefenses:        pending.PendingDefenses,
		},
		UpcomingSchedules: toUpcomingSchedulesView(schedules),
		ActivityStats: ActivityStatsView{
			LoginsToday:               activity.LoginsToday,
			DocumentsUploadedThisWeek: activity.DocumentsUploadedThisWeek,
			ConsultationsThisWeek:     activity.ConsultationsThisWeek,
		},
	}, nil
}

// Student returns the progress dashboard for one mahasiswa (Job 12).
func (uc *DashboardUseCase) Student(ctx context.Context, studentID uuid.UUID) (*StudentProgressResponse, error) {
	progress, err := uc.repo.GetStudentProgress(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if progress == nil {
		return nil, ErrThesisNotFound
	}

	resp := &StudentProgressResponse{
		ThesisID:           progress.ThesisID,
		Title:              progress.Title,
		Status:             progress.Status,
		CurrentStage:       stageLabelFor(progress.Status),
		ProgressPercentage: StageProgress[progress.Status],
		ConsultationCount:  progress.ConsultationCount,
		LastConsultation:   datePtrToString(progress.LastConsultation),
		UpcomingSeminar:    toScheduleInfoView(progress.UpcomingSeminar),
		UpcomingDefense:    toScheduleInfoView(progress.UpcomingDefense),
	}
	for _, s := range progress.Supervisors {
		resp.Supervisors = append(resp.Supervisors, UserSummaryView{FullName: s.FullName, Email: s.Email, NIM: s.NIM})
	}
	for _, d := range progress.Documents {
		resp.Documents = append(resp.Documents, DocStatusView{Type: d.Type, Status: d.Status, Version: d.Version})
	}
	resp.PendingActions = studentPendingActions(progress)
	return resp, nil
}

// Supervisor returns the supervisor dashboard (Job 12).
func (uc *DashboardUseCase) Supervisor(ctx context.Context, supervisorID uuid.UUID) (*SupervisorDashboardResponse, error) {
	dash, err := uc.repo.GetSupervisorDashboard(ctx, supervisorID)
	if err != nil {
		return nil, err
	}
	resp := &SupervisorDashboardResponse{
		TotalStudents:          dash.TotalStudents,
		PendingDocumentReviews: dash.PendingDocumentReviews,
		UpcomingSchedules:      toUpcomingSchedulesView(&dash.UpcomingSchedules),
	}
	for _, s := range dash.Students {
		resp.Students = append(resp.Students, SupervisedStudentView{
			ThesisID:                  s.ThesisID,
			Student:                   UserSummaryView{FullName: s.Student.FullName, Email: s.Student.Email, NIM: s.Student.NIM},
			Title:                     s.Title,
			Status:                    s.Status,
			CurrentStage:              stageLabelFor(s.Status),
			PendingDocumentReviews:    s.PendingDocumentReviews,
			LastConsultation:          datePtrToString(s.LastConsultation),
			ConsultationCount:         s.ConsultationCount,
			DaysSinceLastConsultation: s.DaysSinceLastConsultation,
		})
	}
	return resp, nil
}

// Examiner returns the examiner dashboard (Job 12).
func (uc *DashboardUseCase) Examiner(ctx context.Context, examinerID uuid.UUID) (*ExaminerDashboardResponse, error) {
	dash, err := uc.repo.GetExaminerDashboard(ctx, examinerID)
	if err != nil {
		return nil, err
	}
	resp := &ExaminerDashboardResponse{}
	for _, a := range dash.UpcomingAssignments {
		resp.UpcomingAssignments = append(resp.UpcomingAssignments, toExaminerAssignmentView(a))
	}
	for _, a := range dash.PendingScores {
		resp.PendingScores = append(resp.PendingScores, toExaminerAssignmentView(a))
	}
	for _, a := range dash.ScoringHistory {
		resp.ScoringHistory = append(resp.ScoringHistory, toExaminerAssignmentView(a))
	}
	return resp, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────

func stageLabelFor(status string) string {
	if l, ok := stageLabel[status]; ok {
		return l
	}
	return status
}

// datePtrToString formats a time as date-only (used for last_consultation).
func datePtrToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// timePtrToRFC3339 formats a time as RFC3339 (used for scheduled_at).
func timePtrToRFC3339(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func toScheduleInfoView(info *domainRepo.ScheduleInfo) *ScheduleInfoView {
	if info == nil {
		return nil
	}
	return &ScheduleInfoView{
		ID:          info.ID,
		ScheduledAt: timePtrToRFC3339(info.ScheduledAt),
		Room:        info.Room,
		Status:      info.Status,
	}
}

func toUpcomingSchedulesView(s *domainRepo.UpcomingSchedules) UpcomingSchedulesView {
	view := UpcomingSchedulesView{}
	for _, item := range s.Seminars {
		view.Seminars = append(view.Seminars, toScheduleItemView(item))
	}
	for _, item := range s.Defenses {
		view.Defenses = append(view.Defenses, toScheduleItemView(item))
	}
	return view
}

func toScheduleItemView(item domainRepo.ScheduleItem) ScheduleItemView {
	return ScheduleItemView{
		ID:          item.ID,
		StudentName: item.StudentName,
		ThesisTitle: item.ThesisTitle,
		ScheduledAt: timePtrToRFC3339(item.ScheduledAt),
		Room:        item.Room,
	}
}

func toExaminerAssignmentView(a domainRepo.ExaminerAssignment) ExaminerAssignmentView {
	return ExaminerAssignmentView{
		Type:        a.Type,
		ID:          a.ID,
		ThesisTitle: a.ThesisTitle,
		StudentName: a.StudentName,
		ScheduledAt: timePtrToRFC3339(a.ScheduledAt),
		Room:        a.Room,
		HasScored:   a.HasScored,
	}
}

// studentPendingActions builds the list of actionable strings for a student
// based on thesis status and the latest document states (Job 12).
func studentPendingActions(p *domainRepo.StudentProgress) []string {
	var actions []string
	switch p.Status {
	case "submitted":
		actions = append(actions, "Menunggu review judul oleh Kaprodi")
	case "rejected":
		actions = append(actions, "Perbaiki judul sesuai catatan Kaprodi lalu ajukan ulang")
	case "approved":
		actions = append(actions, "Menunggu penunjukan dosen pembimbing oleh Kaprodi")
	case "in_progress":
		actions = append(actions, pendingDocActions(p.Documents)...)
		if p.ConsultationCount == 0 {
			actions = append(actions, "Mulai bimbingan dan catat log konsultasi")
		}
	case "seminar_ready":
		actions = append(actions, "Ajukan seminar proposal")
	case "seminar_done":
		actions = append(actions, "Persiapkan dokumen sidang dan ajukan sidang skripsi")
	case "defense_ready":
		actions = append(actions, "Ajukan jadwal sidang skripsi")
	case "defense_done":
		actions = append(actions, "Menunggu penetapan yudisium oleh Kaprodi")
	}
	return actions
}

// pendingDocActions returns document-specific actions from latest states.
func pendingDocActions(docs []domainRepo.DocStatus) []string {
	var actions []string
	for _, d := range docs {
		switch d.Status {
		case "pending_review":
			actions = append(actions, fmt.Sprintf("Menunggu review dokumen %s", d.Type))
		case "revision_required":
			actions = append(actions, fmt.Sprintf("Upload revisi dokumen %s", d.Type))
		}
	}
	return actions
}
