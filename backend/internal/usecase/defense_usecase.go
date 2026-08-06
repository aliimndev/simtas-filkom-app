package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
)

// Defense status values (Job 09). Kept in sync with the DB check constraint.
const (
	DefenseStatusPending          = "pending"
	DefenseStatusScheduled        = "scheduled"
	DefenseStatusPassed           = "passed"
	DefenseStatusFailed           = "failed"
	DefenseStatusRevisionRequired = "revision_required"
)

// Defense grading thresholds (Job 09).
const (
	DefenseFailThreshold     = 60.0 // < 60 → failed
	DefenseRevisionThreshold = 75.0 // 60–74 → passed with revision required
)

// MinDefenseLeadDays is the minimum scheduling lead time (stricter than seminar).
const MinDefenseLeadDays = 7

// MinDefenseExaminers is the minimum number of examiners (Job 09).
const MinDefenseExaminers = 2

// UpcomingScheduleWindowDays is the horizon for the upcoming-schedules endpoint.
const UpcomingScheduleWindowDays = 14

var (
	ErrDefenseNotFound         = errors.New("sidang tidak ditemukan")
	ErrDefenseActiveExists     = errors.New("thesis ini sudah memiliki sidang aktif")
	ErrDefenseGateNotMet       = errors.New("syarat sidang belum terpenuhi: seminar harus lulus dan dokumen sidang harus disetujui")
	ErrDefenseScheduleLeadTime = errors.New("jadwal minimal 7 hari dari sekarang")
	ErrDefenseMinExaminers     = errors.New("minimal 2 penguji")
	ErrDefenseInvalidExaminer  = errors.New("penguji harus dosen_penguji yang aktif")
	ErrDefenseRoomConflict     = errors.New("ruangan atau penguji bentrok dengan jadwal lain")
	ErrDefenseNotScheduled     = errors.New("sidang harus berstatus scheduled")
	ErrDefenseNotExaminer      = errors.New("hanya penguji yang ditugaskan yang dapat menilai")
	ErrDefenseAlreadyScored    = errors.New("nilai untuk penguji ini sudah disubmit")
	ErrDefenseIncompleteScore  = errors.New("semua komponen penilaian wajib diisi dengan nilai 0-100")
	ErrDefenseInvalidScore     = errors.New("nilai harus antara 0 dan 100")
	ErrDefenseNotFinalized     = errors.New("catatan revisi hanya dapat ditambahkan setelah sidang dinilai")
	ErrGraduationGateNotMet    = errors.New("syarat yudisium belum terpenuhi: sidang harus lulus dan dokumen final harus disetujui")
)

// ScheduleDefenseRequest is the payload for PUT /defenses/:id/schedule.
type ScheduleDefenseRequest struct {
	ScheduledAt time.Time   `json:"scheduled_at" binding:"required"`
	Room        string      `json:"room" binding:"required"`
	ExaminerIDs []uuid.UUID `json:"examiner_ids" binding:"required"`
}

// SubmitDefenseScoreRequest is the payload for POST /defenses/:id/scores.
type SubmitDefenseScoreRequest struct {
	Scores []ComponentScoreInput `json:"scores" binding:"required"`
}

