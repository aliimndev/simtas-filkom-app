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
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/grading"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/notification"
)

// Seminar status values (Job 08). Kept in sync with the DB check constraint.
const (
	SeminarStatusPending   = "pending"
	SeminarStatusScheduled = "scheduled"
	SeminarStatusPassed    = "passed"
	SeminarStatusFailed    = "failed"
)

// SeminarPassScore is the minimum final score to pass the seminar.
const SeminarPassScore = 60.0

// MinSeminarLeadDays is the minimum scheduling lead time (Job 08).
const MinSeminarLeadDays = 3

// MinSeminarExaminers is the minimum number of examiners (Job 08).
const MinSeminarExaminers = 2

var (
	ErrSeminarNotFound         = errors.New("seminar tidak ditemukan")
	ErrSeminarActiveExists     = errors.New("thesis ini sudah memiliki seminar aktif")
	ErrSeminarGateNotMet       = errors.New("dokumen seminar belum disetujui atau thesis tidak berstatus in_progress")
	ErrSeminarInvalidDecision  = errors.New("decision harus approved atau rejected")
	ErrSeminarScheduleLeadTime = errors.New("jadwal minimal 3 hari dari sekarang")
	ErrSeminarMinExaminers     = errors.New("minimal 2 penguji")
	ErrSeminarInvalidExaminer  = errors.New("penguji harus dosen_penguji yang aktif")
	ErrSeminarRoomConflict     = errors.New("ruangan atau penguji bentrok dengan jadwal lain")
	ErrSeminarNotScheduled     = errors.New("seminar harus berstatus scheduled")
	ErrSeminarNotExaminer      = errors.New("hanya penguji yang ditugaskan yang dapat menilai")
	ErrSeminarAlreadyScored    = errors.New("nilai untuk penguji ini sudah disubmit")
	ErrSeminarIncompleteScore  = errors.New("semua komponen penilaian wajib diisi dengan nilai 0-100")
	ErrSeminarNotPassed        = errors.New("catatan revisi hanya dapat ditambahkan setelah seminar lulus")
	ErrSeminarInvalidScore     = errors.New("nilai harus antara 0 dan 100")
)

// ScheduleSeminarRequest is the payload for PUT /seminars/:id/schedule.
type ScheduleSeminarRequest struct {
	ScheduledAt time.Time   `json:"scheduled_at" binding:"required"`
	Room        string      `json:"room" binding:"required"`
	ExaminerIDs []uuid.UUID `json:"examiner_ids" binding:"required"`
}

// SubmitScoreRequest is the payload for POST /seminars/:id/scores.
type SubmitScoreRequest struct {
	Scores []ComponentScoreInput `json:"scores" binding:"required"`
}

// ComponentScoreInput is one examiner's score for a single component.
type ComponentScoreInput struct {
	ComponentName string  `json:"component_name" binding:"required"`
	Score         float64 `json:"score" binding:"required"`
}

