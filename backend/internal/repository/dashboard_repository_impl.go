package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) domainRepo.DashboardRepository {
	return &dashboardRepository{db: db}
}

// applyFilter narrows a theses-based query by the optional dashboard filters.
func (r *dashboardRepository) applyFilter(q *gorm.DB, filter domainRepo.DashboardFilter) *gorm.DB {
	q = q.Where("theses.deleted_at IS NULL")
	if filter.AcademicYearID != nil {
		q = q.Where("theses.academic_year_id = ?", *filter.AcademicYearID)
	}
	if filter.Semester != "" {
		q = q.Where("ay.semester = ?", filter.Semester)
	}
	if filter.StudyProgram != "" {
		q = q.Where("su.study_program = ?", filter.StudyProgram)
	}
	return q
}

func (r *dashboardRepository) GetAcademicSummary(ctx context.Context, filter domainRepo.DashboardFilter) (*domainRepo.AcademicSummary, error) {
	q := r.db.WithContext(ctx).
		Table("theses").
		Joins("JOIN academic_years ay ON ay.id = theses.academic_year_id").
		Joins("JOIN users su ON su.id = theses.student_id").
		Select(`COUNT(*) FILTER (WHERE theses.status NOT IN ('graduated','cancelled')) AS total_active,
				COUNT(*) FILTER (WHERE theses.status = 'graduated') AS total_graduated,
				COALESCE(AVG(EXTRACT(EPOCH FROM (theses.graduated_at - theses.submitted_at)) / 2592000.0), 0) AS avg_completion_months`)

	q = r.applyFilter(q, filter)

	var summary domainRepo.AcademicSummary
	if err := q.Scan(&summary).Error; err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *dashboardRepository) GetThesisByStatus(ctx context.Context, filter domainRepo.DashboardFilter) ([]domainRepo.StatusCount, error) {
	q := r.db.WithContext(ctx).
		Table("theses").
		Joins("JOIN academic_years ay ON ay.id = theses.academic_year_id").
		Joins("JOIN users su ON su.id = theses.student_id").
		Select("theses.status AS status, COUNT(*) AS count").
		Group("theses.status").
		Order("theses.status")

	q = r.applyFilter(q, filter)

	var rows []domainRepo.StatusCount
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *dashboardRepository) GetGraduationTrend(ctx context.Context, filter domainRepo.DashboardFilter) ([]domainRepo.MonthlyCount, error) {
	q := r.db.WithContext(ctx).
		Table("theses").
		Joins("JOIN academic_years ay ON ay.id = theses.academic_year_id").
		Joins("JOIN users su ON su.id = theses.student_id").
		Where("theses.graduated_at IS NOT NULL").
		Select("TO_CHAR(theses.graduated_at, 'YYYY-MM') AS month, COUNT(*) AS count").
		Group("month").
		Order("month DESC").
		Limit(12)

	q = r.applyFilter(q, filter)

	var rows []domainRepo.MonthlyCount
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	// Newest month first reads better on charts as ascending.
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows, nil
}

// buildThesisFilter returns a SQL predicate (with bound args) that narrows a
// theses-scoped aggregation to the dashboard filters. The caller supplies the
// table alias so the same predicate can be reused per join.
func buildThesisFilter(alias string, filter domainRepo.DashboardFilter) (string, []interface{}) {
	var parts []string
	var args []interface{}
	if filter.AcademicYearID != nil {
		parts = append(parts, alias+".academic_year_id = ?")
		args = append(args, *filter.AcademicYearID)
	}
	if filter.Semester != "" {
		parts = append(parts, "EXISTS (SELECT 1 FROM academic_years ay WHERE ay.id = "+alias+".academic_year_id AND ay.semester = ?)")
		args = append(args, filter.Semester)
	}
	if filter.StudyProgram != "" {
		parts = append(parts, "EXISTS (SELECT 1 FROM users su WHERE su.id = "+alias+".student_id AND su.study_program = ?)")
		args = append(args, filter.StudyProgram)
	}
	if len(parts) == 0 {
		return "TRUE", nil
	}
	return strings.Join(parts, " AND "), args
}