// DefenseDetail is the API response shape for a defense (Job 09).
type DefenseDetail struct {
	ID            uuid.UUID             `json:"id"`
	Thesis        *SeminarThesisInfo    `json:"thesis,omitempty"`
	Status        string                `json:"status"`
	ScheduledAt   *time.Time            `json:"scheduled_at,omitempty"`
	Room          *string               `json:"room,omitempty"`
	Examiners     []*UserBrief          `json:"examiners,omitempty"`
	Scores        []*DefenseScoreDetail `json:"scores,omitempty"`
	FinalScore    *float64              `json:"final_score,omitempty"`
	RevisionNotes *string               `json:"revision_notes,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
}

// DefenseScoreDetail is one examiner's score row in the API response.
type DefenseScoreDetail struct {
	Examiner        *UserBrief `json:"examiner,omitempty"`
	ComponentName   string     `json:"component_name"`
	ComponentWeight float64    `json:"component_weight"`
	Score           float64    `json:"score"`
}

// DefenseResult is the result view (GET /defenses/:id/result).
type DefenseResult struct {
	DefenseID         uuid.UUID                `json:"defense_id"`
	FinalScore        *float64                 `json:"final_score,omitempty"`
	Status            string                   `json:"status"`
	GradeCategory     string                   `json:"grade_category"`
	IsComplete        bool                     `json:"is_complete"`
	ExaminerScores    []ExaminerScoreBreakdown `json:"examiner_scores,omitempty"`
	GradingComponents []GradingComponentView   `json:"grading_components"`
}

// GraduationRequest is the payload for PUT /theses/:thesis_id/graduation.
type GraduationRequest struct {
	Notes string `json:"notes"`
}

// UpcomingSchedules bundles seminars and defenses in the next N days (Job 09).
type UpcomingSchedules struct {
	Seminars []UpcomingScheduleItem `json:"seminars"`
	Defenses []UpcomingScheduleItem `json:"defenses"`
}

// UpcomingScheduleItem is one scheduled event for the upcoming-schedules view.
type UpcomingScheduleItem struct {
	ID          uuid.UUID `json:"id"`
	StudentName string    `json:"student_name"`
	ThesisTitle string    `json:"thesis_title"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Room        string    `json:"room"`
}

// DefenseUseCase contains business logic for the defense module (Job 09).
type DefenseUseCase struct {
	defenseRepo repository.DefenseRepository
	seminarRepo repository.SeminarRepository
	thesisRepo  repository.ThesisRepository
	userRepo    repository.UserRepository
	documentUC  *DocumentUseCase
	access      *ThesisAccess
	emailSvc    email.EmailService
	auditSvc    *audit.AuditService
}

func NewDefenseUseCase(
	defenseRepo repository.DefenseRepository,
	seminarRepo repository.SeminarRepository,
	thesisRepo repository.ThesisRepository,
	userRepo repository.UserRepository,
	documentUC *DocumentUseCase,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
) *DefenseUseCase {
	return &DefenseUseCase{
		defenseRepo: defenseRepo,
		seminarRepo: seminarRepo,
		thesisRepo:  thesisRepo,
		userRepo:    userRepo,
		documentUC:  documentUC,
		access:      NewThesisAccess(thesisRepo),
		emailSvc:    emailSvc,
		auditSvc:    auditSvc,
	}
}

// Submit creates a defense request for a thesis (Mahasiswa pemilik only).
func (uc *DefenseUseCase) Submit(ctx context.Context, thesisID uuid.UUID, actor Actor) (*DefenseDetail, error) {
	thesis, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}
	if !isThesisOwner(thesis, actor.UserID) {
		return nil, ErrForbidden
	}

	// No active defense (status not failed) for this thesis.
	existing, err := uc.defenseRepo.FindByThesisID(ctx, thesisID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil && existing.Status != DefenseStatusFailed {
		return nil, ErrDefenseActiveExists
	}

	// Gate: thesis must be seminar_done + defense_doc approved.
	canSubmit, _, err := uc.CanSubmitDefense(ctx, thesisID)
	if err != nil {
		return nil, err
	}
	if !canSubmit {
		return nil, ErrDefenseGateNotMet
	}

	defense := &entity.ThesisDefense{
		ThesisID: thesisID,
		Status:   DefenseStatusPending,
	}
	if err := uc.defenseRepo.Create(ctx, defense); err != nil {
		return nil, err
	}

	// Update thesis status → defense_ready.
	if err := uc.thesisRepo.UpdateStatus(ctx, thesisID, "defense_ready", ""); err != nil {
		return nil, err
	}

	// Notify kaprodi + admin (async, non-fatal). The closure must capture a
	// local copy: `defense` is reassigned below and goroutines capture by reference.
	submitted := defense
	go func() {
		emails := collectRoleEmails(context.Background(), uc.userRepo, ThesisRoleKaprodi)
		adminEmails, _ := uc.userRepo.FindByRole(context.Background(), ThesisRoleAdminFakultas)
		for _, a := range adminEmails {
			emails = append(emails, a.Email)
		}
		if len(emails) > 0 {
			_ = uc.emailSvc.SendDefenseSubmitted(context.Background(), emails, submitted)
		}
	}()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionDefenseSubmitted,
		EntityType: "defense",
		EntityID:   &defense.ID,
		NewValue:   map[string]interface{}{"thesis_id": thesisID, "status": DefenseStatusPending},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	defense, err = uc.defenseRepo.FindByID(ctx, defense.ID)
	if err != nil {
		return nil, err
	}
	return toDefenseDetail(defense), nil
}