// SeminarDetail is the API response shape for a seminar (Job 08).
type SeminarDetail struct {
	ID          uuid.UUID             `json:"id"`
	Thesis      *SeminarThesisInfo    `json:"thesis"`
	Status      string                `json:"status"`
	ScheduledAt *time.Time            `json:"scheduled_at,omitempty"`
	Room        *string               `json:"room,omitempty"`
	Examiners   []*UserBrief          `json:"examiners,omitempty"`
	Scores      []*SeminarScoreDetail `json:"scores,omitempty"`
	FinalScore  *float64              `json:"final_score,omitempty"`
	Notes       *string               `json:"notes,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
}

// SeminarThesisInfo summarizes the thesis linked to a seminar.
type SeminarThesisInfo struct {
	ID      uuid.UUID           `json:"id"`
	Title   string              `json:"title"`
	Student *ThesisStudentBrief `json:"student,omitempty"`
}

// ThesisStudentBrief is a compact student representation.
type ThesisStudentBrief struct {
	FullName string  `json:"full_name"`
	Nim      *string `json:"nim,omitempty"`
}

// SeminarScoreDetail is one examiner's score row in the API response.
type SeminarScoreDetail struct {
	Examiner        *UserBrief `json:"examiner,omitempty"`
	ComponentName   string     `json:"component_name"`
	ComponentWeight float64    `json:"component_weight"`
	Score           float64    `json:"score"`
}

// SeminarResult is the result view (GET /seminars/:id/result).
type SeminarResult struct {
	SeminarID         uuid.UUID                `json:"seminar_id"`
	FinalScore        *float64                 `json:"final_score,omitempty"`
	Status            string                   `json:"status"`
	IsComplete        bool                     `json:"is_complete"`
	ExaminerScores    []ExaminerScoreBreakdown `json:"examiner_scores,omitempty"`
	GradingComponents []GradingComponentView   `json:"grading_components"`
}

// ExaminerScoreBreakdown groups one examiner's scores by component.
type ExaminerScoreBreakdown struct {
	Examiner      *UserBrief           `json:"examiner"`
	ExaminerScore float64              `json:"examiner_score"`
	Components    []ComponentScoreView `json:"components"`
}

// ComponentScoreView is one weighted component in the result breakdown.
type ComponentScoreView struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Score    float64 `json:"score"`
	Weighted float64 `json:"weighted"`
}

// GradingComponentView exposes the fixed grading components.
type GradingComponentView struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

// SeminarUseCase contains business logic for the seminar module (Job 08).
type SeminarUseCase struct {
	seminarRepo repository.SeminarRepository
	thesisRepo  repository.ThesisRepository
	userRepo    repository.UserRepository
	documentUC  *DocumentUseCase
	access      *ThesisAccess
	emailSvc    email.EmailService
	auditSvc    *audit.AuditService
	notifSvc    *notification.NotificationService
}

func NewSeminarUseCase(
	seminarRepo repository.SeminarRepository,
	thesisRepo repository.ThesisRepository,
	userRepo repository.UserRepository,
	documentUC *DocumentUseCase,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
	notifSvc *notification.NotificationService,
) *SeminarUseCase {
	return &SeminarUseCase{
		seminarRepo: seminarRepo,
		thesisRepo:  thesisRepo,
		userRepo:    userRepo,
		documentUC:  documentUC,
		access:      NewThesisAccess(thesisRepo),
		emailSvc:    emailSvc,
		auditSvc:    auditSvc,
		notifSvc:    notifSvc,
	}
}

// Submit creates a seminar proposal for a thesis (Mahasiswa pemilik only).
func (uc *SeminarUseCase) Submit(ctx context.Context, thesisID uuid.UUID, actor Actor) (*SeminarDetail, error) {
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

	// No active seminar (status not failed) for this thesis — checked first so
	// a second submission reports the precise conflict, not the state-machine gate.
	existing, err := uc.seminarRepo.FindByThesisID(ctx, thesisID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil && existing.Status != SeminarStatusFailed {
		return nil, ErrSeminarActiveExists
	}

	// Gate: thesis must be in_progress and seminar_doc approved.
	if thesis.Status != "in_progress" {
		return nil, ErrSeminarGateNotMet
	}
	gateOK, err := uc.documentUC.CanSubmitSeminar(ctx, thesisID)
	if err != nil {
		return nil, err
	}
	if !gateOK {
		return nil, ErrSeminarGateNotMet
	}

	seminar := &entity.Seminar{
		ThesisID: thesisID,
		Status:   SeminarStatusPending,
	}
	if err := uc.seminarRepo.Create(ctx, seminar); err != nil {
		return nil, err
	}

	// Update thesis status → seminar_ready.
	if err := uc.thesisRepo.UpdateStatus(ctx, thesisID, "seminar_ready", ""); err != nil {
		return nil, err
	}

	// Notify kaprodi + admin (async, non-fatal).
	go func() {
		ids := make([]uuid.UUID, 0, 4)
		emails := make([]string, 0, 4)
		kaprodi, _ := uc.userRepo.FindByRole(context.Background(), ThesisRoleKaprodi)
		adminUsers, _ := uc.userRepo.FindByRole(context.Background(), ThesisRoleAdminFakultas)
		for _, k := range kaprodi {
			emails = append(emails, k.Email)
			ids = append(ids, k.ID)
		}
		for _, a := range adminUsers {
			emails = append(emails, a.Email)
			ids = append(ids, a.ID)
		}
		if len(emails) > 0 {
			_ = uc.emailSvc.SendSeminarSubmitted(context.Background(), emails, seminar)
		}
		if len(ids) > 0 {
			uc.notifSvc.Notify(notification.Params{
				UserIDs: ids,
				Title:   "Pengajuan Seminar Proposal Baru",
				Message: thesis.Student.FullName + " mengajukan seminar proposal.",
				Type:    "seminar",
				Link:    notification.Path("/seminars/%s", seminar.ID),
			})
		}
	}()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionSeminarSubmitted,
		EntityType: "seminar",
		EntityID:   &seminar.ID,
		NewValue:   map[string]interface{}{"thesis_id": thesisID, "status": SeminarStatusPending},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	seminar, err = uc.seminarRepo.FindByID(ctx, seminar.ID)
	if err != nil {
		return nil, err
	}
	return toSeminarDetail(seminar), nil
}

// List returns seminars scoped by role (Admin/Kaprodi: all; Penguji: assigned;
// Mahasiswa/Pembimbing: related to their theses).
func (uc *SeminarUseCase) List(ctx context.Context, filter repository.SeminarFilter, userID uuid.UUID, role string) ([]*SeminarDetail, int64, error) {
	switch role {
	case ThesisRoleMahasiswa:
		filter.StudentID = userID
	case ThesisRoleDosenPembimbing:
		filter.SupervisorID = userID
	case ThesisRoleDosenPenguji:
		filter.ExaminerID = userID
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		// all seminars
	}

	seminars, total, err := uc.seminarRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	details := make([]*SeminarDetail, 0, len(seminars))
	for _, s := range seminars {
		details = append(details, toSeminarDetail(s))
	}
	return details, total, nil
}

// GetByID returns a single seminar (same access as List).
func (uc *SeminarUseCase) GetByID(ctx context.Context, id, userID uuid.UUID, role string) (*SeminarDetail, error) {
	seminar, err := uc.seminarRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSeminarNotFound
		}
		return nil, err
	}
	ok, err := uc.canAccessSeminar(ctx, seminar, userID, role)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return toSeminarDetail(seminar), nil
}

// Schedule schedules or reschedules a seminar (Admin + Kaprodi only).
func (uc *SeminarUseCase) Schedule(ctx context.Context, id uuid.UUID, req ScheduleSeminarRequest, actor Actor) (*SeminarDetail, error) {
	seminar, err := uc.seminarRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSeminarNotFound
		}
		return nil, err
	}
	if seminar.Status != SeminarStatusPending && seminar.Status != SeminarStatusScheduled {
		return nil, ErrSeminarNotScheduled
	}

	if time.Until(req.ScheduledAt) < MinSeminarLeadDays*24*time.Hour {
		return nil, ErrSeminarScheduleLeadTime
	}
	if len(req.ExaminerIDs) < MinSeminarExaminers {
		return nil, ErrSeminarMinExaminers
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
			return nil, ErrSeminarInvalidExaminer
		}
		seen[eid] = true
		examinerIDs = append(examinerIDs, eid)
		examinerEmails = append(examinerEmails, u.Email)
	}
	if len(seen) < MinSeminarExaminers {
		return nil, ErrSeminarMinExaminers
	}

	// Check room + examiner conflicts.
	excludeID := &id
	conflict, err := uc.seminarRepo.CheckScheduleConflict(ctx, req.Room, req.ScheduledAt, examinerIDs, excludeID)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, ErrSeminarRoomConflict
	}

	// Replace examiners (remove old, insert new — deduped list only).
	if err := uc.seminarRepo.RemoveAllExaminers(ctx, id); err != nil {
		return nil, err
	}
	for _, eid := range examinerIDs {
		if err := uc.seminarRepo.AssignExaminer(ctx, id, eid, actor.UserID); err != nil {
			return nil, err
		}
	}

	// Update schedule + status.
	if err := uc.seminarRepo.UpdateSchedule(ctx, id, req.ScheduledAt, req.Room); err != nil {
		return nil, err
	}
	wasScheduled := seminar.Status == SeminarStatusScheduled
	if err := uc.seminarRepo.UpdateStatus(ctx, id, SeminarStatusScheduled); err != nil {
		return nil, err
	}

	// Notify student, supervisors, examiners (async, non-fatal). Snapshot the
	// values so the goroutine does not race with the re-fetch below (line ~417)
	// that reassigns `seminar`.
	sem := *seminar
	emails := append([]string(nil), examinerEmails...)
	eids := append([]uuid.UUID(nil), examinerIDs...)
	go func() {
		recipients := seminarRecipients(&sem, emails)
		if len(recipients) > 0 {
			_ = uc.emailSvc.SendSeminarScheduled(context.Background(), recipients, &sem)
		}
		uc.notifSvc.Notify(notification.Params{
			UserIDs: append(append(userIDs(sem.Thesis.Supervisors), eids...), sem.Thesis.Student.ID),
			Title:   "Jadwal Seminar Proposal",
			Message: "Jadwal seminar proposal " + sem.Thesis.Student.FullName + " telah ditetapkan.",
			Type:    "seminar",
			Link:    notification.Path("/seminars/%s", sem.ID),
		})
	}()

	action := audit.ActionSeminarScheduled
	if wasScheduled {
		action = audit.ActionSeminarRescheduled
	}
	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     action,
		EntityType: "seminar",
		EntityID:   &seminar.ID,
		NewValue: map[string]interface{}{
			"scheduled_at": req.ScheduledAt,
			"room":         req.Room,
			"examiner_ids": req.ExaminerIDs,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})

	seminar, err = uc.seminarRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toSeminarDetail(seminar), nil
}

// SubmitScores records an examiner's scores and triggers finalization when all
// examiners have submitted (Dosen Penguji yang ditugaskan only).
func (uc *SeminarUseCase) SubmitScores(ctx context.Context, id uuid.UUID, req SubmitScoreRequest, actor Actor) (*SeminarResult, error) {
	seminar, err := uc.seminarRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSeminarNotFound
		}
		return nil, err
	}
	if seminar.Status != SeminarStatusScheduled {
		return nil, ErrSeminarNotScheduled
	}

	// Examiner must be assigned.
	examiners, err := uc.seminarRepo.GetExaminers(ctx, id)
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
		return nil, ErrSeminarNotExaminer
	}

	// Not already scored.
	already, err := uc.seminarRepo.HasExaminerScored(ctx, id, actor.UserID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, ErrSeminarAlreadyScored
	}

	// Validate: all components present, names match, scores 0-100.
	if err := validateScoreComponents(entity.SeminarGradingComponents, req.Scores, ErrSeminarIncompleteScore, ErrSeminarInvalidScore); err != nil {
		return nil, err
	}

	for _, c := range req.Scores {
		weight := 0.0
		for _, g := range entity.SeminarGradingComponents {
			if g.Name == c.ComponentName {
				weight = g.Weight
				break
			}
		}
		score := &entity.SeminarScore{
			SeminarID:       id,
			ExaminerID:      actor.UserID,
			ComponentName:   c.ComponentName,
			ComponentWeight: weight,
			Score:           c.Score,
		}
		if err := uc.seminarRepo.AddScore(ctx, score); err != nil {
			return nil, err
		}
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionSeminarScoreSubmit,
		EntityType: "seminar",
		EntityID:   &id,
		NewValue:   map[string]interface{}{"examiner_id": actor.UserID},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	// Try to finalize when all examiners have submitted.
	if err := uc.TryFinalizeSeminar(ctx, id); err != nil {
		return nil, err
	}
	return uc.Result(ctx, id, actor.UserID, ThesisRoleDosenPenguji)
}

// Result returns the full breakdown for a seminar (all related parties).
func (uc *SeminarUseCase) Result(ctx context.Context, id, userID uuid.UUID, role string) (*SeminarResult, error) {
	seminar, err := uc.seminarRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSeminarNotFound
		}
		return nil, err
	}
	ok, err := uc.canAccessSeminar(ctx, seminar, userID, role)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}

	scores, err := uc.seminarRepo.GetAllScores(ctx, id)
	if err != nil {
		return nil, err
	}
	examiners, err := uc.seminarRepo.GetExaminers(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &SeminarResult{
		SeminarID:         id,
		Status:            seminar.Status,
		FinalScore:        seminar.FinalScore,
		GradingComponents: toGradingComponentViews(entity.SeminarGradingComponents),
	}

	// Group scores by examiner.
	byExaminer := map[uuid.UUID][]*entity.SeminarScore{}
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

	// is_complete: every assigned examiner has submitted.
	scoredCount := len(order)
	result.IsComplete = scoredCount >= len(examiners) && len(examiners) > 0
	return result, nil
}

// TryFinalizeSeminar computes the final score once all examiners submitted and
// updates seminar + thesis statuses (triggered after every score submission).
func (uc *SeminarUseCase) TryFinalizeSeminar(ctx context.Context, seminarID uuid.UUID) error {
	examiners, err := uc.seminarRepo.GetExaminers(ctx, seminarID)
	if err != nil {
		return err
	}
	scoredCount, err := uc.seminarRepo.CountDistinctScoredExaminers(ctx, seminarID)
	if err != nil {
		return err
	}
	if scoredCount < len(examiners) || len(examiners) == 0 {
		return nil // not all examiners have submitted yet
	}

	allScores, err := uc.seminarRepo.GetAllScores(ctx, seminarID)
	if err != nil {
		return err
	}

	// Compute per-examiner weighted scores then average them.
	order := []string{}
	perExaminer := map[string]float64{}
	for _, s := range allScores {
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
	finalScore := grading.CalculateFinalScore(examinerScores)

	status := SeminarStatusPassed
	if finalScore < SeminarPassScore {
		status = SeminarStatusFailed
	}

	if err := uc.seminarRepo.UpdateFinalScore(ctx, seminarID, finalScore); err != nil {
		return err
	}
	if err := uc.seminarRepo.UpdateStatus(ctx, seminarID, status); err != nil {
		return err
	}

	// Update thesis status: seminar_done on pass; on fail it stays seminar_ready
	// (the student may submit again).
	seminar, err := uc.seminarRepo.FindByID(ctx, seminarID)
	if err != nil {
		return err
	}
	if status == SeminarStatusPassed {
		if err := uc.thesisRepo.UpdateStatus(ctx, seminar.ThesisID, "seminar_done", ""); err != nil {
			return err
		}
	}

	// Notify the student (async, non-fatal).
	thesis, err := uc.thesisRepo.FindByID(ctx, seminar.ThesisID)
	if err == nil {
		studentEmail := thesis.Student.Email
		go func() {
			_ = uc.emailSvc.SendSeminarFinalized(context.Background(), studentEmail, seminar)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: []uuid.UUID{thesis.Student.ID},
				Title:   "Hasil Seminar Proposal",
				Message: "Hasil seminar proposal Anda telah dirilis.",
				Type:    "seminar",
				Link:    notification.Path("/seminars/%s", seminarID),
			})
		}()
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     nil,
		Action:     audit.ActionSeminarFinalized,
		EntityType: "seminar",
		EntityID:   &seminarID,
		NewValue: map[string]interface{}{
			"final_score": finalScore,
			"status":      status,
		},
	})
	return nil
}

// SetRevisionNotes records post-seminar revision notes (Admin + Kaprodi only).
func (uc *SeminarUseCase) SetRevisionNotes(ctx context.Context, id uuid.UUID, notes string, actor Actor) (*SeminarDetail, error) {
	seminar, err := uc.seminarRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSeminarNotFound
		}
		return nil, err
	}
	if seminar.Status != SeminarStatusPassed {
		return nil, ErrSeminarNotPassed
	}

	if err := uc.seminarRepo.UpdateNotes(ctx, id, notes); err != nil {
		return nil, err
	}
	seminar.Notes = &notes

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionSeminarRevisionNote,
		EntityType: "seminar",
		EntityID:   &seminar.ID,
		NewValue:   map[string]interface{}{"notes": notes},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return toSeminarDetail(seminar), nil
}

// canAccessSeminar enforces per-role access to a single seminar.
func (uc *SeminarUseCase) canAccessSeminar(ctx context.Context, seminar *entity.Seminar, userID uuid.UUID, role string) (bool, error) {
	switch role {
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		return true, nil
	case ThesisRoleMahasiswa:
		return uc.access.IsThesisOwner(ctx, userID, seminar.ThesisID)
	case ThesisRoleDosenPembimbing:
		return uc.access.IsSupervisor(ctx, userID, seminar.ThesisID)
	case ThesisRoleDosenPenguji:
		examiners, err := uc.seminarRepo.GetExaminers(ctx, seminar.ID)
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

// validateScoreComponents checks that every required component is present with
// a valid score, and no unknown component names are supplied. The caller passes
// its own errors so the returned error belongs to the calling module (seminar
// vs defense) and the handler maps it to the right status code.
func validateScoreComponents(components []entity.GradingComponent, input []ComponentScoreInput, incompleteErr, invalidErr error) error {
	if len(input) != len(components) {
		return incompleteErr
	}
	present := map[string]bool{}
	for _, c := range input {
		if !entity.ValidGradingComponents(components, c.ComponentName) {
			return incompleteErr
		}
		if c.Score < 0 || c.Score > 100 {
			return invalidErr
		}
		present[c.ComponentName] = true
	}
	for _, g := range components {
		if !present[g.Name] {
			return incompleteErr
		}
	}
	return nil
}

// toSeminarDetail maps a persisted seminar into the API response shape.
func toSeminarDetail(s *entity.Seminar) *SeminarDetail {
	d := &SeminarDetail{
		ID:          s.ID,
		Status:      s.Status,
		ScheduledAt: s.ScheduledAt,
		Room:        s.Room,
		FinalScore:  s.FinalScore,
		Notes:       s.Notes,
		CreatedAt:   s.CreatedAt,
	}
	if s.Thesis.ID != uuid.Nil {
		d.Thesis = &SeminarThesisInfo{ID: s.Thesis.ID, Title: s.Thesis.Title}
		if s.Thesis.Student.ID != uuid.Nil {
			d.Thesis.Student = &ThesisStudentBrief{
				FullName: s.Thesis.Student.FullName,
				Nim:      s.Thesis.Student.NimNidn,
			}
		}
	}
	for _, e := range s.Examiners {
		if e.ID != uuid.Nil {
			d.Examiners = append(d.Examiners, &UserBrief{ID: e.ID, FullName: e.FullName})
		}
	}
	for _, sc := range s.Scores {
		detail := &SeminarScoreDetail{
			ComponentName:   sc.ComponentName,
			ComponentWeight: sc.ComponentWeight,
			Score:           sc.Score,
		}
		if sc.Examiner.ID != uuid.Nil {
			detail.Examiner = &UserBrief{ID: sc.Examiner.ID, FullName: sc.Examiner.FullName}
		}
		d.Scores = append(d.Scores, detail)
	}
	return d
}

// toGradingComponentViews maps entity grading components to the response views.
func toGradingComponentViews(components []entity.GradingComponent) []GradingComponentView {
	views := make([]GradingComponentView, 0, len(components))
	for _, c := range components {
		views = append(views, GradingComponentView{Name: c.Name, Weight: c.Weight})
	}
	return views
}

// seminarRecipients collects student + supervisor emails for a seminar.
func seminarRecipients(seminar *entity.Seminar, examinerEmails []string) []string {
	recipients := []string{}
	if seminar.Thesis.Student.Email != "" {
		recipients = append(recipients, seminar.Thesis.Student.Email)
	}
	for _, s := range seminar.Thesis.Supervisors {
		if s.Email != "" {
			recipients = append(recipients, s.Email)
		}
	}
	recipients = append(recipients, examinerEmails...)
	return recipients
}

// collectRoleEmails returns active emails for a role (best effort).
func collectRoleEmails(ctx context.Context, userRepo repository.UserRepository, role string) []string {
	users, err := userRepo.FindByRole(ctx, role)
	if err != nil {
		return nil
	}
	emails := make([]string, 0, len(users))
	for _, u := range users {
		emails = append(emails, u.Email)
	}
	return emails
}

// findUserByID returns a UserBrief for the given ID or nil.
func findUserByID(users []*entity.User, id uuid.UUID) *UserBrief {
	for _, u := range users {
		if u.ID == id {
			return &UserBrief{ID: u.ID, FullName: u.FullName}
		}
	}
	return &UserBrief{ID: id}
}