func (r *dashboardRepository) GetLecturerWorkload(ctx context.Context, filter domainRepo.DashboardFilter) ([]domainRepo.LecturerWorkload, error) {
	// All lecturers (dosen_pembimbing + dosen_penguji), with left-joined
	// supervision / seminar / defense counts aggregated in SQL. Each count is
	// filtered by the dashboard filters via its own theses alias.
	supExpr, supArgs := buildThesisFilter("sup_t", filter)
	semExpr, semArgs := buildThesisFilter("sem_t", filter)
	defExpr, defArgs := buildThesisFilter("def_t", filter)

	query := fmt.Sprintf(`
		SELECT u.id AS lecturer_id, u.full_name, COALESCE(u.nim_nidn, '') AS nidn,
			COUNT(DISTINCT ts.thesis_id) FILTER (WHERE sup_t.deleted_at IS NULL AND (%s)) AS supervision_count,
			COUNT(DISTINCT se.seminar_id) FILTER (WHERE sem_t.deleted_at IS NULL AND (%s)) AS seminar_count,
			COUNT(DISTINCT de.defense_id) FILTER (WHERE def_t.deleted_at IS NULL AND (%s)) AS defense_count
		FROM users u
		JOIN roles r ON r.id = u.role_id
		LEFT JOIN thesis_supervisors ts ON ts.supervisor_id = u.id
		LEFT JOIN theses sup_t ON sup_t.id = ts.thesis_id
		LEFT JOIN seminar_examiners se ON se.examiner_id = u.id
		LEFT JOIN seminars sem ON sem.id = se.seminar_id
		LEFT JOIN theses sem_t ON sem_t.id = sem.thesis_id
		LEFT JOIN defense_examiners de ON de.examiner_id = u.id
		LEFT JOIN thesis_defenses def ON def.id = de.defense_id
		LEFT JOIN theses def_t ON def_t.id = def.thesis_id
		WHERE r.name IN ('dosen_pembimbing', 'dosen_penguji') AND u.deleted_at IS NULL
		GROUP BY u.id, u.full_name, u.nim_nidn
		ORDER BY supervision_count DESC`, supExpr, semExpr, defExpr)

	args := make([]interface{}, 0, len(supArgs)+len(semArgs)+len(defArgs))
	args = append(args, supArgs...)
	args = append(args, semArgs...)
	args = append(args, defArgs...)

	var rows []domainRepo.LecturerWorkload
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *dashboardRepository) GetPendingActions(ctx context.Context) (*domainRepo.PendingActions, error) {
	var pa domainRepo.PendingActions
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM theses WHERE status = 'submitted' AND deleted_at IS NULL) AS pending_title_reviews,
			(SELECT COUNT(*) FROM documents WHERE status = 'pending_review') AS pending_document_reviews,
			(SELECT COUNT(*) FROM seminars WHERE status = 'pending') AS pending_seminars,
			(SELECT COUNT(*) FROM thesis_defenses WHERE status = 'pending') AS pending_defenses
	`).Scan(&pa).Error; err != nil {
		return nil, err
	}
	return &pa, nil
}

func (r *dashboardRepository) GetUpcomingSchedules(ctx context.Context, days int) (*domainRepo.UpcomingSchedules, error) {
	if days <= 0 {
		days = 7
	}
	end := time.Now().AddDate(0, 0, days)

	// Seminars
	var seminars []domainRepo.ScheduleItem
	err := r.db.WithContext(ctx).Raw(`
		SELECT s.id, u.full_name AS student_name, t.title AS thesis_title, s.scheduled_at, COALESCE(s.room, '') AS room
		FROM seminars s
		JOIN theses t ON t.id = s.thesis_id AND t.deleted_at IS NULL
		JOIN users u ON u.id = t.student_id
		WHERE s.status = 'scheduled' AND s.scheduled_at >= NOW() AND s.scheduled_at <= ?
		ORDER BY s.scheduled_at
	`, end).Scan(&seminars).Error
	if err != nil {
		return nil, err
	}

	// Defenses
	var defenses []domainRepo.ScheduleItem
	err = r.db.WithContext(ctx).Raw(`
		SELECT d.id, u.full_name AS student_name, t.title AS thesis_title, d.scheduled_at, COALESCE(d.room, '') AS room
		FROM thesis_defenses d
		JOIN theses t ON t.id = d.thesis_id AND t.deleted_at IS NULL
		JOIN users u ON u.id = t.student_id
		WHERE d.status = 'scheduled' AND d.scheduled_at >= NOW() AND d.scheduled_at <= ?
		ORDER BY d.scheduled_at
	`, end).Scan(&defenses).Error
	if err != nil {
		return nil, err
	}

	return &domainRepo.UpcomingSchedules{Seminars: seminars, Defenses: defenses}, nil
}

func (r *dashboardRepository) GetActivityStats(ctx context.Context) (*domainRepo.ActivityStats, error) {
	var stats domainRepo.ActivityStats
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM users WHERE last_login_at >= CURRENT_DATE) AS logins_today,
			(SELECT COUNT(*) FROM documents WHERE created_at >= NOW() - INTERVAL '7 days') AS documents_uploaded_this_week,
			(SELECT COUNT(*) FROM consultation_logs WHERE created_at >= NOW() - INTERVAL '7 days') AS consultations_this_week
	`).Scan(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

// derefStr safely dereferences a string pointer for display.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *dashboardRepository) GetStudentProgress(ctx context.Context, studentID uuid.UUID) (*domainRepo.StudentProgress, error) {
	var thesis entity.Thesis
	err := r.db.WithContext(ctx).
		Preload("Supervisors").
		Where("student_id = ? AND deleted_at IS NULL", studentID).
		Order("created_at DESC").
		First(&thesis).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	progress := &domainRepo.StudentProgress{
		ThesisID: thesis.ID,
		Title:    thesis.Title,
		Status:   thesis.Status,
	}
	for _, sup := range thesis.Supervisors {
		progress.Supervisors = append(progress.Supervisors, domainRepo.UserSummary{
			FullName: sup.FullName,
			Email:    sup.Email,
			NIM:      derefStr(sup.NimNidn),
		})
	}

	// Documents (latest version per type)
	var docs []entity.Document
	if err := r.db.WithContext(ctx).
		Where("thesis_id = ?", thesis.ID).
		Order("version DESC").
		Find(&docs).Error; err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, d := range docs {
		key := d.DocumentType
		if d.ChapterNumber != nil {
			key = d.DocumentType + "#" + strconv.Itoa(*d.ChapterNumber)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		progress.Documents = append(progress.Documents, domainRepo.DocStatus{
			Type:    d.DocumentType,
			Status:  d.Status,
			Version: d.Version,
		})
	}

	// Consultation stats
	var consult struct {
		Count int
		Last  *time.Time
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) AS count, MAX(consultation_date) AS last
		FROM consultation_logs WHERE thesis_id = ?
	`, thesis.ID).Scan(&consult).Error; err != nil {
		return nil, err
	}
	progress.ConsultationCount = consult.Count
	progress.LastConsultation = consult.Last

	// Upcoming seminar
	var sem entity.Seminar
	if err := r.db.WithContext(ctx).
		Where("thesis_id = ? AND status = 'scheduled' AND scheduled_at >= NOW()", thesis.ID).
		Order("scheduled_at ASC").
		First(&sem).Error; err == nil {
		progress.UpcomingSeminar = &domainRepo.ScheduleInfo{
			ID:          sem.ID,
			ScheduledAt: sem.ScheduledAt, Room: derefStr(sem.Room),
			Status: sem.Status,
		}
	}

	// Upcoming defense
	var def entity.ThesisDefense
	if err := r.db.WithContext(ctx).
		Where("thesis_id = ? AND status = 'scheduled' AND scheduled_at >= NOW()", thesis.ID).
		Order("scheduled_at ASC").
		First(&def).Error; err == nil {
		progress.UpcomingDefense = &domainRepo.ScheduleInfo{
			ID:          def.ID,
			ScheduledAt: def.ScheduledAt,
			Room:        derefStr(def.Room),
			Status:      def.Status,
		}
	}

	return progress, nil
}

func (r *dashboardRepository) GetSupervisorDashboard(ctx context.Context, supervisorID uuid.UUID) (*domainRepo.SupervisorDashboard, error) {
	type row struct {
		ThesisID          uuid.UUID
		StudentName       string
		StudentEmail      string
		StudentNIM        string
		Title             string
		Status            string
		PendingDocReviews int
		LastConsultation  *time.Time
		ConsultationCount int
		DaysSinceLastCons int
	}
	var rows []row
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			t.id AS thesis_id,
			su.full_name AS student_name,
			su.email AS student_email,
			COALESCE(su.nim_nidn, '') AS student_nim,
			t.title AS title,
			t.status AS status,
			(SELECT COUNT(*) FROM documents d WHERE d.thesis_id = t.id AND d.status = 'pending_review') AS pending_doc_reviews,
			(SELECT MAX(cl.consultation_date) FROM consultation_logs cl WHERE cl.thesis_id = t.id) AS last_consultation,
			(SELECT COUNT(*) FROM consultation_logs cl WHERE cl.thesis_id = t.id) AS consultation_count,
			COALESCE((SELECT EXTRACT(DAY FROM (NOW() - MAX(cl.consultation_date)))::int FROM consultation_logs cl WHERE cl.thesis_id = t.id), 0) AS days_since_last_cons
		FROM theses t
		JOIN users su ON su.id = t.student_id
		WHERE t.deleted_at IS NULL
		  AND t.status NOT IN ('cancelled', 'graduated')
		  AND EXISTS (
				SELECT 1 FROM thesis_supervisors ts
				WHERE ts.thesis_id = t.id AND ts.supervisor_id = ?
			  )
		ORDER BY days_since_last_cons DESC
	`, supervisorID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	dash := &domainRepo.SupervisorDashboard{}
	for _, row := range rows {
		dash.Students = append(dash.Students, domainRepo.SupervisedStudent{
			ThesisID:                  row.ThesisID,
			Student:                   domainRepo.UserSummary{FullName: row.StudentName, Email: row.StudentEmail, NIM: row.StudentNIM},
			Title:                     row.Title,
			Status:                    row.Status,
			PendingDocumentReviews:    row.PendingDocReviews,
			LastConsultation:          row.LastConsultation,
			ConsultationCount:         row.ConsultationCount,
			DaysSinceLastConsultation: row.DaysSinceLastCons,
		})
		dash.TotalStudents++
		dash.PendingDocumentReviews += row.PendingDocReviews
	}

	upcoming, err := r.GetUpcomingSchedules(ctx, 14)
	if err != nil {
		return nil, err
	}
	dash.UpcomingSchedules = *upcoming
	return dash, nil
}

func (r *dashboardRepository) GetExaminerDashboard(ctx context.Context, examinerID uuid.UUID) (*domainRepo.ExaminerDashboard, error) {
	dash := &domainRepo.ExaminerDashboard{}

	type row struct {
		Type        string
		ID          uuid.UUID
		ThesisTitle string
		StudentName string
		ScheduledAt *time.Time
		Room        string
		HasScored   bool
	}

	var rows []row
	if err := r.db.WithContext(ctx).Raw(`
		SELECT 'seminar' AS type, s.id, t.title AS thesis_title, u.full_name AS student_name,
			   s.scheduled_at, COALESCE(s.room, '') AS room,
			   EXISTS(SELECT 1 FROM seminar_scores sc WHERE sc.seminar_id = s.id AND sc.examiner_id = ?) AS has_scored
		FROM seminar_examiners se
		JOIN seminars s ON s.id = se.seminar_id
		JOIN theses t ON t.id = s.thesis_id
		JOIN users u ON u.id = t.student_id
		WHERE se.examiner_id = ? AND s.status IN ('scheduled','passed','failed')
		UNION ALL
		SELECT 'defense' AS type, d.id, t.title AS thesis_title, u.full_name AS student_name,
			   d.scheduled_at, COALESCE(d.room, '') AS room,
			   EXISTS(SELECT 1 FROM defense_scores sc WHERE sc.defense_id = d.id AND sc.examiner_id = ?) AS has_scored
		FROM defense_examiners de
		JOIN thesis_defenses d ON d.id = de.defense_id
		JOIN theses t ON t.id = d.thesis_id
		JOIN users u ON u.id = t.student_id
		WHERE de.examiner_id = ? AND d.status IN ('scheduled','passed','failed','revision_required')
		ORDER BY scheduled_at DESC
	`, examinerID, examinerID, examinerID, examinerID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	for _, row := range rows {
		item := domainRepo.ExaminerAssignment{
			Type:        row.Type,
			ID:          row.ID,
			ThesisTitle: row.ThesisTitle,
			StudentName: row.StudentName,
			ScheduledAt: row.ScheduledAt,
			Room:        row.Room,
			HasScored:   row.HasScored,
		}
		switch {
		case row.ScheduledAt != nil && row.ScheduledAt.After(now):
			dash.UpcomingAssignments = append(dash.UpcomingAssignments, item)
		case !row.HasScored:
			dash.PendingScores = append(dash.PendingScores, item)
		default:
			dash.ScoringHistory = append(dash.ScoringHistory, item)
		}
	}
	return dash, nil
}