// List returns defenses scoped by role (Admin/Kaprodi: all; Penguji: assigned;
// Mahasiswa/Pembimbing: related to their theses).
func (uc *DefenseUseCase) List(ctx context.Context, filter repository.DefenseFilter, userID uuid.UUID, role string) ([]*DefenseDetail, int64, error) {
	switch role {
	case ThesisRoleMahasiswa:
		filter.StudentID = userID
	case ThesisRoleDosenPembimbing:
		filter.SupervisorID = userID
	case ThesisRoleDosenPenguji:
		filter.ExaminerID = userID
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		// all defenses
	}

	defenses, total, err := uc.defenseRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	details := make([]*DefenseDetail, 0, len(defenses))
	for _, d := range defenses {
		details = append(details, toDefenseDetail(d))
	}
	return details, total, nil
}

// GetByID returns a single defense (same access as List).
func (uc *DefenseUseCase) GetByID(ctx context.Context, id, userID uuid.UUID, role string) (*DefenseDetail, error) {
	defense, err := uc.defenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDefenseNotFound
		}
		return nil, err
	}
	ok, err := uc.canAccessDefense(ctx, defense, userID, role)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return toDefenseDetail(defense), nil
}

// Schedule schedules or reschedules a defense (Admin + Kaprodi only).
func (uc *DefenseUseCase) Schedule(ctx context.Context, id uuid.UUID, req ScheduleDefenseRequest, actor Actor) (*DefenseDetail, error) {
	defense, err := uc.defenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDefenseNotFound
		}
		return nil, err
	}
	if defense.Status != DefenseStatusPending && defense.Status != DefenseStatusScheduled {
		return nil, ErrDefenseNotScheduled
	}

	if time.Until(req.ScheduledAt) < MinDefenseLeadDays*24*time.Hour {
		return nil, ErrDefenseScheduleLeadTime
	}
	if len(req.ExaminerIDs) < MinDefenseExaminers {
		return nil, ErrDefenseMinExaminers
	}

	// Every examiner must be an active dosen_penguji (dedupe while validating).
	examiners, err := uc.userRepo.FindByRole(ctx, ThesisRoleDosenPenguji)
	if err != nil {
		return nil, err
	}
	valid := make(map[uuid.UUID]*entity.User, len(examiners))
	for _, e := range examiners {
		valid[e.ID] = e
	}
	seen := map[uuid.UUID]bool{}
	examinerEmails := make([]string, 0, len(req.ExaminerIDs))
	examinerIDs := make([]uuid.UUID, 0, len(req.ExaminerIDs))
	for _, eid := range req.ExaminerIDs {
		if seen[eid] {
			continue
		}
		u, ok := valid[eid]
		if !ok {
			return nil, ErrDefenseInvalidExaminer
		}
		seen[eid] = true
		examinerIDs = append(examinerIDs, eid)
		examinerEmails = append(examinerEmails, u.Email)
	}
	if len(seen) < MinDefenseExaminers {
		return nil, ErrDefenseMinExaminers
	}

	// Check room + examiner conflicts.
	excludeID := &id
	conflict, err := uc.defenseRepo.CheckScheduleConflict(ctx, req.Room, req.ScheduledAt, examinerIDs, excludeID)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, ErrDefenseRoomConflict
	}

	// Replace examiners.
	if err := uc.defenseRepo.RemoveAllExaminers(ctx, id); err != nil {
		return nil, err
	}
	for _, eid := range examinerIDs {
		if err := uc.defenseRepo.AssignExaminer(ctx, id, eid, actor.UserID); err != nil {
			return nil, err
		}
	}

	if err := uc.defenseRepo.UpdateSchedule(ctx, id, req.ScheduledAt, req.Room); err != nil {
		return nil, err
	}
	wasScheduled := defense.Status == DefenseStatusScheduled
	if err := uc.defenseRepo.UpdateStatus(ctx, id, DefenseStatusScheduled); err != nil {
		return nil, err
	}

	// Keep the in-memory record in sync so the notification goroutine below
	// observes the new schedule (the repo updates only touch the database).
	defense.ScheduledAt = &req.ScheduledAt
	defense.Room = &req.Room
	defense.Status = DefenseStatusScheduled

	// Notify student, supervisors, examiners (async, non-fatal).
	scheduled := defense
	go func() {
		recipients := defenseRecipients(scheduled, examinerEmails)
		if len(recipients) > 0 {
			_ = uc.emailSvc.SendDefenseScheduled(context.Background(), recipients, scheduled)
		}
	}()

	action := audit.ActionDefenseScheduled
	if wasScheduled {
		action = audit.ActionDefenseRescheduled
	}
	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     action,
		EntityType: "defense",
		EntityID:   &defense.ID,
		NewValue: map[string]interface{}{
			"scheduled_at": req.ScheduledAt,
			"room":         req.Room,
			"examiner_ids": examinerIDs,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})

	defense, err = uc.defenseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDefenseDetail(defense), nil
}

// SubmitScores records an examiner's scores and triggers finalization when all
// examiners have submitted (Dosen Penguji yang ditugaskan only).
func (uc *DefenseUseCase) SubmitScores(ctx context.Context, id uuid.UUID, req SubmitDefenseScoreRequest, actor Actor) (*DefenseResult, error) {
	defense, err := uc.defenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDefenseNotFound
		}
		return nil, err
	}
	if defense.Status != DefenseStatusScheduled {
		return nil, ErrDefenseNotScheduled
	}

	examiners, err := uc.defenseRepo.GetExaminers(ctx, id)
	if err != nil {
		return nil, err
	}
	assigned := false
	for _, e := range examiners {
		if e.ID == actor.UserID {
			assigned = true
			break
		}
	}
	if !assigned {
		return nil, ErrDefenseNotExaminer
	}

	already, err := uc.defenseRepo.HasExaminerScored(ctx, id, actor.UserID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, ErrDefenseAlreadyScored
	}

	if err := validateScoreComponents(entity.DefenseGradingComponents, req.Scores, ErrDefenseIncompleteScore, ErrDefenseInvalidScore); err != nil {
		return nil, err
	}

	for _, c := range req.Scores {
		weight := 0.0
		for _, g := range entity.DefenseGradingComponents {
			if g.Name == c.ComponentName {
				weight = g.Weight
				break
			}
		}
		score := &entity.DefenseScore{
			DefenseID:       id,
			ExaminerID:      actor.UserID,
			ComponentName:   c.ComponentName,
			ComponentWeight: weight,
			Score:           c.Score,
		}
		if err := uc.defenseRepo.AddScore(ctx, score); err != nil {
			return nil, err
		}
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionDefenseScoreSubmit,
		EntityType: "defense",
		EntityID:   &id,
		NewValue:   map[string]interface{}{"examiner_id": actor.UserID},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	if err := uc.TryFinalizeDefense(ctx, id); err != nil {
		return nil, err
	}
	return uc.Result(ctx, id, actor.UserID, ThesisRoleDosenPenguji)
}

// Result returns the full breakdown for a defense (all related parties).
func (uc *DefenseUseCase) Result(ctx context.Context, id, userID uuid.UUID, role string) (*DefenseResult, error) {
	defense, err := uc.defenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDefenseNotFound
		}
		return nil, err
	}
	ok, err := uc.canAccessDefense(ctx, defense, userID, role)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}

	scores, err := uc.defenseRepo.GetAllScores(ctx, id)
	if err != nil {
		return nil, err
	}
	examiners, err := uc.defenseRepo.GetExaminers(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &DefenseResult{
		DefenseID:         id,
		Status:            defense.Status,
		FinalScore:        defense.FinalScore,
		GradeCategory:     GetGradeCategory(defense.FinalScore),
		GradingComponents: toGradingComponentViews(entity.DefenseGradingComponents),
	}

	byExaminer := map[uuid.UUID][]*entity.DefenseScore{}
	order := []uuid.UUID{}
	for _, s := range scores {
		if _, ok := byExaminer[s.ExaminerID]; !ok {
			order = append(order, s.ExaminerID)
		}
		byExaminer[s.ExaminerID] = append(byExaminer[s.ExaminerID], s)
	}

	for _, eid := range order {
		rows := byExaminer[eid]
		examiner := findUserByID(examiners, eid)
		var examinerScore float64
		components := make([]ComponentScoreView, 0, len(rows))
		for _, r := range rows {
			examinerScore += r.Score * r.ComponentWeight / 100.0
			components = append(components, ComponentScoreView{
				Name:     r.ComponentName,
				Weight:   r.ComponentWeight,
				Score:    r.Score,
				Weighted: r.Score * r.ComponentWeight / 100.0,
			})
		}
		result.ExaminerScores = append(result.ExaminerScores, ExaminerScoreBreakdown{
			Examiner:      examiner,
			ExaminerScore: examinerScore,
			Components:    components,
		})
	}

	scoredCount := len(order)
	result.IsComplete = scoredCount >= len(examiners) && len(examiners) > 0
	return result, nil
}

// TryFinalizeDefense finalizes a defense once all examiners have submitted.
// The score aggregation and defense status update happen atomically inside the
// repository (row-locked transaction), so concurrent submissions cannot both
// finalize (the second observer sees status != "scheduled" and no-ops). The
// thesis status update is coordinated here across repositories afterwards.
func (uc *DefenseUseCase) TryFinalizeDefense(ctx context.Context, defenseID uuid.UUID) error {
	finalScore, status, thesisID, err := uc.defenseRepo.FinalizeDefense(ctx, defenseID)
	if err != nil {
		return err
	}
	// status == "" means not yet complete or already finalized → nothing to do.
	if status == "" {
		return nil
	}

	// Move the thesis forward (or back to defense_ready on failure).
	thesisStatus := "defense_done"
	if status == DefenseStatusFailed {
		thesisStatus = "defense_ready"
	}
	if err := uc.thesisRepo.UpdateStatus(ctx, thesisID, thesisStatus, ""); err != nil {
		return err
	}

	if thesisID != uuid.Nil {
		thesis, terr := uc.thesisRepo.FindByID(ctx, thesisID)
		if terr == nil {
			go func() {
				_ = uc.emailSvc.SendDefenseFinalized(context.Background(), thesis.Student.Email, &entity.ThesisDefense{
					ID:         defenseID,
					ThesisID:   thesisID,
					Status:     status,
					FinalScore: &finalScore,
				})
			}()
		}
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     nil,
		Action:     audit.ActionDefenseFinalized,
		EntityType: "defense",
		EntityID:   &defenseID,
		NewValue: map[string]interface{}{
			"final_score": finalScore,
			"status":      status,
		},
	})
	return nil
}

// SetRevisionNotes records post-defense revision notes (Admin + Kaprodi only).
func (uc *DefenseUseCase) SetRevisionNotes(ctx context.Context, id uuid.UUID, notes string, actor Actor) (*DefenseDetail, error) {
	defense, err := uc.defenseRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDefenseNotFound
		}
		return nil, err
	}
	// Only after the defense has been scored (passed or revision_required).
	if defense.Status != DefenseStatusPassed && defense.Status != DefenseStatusRevisionRequired {
		return nil, ErrDefenseNotFinalized
	}

	if err := uc.defenseRepo.SetRevisionNotes(ctx, id, notes); err != nil {
		return nil, err
	}
	defense.RevisionNotes = &notes
	// Mark revision_required if it isn't already.
	if defense.Status == DefenseStatusPassed {
		if err := uc.defenseRepo.UpdateStatus(ctx, id, DefenseStatusRevisionRequired); err != nil {
			return nil, err
		}
		defense.Status = DefenseStatusRevisionRequired
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionDefenseRevisionNote,
		EntityType: "defense",
		EntityID:   &defense.ID,
		NewValue:   map[string]interface{}{"revision_notes": notes},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	// Notify the student (async, non-fatal).
	revisioned := defense
	thesis, err := uc.thesisRepo.FindByID(ctx, defense.ThesisID)
	if err == nil {
		go func() { _ = uc.emailSvc.SendDefenseFinalized(context.Background(), thesis.Student.Email, revisioned) }()
	}

	defense, err = uc.defenseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDefenseDetail(defense), nil
}

// Graduate marks a thesis as graduated (Kaprodi only).
func (uc *DefenseUseCase) Graduate(ctx context.Context, thesisID uuid.UUID, req GraduationRequest, actor Actor) error {
	thesis, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrThesisNotFound
		}
		return err
	}
	if thesis.Status != "defense_done" {
		return ErrGraduationGateNotMet
	}

	defense, err := uc.defenseRepo.FindByThesisID(ctx, thesisID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// Defense must be passed or revision_required (revisions complete).
	if defense == nil ||
		(defense.Status != DefenseStatusPassed && defense.Status != DefenseStatusRevisionRequired) {
		return ErrGraduationGateNotMet
	}

	// final_thesis must be approved.
	finalOK, err := uc.documentUC.IsFinalThesisApproved(ctx, thesisID)
	if err != nil {
		return err
	}
	if !finalOK {
		return ErrGraduationGateNotMet
	}

	// Mutate the fetched thesis (full record) — the repository Update uses
	// GORM Save, which writes every field including zero values.
	now := time.Now()
	thesis.Status = "graduated"
	thesis.GraduatedAt = &now
	if err := uc.thesisRepo.Update(ctx, thesis); err != nil {
		return err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionThesisGraduated,
		EntityType: "thesis",
		EntityID:   &thesisID,
		NewValue:   map[string]interface{}{"status": "graduated", "notes": req.Notes},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	// Notify the student (async, non-fatal).
	go func() { _ = uc.emailSvc.SendGraduated(context.Background(), thesis.Student.Email, thesis) }()
	return nil
}

// Upcoming returns seminars + defenses scheduled within the next N days
// (Admin + Kaprodi, for the operational dashboard).
func (uc *DefenseUseCase) Upcoming(ctx context.Context, days int) (*UpcomingSchedules, error) {
	if days <= 0 {
		days = UpcomingScheduleWindowDays
	}
	now := time.Now()
	from := now.Add(-2 * time.Hour) // include events already in progress
	to := now.Add(time.Duration(days) * 24 * time.Hour)

	seminarFilter := repository.SeminarFilter{
		Status:   SeminarStatusScheduled,
		DateFrom: &from,
		DateTo:   &to,
	}
	seminars, _, err := uc.seminarRepo.FindAll(ctx, seminarFilter)
	if err != nil {
		return nil, err
	}

	defenseFilter := repository.DefenseFilter{
		Status:   DefenseStatusScheduled,
		DateFrom: &from,
		DateTo:   &to,
	}
	defenses, _, err := uc.defenseRepo.FindAll(ctx, defenseFilter)
	if err != nil {
		return nil, err
	}

	result := &UpcomingSchedules{
		Seminars: make([]UpcomingScheduleItem, 0, len(seminars)),
		Defenses: make([]UpcomingScheduleItem, 0, len(defenses)),
	}
	for _, s := range seminars {
		result.Seminars = append(result.Seminars, toUpcomingScheduleItem(s.ID, s.Thesis, s.ScheduledAt, s.Room))
	}
	for _, d := range defenses {
		result.Defenses = append(result.Defenses, toUpcomingScheduleItem(d.ID, d.Thesis, d.ScheduledAt, d.Room))
	}
	return result, nil
}

// CanSubmitDefense implements the defense gate (Job 09):
//
//	Cek 1: thesis.status harus "seminar_done"
//	Cek 2: dokumen "defense_doc" harus berstatus "approved"
//
// It returns (canSubmit, reason, error); reason is "seminar" or "defense_doc".
func (uc *DefenseUseCase) CanSubmitDefense(ctx context.Context, thesisID uuid.UUID) (bool, string, error) {
	thesis, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", ErrThesisNotFound
		}
		return false, "", err
	}
	if thesis.Status != "seminar_done" {
		return false, "seminar", nil
	}
	docOK, err := uc.documentUC.CanSubmitDefense(ctx, thesisID)
	if err != nil {
		return false, "", err
	}
	if !docOK {
		return false, "defense_doc", nil
	}
	return true, "", nil
}

// canAccessDefense enforces per-role access to a single defense.
func (uc *DefenseUseCase) canAccessDefense(ctx context.Context, defense *entity.ThesisDefense, userID uuid.UUID, role string) (bool, error) {
	switch role {
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		return true, nil
	case ThesisRoleMahasiswa:
		return uc.access.IsThesisOwner(ctx, userID, defense.ThesisID)
	case ThesisRoleDosenPembimbing:
		return uc.access.IsSupervisor(ctx, userID, defense.ThesisID)
	case ThesisRoleDosenPenguji:
		examiners, err := uc.defenseRepo.GetExaminers(ctx, defense.ID)
		if err != nil {
			return false, err
		}
		for _, e := range examiners {
			if e.ID == userID {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

// GetGradeCategory maps a final score to a letter grade (Job 09 spec).
func GetGradeCategory(score *float64) string {
	if score == nil {
		return "-"
	}
	switch {
	case *score >= 85:
		return "A"
	case *score >= 75:
		return "B+"
	case *score >= 70:
		return "B"
	case *score >= 60:
		return "C"
	default:
		return "Tidak Lulus"
	}
}

// toDefenseDetail maps a persisted defense into the API response shape.
func toDefenseDetail(d *entity.ThesisDefense) *DefenseDetail {
	detail := &DefenseDetail{
		ID:            d.ID,
		Status:        d.Status,
		ScheduledAt:   d.ScheduledAt,
		Room:          d.Room,
		FinalScore:    d.FinalScore,
		RevisionNotes: d.RevisionNotes,
		CreatedAt:     d.CreatedAt,
	}
	if d.Thesis.ID != uuid.Nil {
		detail.Thesis = &SeminarThesisInfo{ID: d.Thesis.ID, Title: d.Thesis.Title}
		if d.Thesis.Student.ID != uuid.Nil {
			detail.Thesis.Student = &ThesisStudentBrief{
				FullName: d.Thesis.Student.FullName,
				Nim:      d.Thesis.Student.NimNidn,
			}
		}
	}
	for _, e := range d.Examiners {
		if e.ID != uuid.Nil {
			detail.Examiners = append(detail.Examiners, &UserBrief{ID: e.ID, FullName: e.FullName})
		}
	}
	for _, sc := range d.Scores {
		row := &DefenseScoreDetail{
			ComponentName:   sc.ComponentName,
			ComponentWeight: sc.ComponentWeight,
			Score:           sc.Score,
		}
		if sc.Examiner.ID != uuid.Nil {
			row.Examiner = &UserBrief{ID: sc.Examiner.ID, FullName: sc.Examiner.FullName}
		}
		detail.Scores = append(detail.Scores, row)
	}
	return detail
}

// defenseRecipients collects student + supervisor emails for a defense.
func defenseRecipients(defense *entity.ThesisDefense, examinerEmails []string) []string {
	recipients := []string{}
	if defense.Thesis.Student.Email != "" {
		recipients = append(recipients, defense.Thesis.Student.Email)
	}
	for _, s := range defense.Thesis.Supervisors {
		if s.Email != "" {
			recipients = append(recipients, s.Email)
		}
	}
	recipients = append(recipients, examinerEmails...)
	return recipients
}

// toUpcomingScheduleItem builds an upcoming-schedule item.
func toUpcomingScheduleItem(id uuid.UUID, thesis entity.Thesis, scheduledAt *time.Time, room *string) UpcomingScheduleItem {
	item := UpcomingScheduleItem{ID: id, ThesisTitle: thesis.Title}
	if scheduledAt != nil {
		item.ScheduledAt = *scheduledAt
	}
	if room != nil {
		item.Room = *room
	}
	if thesis.Student.FullName != "" {
		item.StudentName = thesis.Student.FullName
	}
	return item
}
