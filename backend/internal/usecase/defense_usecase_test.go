package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/grading"
)

// fakeDefenseRepo is a minimal in-memory DefenseRepository for usecase tests.
type fakeDefenseRepo struct {
	defenses       map[uuid.UUID]*entity.ThesisDefense
	byThesis       map[uuid.UUID]uuid.UUID // thesisID -> defenseID (latest)
	examiners      map[uuid.UUID][]uuid.UUID
	examinerUser   map[uuid.UUID]*entity.User // examinerID -> user (for role checks)
	scores         map[uuid.UUID][]*entity.DefenseScore
	scheduled      []entity.ThesisDefense
	revisionNotes  map[uuid.UUID]string
	thesisStudents map[uuid.UUID]uuid.UUID // thesisID -> studentID (for FindAll scoping)
}

func newFakeDefenseRepo() *fakeDefenseRepo {
	return &fakeDefenseRepo{
		defenses:       map[uuid.UUID]*entity.ThesisDefense{},
		byThesis:       map[uuid.UUID]uuid.UUID{},
		examiners:      map[uuid.UUID][]uuid.UUID{},
		examinerUser:   map[uuid.UUID]*entity.User{},
		scores:         map[uuid.UUID][]*entity.DefenseScore{},
		revisionNotes:  map[uuid.UUID]string{},
		thesisStudents: map[uuid.UUID]uuid.UUID{},
	}
}

var _ repository.DefenseRepository = (*fakeDefenseRepo)(nil)

func (f *fakeDefenseRepo) Create(_ context.Context, defense *entity.ThesisDefense) error {
	defense.ID = uuid.New()
	f.defenses[defense.ID] = defense
	f.byThesis[defense.ThesisID] = defense.ID
	return nil
}

func (f *fakeDefenseRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.ThesisDefense, error) {
	d, ok := f.defenses[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *d
	for _, eid := range f.examiners[id] {
		if u, ok := f.examinerUser[eid]; ok {
			clone.Examiners = append(clone.Examiners, *u)
		} else {
			clone.Examiners = append(clone.Examiners, entity.User{ID: eid})
		}
	}
	return &clone, nil
}

func (f *fakeDefenseRepo) FindByThesisID(_ context.Context, thesisID uuid.UUID) (*entity.ThesisDefense, error) {
	id, ok := f.byThesis[thesisID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return f.FindByID(context.Background(), id)
}

func (f *fakeDefenseRepo) FindAll(_ context.Context, filter repository.DefenseFilter) ([]*entity.ThesisDefense, int64, error) {
	var out []*entity.ThesisDefense
	for _, d := range f.defenses {
		if filter.Status != "" && d.Status != filter.Status {
			continue
		}
		// Role scoping: student sees defenses of their own theses only.
		if filter.StudentID != uuid.Nil && f.thesisStudents[d.ThesisID] != filter.StudentID {
			continue
		}
		clone := *d
		out = append(out, &clone)
	}
	return out, int64(len(out)), nil
}

func (f *fakeDefenseRepo) UpdateSchedule(_ context.Context, id uuid.UUID, scheduledAt time.Time, room string) error {
	d, ok := f.defenses[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	d.ScheduledAt = &scheduledAt
	d.Room = &room
	return nil
}

func (f *fakeDefenseRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	d, ok := f.defenses[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	d.Status = status
	return nil
}

func (f *fakeDefenseRepo) UpdateFinalScore(_ context.Context, id uuid.UUID, score float64) error {
	d, ok := f.defenses[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	d.FinalScore = &score
	return nil
}

func (f *fakeDefenseRepo) SetRevisionNotes(_ context.Context, id uuid.UUID, notes string) error {
	d, ok := f.defenses[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	d.RevisionNotes = &notes
	f.revisionNotes[id] = notes
	return nil
}

func (f *fakeDefenseRepo) AssignExaminer(_ context.Context, defenseID, examinerID, _ uuid.UUID) error {
	if _, ok := f.defenses[defenseID]; !ok {
		return gorm.ErrRecordNotFound
	}
	f.examiners[defenseID] = append(f.examiners[defenseID], examinerID)
	return nil
}

func (f *fakeDefenseRepo) RemoveAllExaminers(_ context.Context, defenseID uuid.UUID) error {
	f.examiners[defenseID] = nil
	return nil
}

func (f *fakeDefenseRepo) GetExaminers(_ context.Context, defenseID uuid.UUID) ([]*entity.User, error) {
	var out []*entity.User
	for _, eid := range f.examiners[defenseID] {
		if u, ok := f.examinerUser[eid]; ok {
			out = append(out, u)
		} else {
			out = append(out, &entity.User{ID: eid, FullName: "Penguji"})
		}
	}
	return out, nil
}

func (f *fakeDefenseRepo) AddScore(_ context.Context, score *entity.DefenseScore) error {
	score.ID = uuid.New()
	f.scores[score.DefenseID] = append(f.scores[score.DefenseID], score)
	return nil
}

func (f *fakeDefenseRepo) GetAllScores(_ context.Context, defenseID uuid.UUID) ([]*entity.DefenseScore, error) {
	return f.scores[defenseID], nil
}

func (f *fakeDefenseRepo) HasExaminerScored(_ context.Context, defenseID, examinerID uuid.UUID) (bool, error) {
	for _, s := range f.scores[defenseID] {
		if s.ExaminerID == examinerID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeDefenseRepo) CountDistinctScoredExaminers(_ context.Context, defenseID uuid.UUID) (int, error) {
	seen := map[uuid.UUID]bool{}
	for _, s := range f.scores[defenseID] {
		seen[s.ExaminerID] = true
	}
	return len(seen), nil
}

// FinalizeDefense mirrors the repository's atomic finalize in-memory so use case
// tests can exercise the score-submission → finalize flow.
func (f *fakeDefenseRepo) FinalizeDefense(_ context.Context, defenseID uuid.UUID) (float64, string, uuid.UUID, error) {
	d, ok := f.defenses[defenseID]
	if !ok {
		return 0, "", uuid.Nil, gorm.ErrRecordNotFound
	}
	if d.Status != DefenseStatusScheduled {
		return 0, "", uuid.Nil, nil
	}
	examiners := f.examiners[defenseID]
	scored := map[uuid.UUID]bool{}
	for _, s := range f.scores[defenseID] {
		scored[s.ExaminerID] = true
	}
	if len(scored) < len(examiners) {
		return 0, "", uuid.Nil, nil
	}
	order := []string{}
	perExaminer := map[string]float64{}
	for _, s := range f.scores[defenseID] {
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
	status := DefenseStatusPassed
	switch {
	case fs < DefenseFailThreshold:
		status = DefenseStatusFailed
	case fs < DefenseRevisionThreshold:
		status = DefenseStatusRevisionRequired
	}
	fsCopy := fs
	d.FinalScore = &fsCopy
	d.Status = status
	return fs, status, d.ThesisID, nil
}

func (f *fakeDefenseRepo) CheckScheduleConflict(_ context.Context, room string, _ time.Time, examinerIDs []uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	for _, d := range f.scheduled {
		if excludeID != nil && d.ID == *excludeID {
			continue
		}
		if d.Room != nil && *d.Room == room {
			return true, nil
		}
	}
	for _, d := range f.scheduled {
		if excludeID != nil && d.ID == *excludeID {
			continue
		}
		for _, eid := range examinerIDs {
			for _, assigned := range f.examiners[d.ID] {
				if assigned == eid {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// registerExaminerUser stores a dosen_penguji user so role validation passes.
func (f *fakeDefenseRepo) registerExaminerUser(u *entity.User) {
	f.examinerUser[u.ID] = u
}

// recordingDefenseEmailService records defense email recipients for assertions.
type recordingDefenseEmailService struct {
	submittedTo chan []string
	scheduledTo chan []string
	finalizedTo chan string
	graduatedTo chan string
}

func newRecordingDefenseEmailService() *recordingDefenseEmailService {
	return &recordingDefenseEmailService{
		submittedTo: make(chan []string, 16),
		scheduledTo: make(chan []string, 16),
		finalizedTo: make(chan string, 16),
		graduatedTo: make(chan string, 16),
	}
}

func (r *recordingDefenseEmailService) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (r *recordingDefenseEmailService) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
func (r *recordingDefenseEmailService) SendThesisSubmitted(context.Context, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingDefenseEmailService) SendThesisApproved(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingDefenseEmailService) SendThesisRejected(context.Context, string, *entity.Thesis, string) error {
	return nil
}
func (r *recordingDefenseEmailService) SendSupervisorAssigned(context.Context, string, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingDefenseEmailService) SendConsultationCreated(context.Context, []string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingDefenseEmailService) SendConsultationApproved(context.Context, string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingDefenseEmailService) SendDocumentUploaded(context.Context, []string, *entity.Document) error {
	return nil
}
func (r *recordingDefenseEmailService) SendDocumentReviewed(context.Context, string, *entity.Document, string) error {
	return nil
}
func (r *recordingDefenseEmailService) SendSeminarSubmitted(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingDefenseEmailService) SendSeminarScheduled(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingDefenseEmailService) SendSeminarFinalized(context.Context, string, *entity.Seminar) error {
	return nil
}
func (r *recordingDefenseEmailService) SendDefenseSubmitted(_ context.Context, to []string, _ *entity.ThesisDefense) error {
	r.submittedTo <- to
	return nil
}
func (r *recordingDefenseEmailService) SendDefenseScheduled(_ context.Context, to []string, _ *entity.ThesisDefense) error {
	r.scheduledTo <- to
	return nil
}
func (r *recordingDefenseEmailService) SendDefenseFinalized(_ context.Context, to string, _ *entity.ThesisDefense) error {
	r.finalizedTo <- to
	return nil
}
func (r *recordingDefenseEmailService) SendGraduated(_ context.Context, to string, _ *entity.Thesis) error {
	r.graduatedTo <- to
	return nil
}
func (r *recordingDefenseEmailService) SendArchiveCreated(context.Context, string, *entity.ThesisArchive) error {
	return nil
}

func (r *recordingDefenseEmailService) SendTitleChangeRequested(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingDefenseEmailService) SendTitleChangeCancelled(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingDefenseEmailService) SendTitleChangeApproved(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingDefenseEmailService) SendTitleChangeRejected(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

// newTestDefenseUseCase wires a fresh defense use case with in-memory fakes.
func newTestDefenseUseCase(t *testing.T) (*DefenseUseCase, *fakeDefenseRepo, *fakeSeminarRepo, *fakeDocumentRepo, *fakeThesisRepo, *fakeUserRepo, *recordingDefenseEmailService) {
	t.Helper()
	defRepo := newFakeDefenseRepo()
	semRepo := newFakeSeminarRepo()
	docRepo := newFakeDocumentRepo()
	thesisRepo := newFakeThesisRepo()
	userRepo := newFakeUserRepo()
	email := newRecordingDefenseEmailService()

	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	userRepo.roles["dosen_pembimbing"] = &entity.Role{ID: 4, Name: "dosen_pembimbing"}
	userRepo.roles["dosen_penguji"] = &entity.Role{ID: 5, Name: "dosen_penguji"}
	userRepo.roles["kaprodi"] = &entity.Role{ID: 2, Name: "kaprodi"}
	userRepo.roles["admin_fakultas"] = &entity.Role{ID: 1, Name: "admin_fakultas"}

	// Seed a kaprodi so the submit-notification email actually fires.
	kaprodi := &entity.User{
		Email:    "kaprodi@example.com",
		FullName: "Kaprodi",
		Role:     entity.Role{Name: "kaprodi"},
		RoleID:   2,
		IsActive: true,
	}
	if err := userRepo.Create(context.Background(), kaprodi); err != nil {
		t.Fatalf("seed kaprodi: %v", err)
	}

	auditSvc := audit.NewAuditService(nil)
	documentUC := NewDocumentUseCase(docRepo, thesisRepo, nil, email, auditSvc)
	uc := NewDefenseUseCase(defRepo, semRepo, thesisRepo, userRepo, documentUC, email, auditSvc)
	return uc, defRepo, semRepo, docRepo, thesisRepo, userRepo, email
}

// seedDefenseReadyThesis builds a seminar_done thesis with an approved
// defense_doc so the defense gate passes; returns thesisID, studentID, supervisorID.
func seedDefenseReadyThesis(t *testing.T, defRepo *fakeDefenseRepo, docRepo *fakeDocumentRepo, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)
	// Advance the thesis past the seminar milestone (defense gate requires seminar_done).
	thesisRepo.theses[thesisID].Status = "seminar_done"
	// Register the thesis→student mapping so FindAll role-scoping works.
	defRepo.thesisStudents[thesisID] = studentID
	// Simulate an approved defense_doc via the doc repo's approvedTypes map.
	docRepo.approvedTypes[thesisID.String()+":"+entity.DocTypeDefenseDoc] = true
	return thesisID, studentID, supervisorID
}

// seedDefenseExaminers registers N active dosen_penguji and returns their IDs.
func seedDefenseExaminers(t *testing.T, userRepo *fakeUserRepo, defRepo *fakeDefenseRepo, n int) []uuid.UUID {
	t.Helper()
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		u := &entity.User{
			// Unique email so multiple calls (e.g. adding a non-examiner) never collide.
			Email:    "penguji-def-" + uuid.New().String()[:8] + "@example.com",
			FullName: "Penguji " + string(rune('A'+i)),
			Role:     entity.Role{Name: "dosen_penguji"},
			RoleID:   5,
			IsActive: true,
		}
		if err := userRepo.Create(context.Background(), u); err != nil {
			t.Fatalf("seed penguji: %v", err)
		}
		ids = append(ids, u.ID)
		defRepo.registerExaminerUser(u)
	}
	return ids
}

func TestSubmitDefenseGateNotMet(t *testing.T) {
	uc, _, _, _, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	// Thesis still in_progress (seminar not done) → gate fails.
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	_, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if !errors.Is(err, ErrDefenseGateNotMet) {
		t.Errorf("expected ErrDefenseGateNotMet, got %v", err)
	}
}

func TestSubmitDefenseDocNotApproved(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	// Remove the approved defense_doc → gate fails at the doc check.
	delete(docRepo.approvedTypes, thesisID.String()+":"+entity.DocTypeDefenseDoc)

	_, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if !errors.Is(err, ErrDefenseGateNotMet) {
		t.Errorf("expected ErrDefenseGateNotMet, got %v", err)
	}
}

func TestSubmitDefenseSuccess(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, email := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)

	detail, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if detail.Status != DefenseStatusPending {
		t.Errorf("status = %q, want pending", detail.Status)
	}
	if len(defRepo.defenses) != 1 {
		t.Errorf("expected 1 defense, got %d", len(defRepo.defenses))
	}
	// Thesis moved to defense_ready.
	th := thesisRepo.theses[thesisID]
	if th.Status != "defense_ready" {
		t.Errorf("thesis status = %q, want defense_ready", th.Status)
	}
	// Kaprodi + admin are notified (async).
	select {
	case to := <-email.submittedTo:
		if len(to) < 1 {
			t.Errorf("expected kaprodi/admin recipients, got %v", to)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for defense-submitted email")
	}
}

func TestSubmitDefenseForbiddenForOtherStudent(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, _, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	other := seedStudent(t, userRepo, "mahasiswa")

	_, err := uc.Submit(context.Background(), thesisID, Actor{UserID: other})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestSubmitDefenseActiveExists(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)

	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	_, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if !errors.Is(err, ErrDefenseActiveExists) {
		t.Errorf("expected ErrDefenseActiveExists, got %v", err)
	}
}

func TestScheduleDefense(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)

	detail, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang Sidang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Schedule returned error: %v", err)
	}
	if detail.Status != DefenseStatusScheduled {
		t.Errorf("status = %q, want scheduled", detail.Status)
	}
	if len(defRepo.examiners[defenseID]) != 2 {
		t.Errorf("expected 2 examiners, got %d", len(defRepo.examiners[defenseID]))
	}
}

func TestScheduleDefenseLeadTimeTooShort(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)

	_, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(3), // < 7 days
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrDefenseScheduleLeadTime) {
		t.Errorf("expected ErrDefenseScheduleLeadTime, got %v", err)
	}
}

func TestScheduleDefenseMinExaminers(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)

	_, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners[:1], // only 1
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrDefenseMinExaminers) {
		t.Errorf("expected ErrDefenseMinExaminers, got %v", err)
	}
}

func TestScheduleDefenseRoomConflict(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)

	// Mark an existing defense occupying "Ruang Sidang A".
	existing := &entity.ThesisDefense{ID: uuid.New(), Status: DefenseStatusScheduled}
	room := "Ruang Sidang A"
	existing.Room = &room
	defRepo.scheduled = append(defRepo.scheduled, *existing)

	_, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang Sidang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrDefenseRoomConflict) {
		t.Errorf("expected ErrDefenseRoomConflict, got %v", err)
	}
}

func TestSubmitScoresAndFinalizeDefense(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, email := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Examiner 1 submits 80 across all components.
	result1, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 80),
	}, Actor{UserID: examiners[0]})
	if err != nil {
		t.Fatalf("submit scores #1: %v", err)
	}
	if result1.IsComplete {
		t.Error("result should not be complete after one examiner")
	}

	// Examiner 2 submits 90 → finalize triggered.
	result2, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 90),
	}, Actor{UserID: examiners[1]})
	if err != nil {
		t.Fatalf("submit scores #2: %v", err)
	}
	if !result2.IsComplete {
		t.Error("result should be complete after all examiners")
	}
	// Average of 80 and 90 = 85.
	if result2.FinalScore == nil || *result2.FinalScore != 85 {
		t.Errorf("final score = %v, want 85", result2.FinalScore)
	}
	if result2.Status != DefenseStatusPassed {
		t.Errorf("status = %q, want passed", result2.Status)
	}
	if result2.GradeCategory != "A" {
		t.Errorf("grade category = %q, want A", result2.GradeCategory)
	}

	// Thesis moved to defense_done.
	th := thesisRepo.theses[thesisID]
	if th.Status != "defense_done" {
		t.Errorf("thesis status = %q, want defense_done", th.Status)
	}

	// Student notified of the final result.
	select {
	case to := <-email.finalizedTo:
		student := userRepo.users[studentID]
		if to != student.Email {
			t.Errorf("finalized email to = %q, want %q", to, student.Email)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for finalized email")
	}
}

// TestFinalizeDefenseIdempotent guards against the production-review race: a
// defense that is already finalized (status != "scheduled") must not be
// finalized a second time, so concurrent score submissions cannot both finalize.
func TestFinalizeDefenseIdempotent(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, email := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	for _, ex := range examiners {
		if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
			Scores: fullScoreInput(entity.DefenseGradingComponents, 80),
		}, Actor{UserID: ex}); err != nil {
			t.Fatalf("submit scores for %s: %v", ex, err)
		}
	}

	// Defense is now finalized (one finalized email is queued from SubmitScores).
	// Drain it, then attempt a second finalize which must be a no-op.
	select {
	case <-email.finalizedTo:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first finalized email")
	}

	if err := uc.TryFinalizeDefense(context.Background(), defenseID); err != nil {
		t.Fatalf("second finalize returned error: %v", err)
	}
	th := thesisRepo.theses[thesisID]
	if th.Status != "defense_done" {
		t.Errorf("thesis status = %q, want defense_done", th.Status)
	}
	// No second finalized email should have been emitted.
	select {
	case to := <-email.finalizedTo:
		t.Errorf("unexpected second finalized email to %q", to)
	case <-time.After(200 * time.Millisecond):
		// expected: no email
	}
}

func TestSubmitScoresFailedDefense(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Both examiners score 50 → final = 50 < 60 → failed.
	if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 50),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("scores #1: %v", err)
	}
	result, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 50),
	}, Actor{UserID: examiners[1]})
	if err != nil {
		t.Fatalf("scores #2: %v", err)
	}
	if result.Status != DefenseStatusFailed {
		t.Errorf("status = %q, want failed", result.Status)
	}
	if result.GradeCategory != "Tidak Lulus" {
		t.Errorf("grade category = %q, want Tidak Lulus", result.GradeCategory)
	}
	// Thesis returns to defense_ready on failure.
	th := thesisRepo.theses[thesisID]
	if th.Status != "defense_ready" {
		t.Errorf("thesis status = %q, want defense_ready (retry)", th.Status)
	}
}

func TestSubmitDefenseScoresByNonExaminer(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// A third dosen_penguji who was NOT assigned.
	other := seedDefenseExaminers(t, userRepo, defRepo, 1)[0]

	_, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 80),
	}, Actor{UserID: other})
	if !errors.Is(err, ErrDefenseNotExaminer) {
		t.Errorf("expected ErrDefenseNotExaminer, got %v", err)
	}
}

func TestSubmitDefenseScoresAlreadyScored(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 80),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	_, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 85),
	}, Actor{UserID: examiners[0]})
	if !errors.Is(err, ErrDefenseAlreadyScored) {
		t.Errorf("expected ErrDefenseAlreadyScored, got %v", err)
	}
}

func TestSubmitDefenseScoresIncompleteComponents(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Only one component supplied.
	_, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: []ComponentScoreInput{{ComponentName: "Presentasi", Score: 80}},
	}, Actor{UserID: examiners[0]})
	if !errors.Is(err, ErrDefenseIncompleteScore) {
		t.Errorf("expected ErrDefenseIncompleteScore, got %v", err)
	}
}

func TestDefenseResultBreakdown(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Examiner A scores 80 everywhere → 80.0.
	if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 80),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("scores #1: %v", err)
	}

	// Admin sees a partial breakdown with 1 examiner scored.
	result, err := uc.Result(context.Background(), defenseID, uuid.New(), ThesisRoleAdminFakultas)
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	if len(result.ExaminerScores) != 1 {
		t.Fatalf("examiner scores = %d, want 1", len(result.ExaminerScores))
	}
	if result.ExaminerScores[0].ExaminerScore != 80 {
		t.Errorf("examiner score = %v, want 80", result.ExaminerScores[0].ExaminerScore)
	}
	if result.IsComplete {
		t.Error("result should not be complete yet")
	}
	if len(result.GradingComponents) != 4 {
		t.Errorf("grading components = %d, want 4", len(result.GradingComponents))
	}
}

func TestSetDefenseRevisionNotes(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 80),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("scores #1: %v", err)
	}
	if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 90),
	}, Actor{UserID: examiners[1]}); err != nil {
		t.Fatalf("scores #2: %v", err)
	}

	detail, err := uc.SetRevisionNotes(context.Background(), defenseID, "Perbaiki kesimpulan dan saran di BAB 5", Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("SetRevisionNotes: %v", err)
	}
	if detail.Status != DefenseStatusRevisionRequired {
		t.Errorf("status = %q, want revision_required", detail.Status)
	}
	if detail.RevisionNotes == nil || *detail.RevisionNotes != "Perbaiki kesimpulan dan saran di BAB 5" {
		t.Errorf("revision notes not set in response")
	}
	if defRepo.revisionNotes[defenseID] != "Perbaiki kesimpulan dan saran di BAB 5" {
		t.Errorf("revision notes not persisted in repo")
	}
}

func TestSetRevisionNotesBeforeFinalizedRejected(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]

	_, err := uc.SetRevisionNotes(context.Background(), defenseID, "catatan", Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrDefenseNotFinalized) {
		t.Errorf("expected ErrDefenseNotFinalized, got %v", err)
	}
}

func TestGraduate(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, email := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	// Walk the full path: submit → schedule → pass → defense_done.
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 80),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("scores #1: %v", err)
	}
	if _, err := uc.SubmitScores(context.Background(), defenseID, SubmitDefenseScoreRequest{
		Scores: fullScoreInput(entity.DefenseGradingComponents, 90),
	}, Actor{UserID: examiners[1]}); err != nil {
		t.Fatalf("scores #2: %v", err)
	}
	// final_thesis must be approved for the graduation gate.
	docRepo.approvedTypes[thesisID.String()+":"+entity.DocTypeFinalThesis] = true

	err := uc.Graduate(context.Background(), thesisID, GraduationRequest{
		Notes: "Selamat, skripsi Anda telah memenuhi semua persyaratan.",
	}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Graduate returned error: %v", err)
	}
	th := thesisRepo.theses[thesisID]
	if th.Status != "graduated" {
		t.Errorf("thesis status = %q, want graduated", th.Status)
	}
	if th.GraduatedAt == nil {
		t.Error("graduated_at should be set")
	}

	// Student receives the graduation email.
	select {
	case to := <-email.graduatedTo:
		student := userRepo.users[studentID]
		if to != student.Email {
			t.Errorf("graduated email to = %q, want %q", to, student.Email)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for graduation email")
	}
}

func TestGraduateGateNotMet(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Thesis is defense_ready (not defense_done) → gate fails.
	err := uc.Graduate(context.Background(), thesisID, GraduationRequest{}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrGraduationGateNotMet) {
		t.Errorf("expected ErrGraduationGateNotMet, got %v", err)
	}
}

func TestUpcomingSchedules(t *testing.T) {
	uc, defRepo, semRepo, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)

	// A scheduled seminar within the window.
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit defense: %v", err)
	}
	// Seed a scheduled seminar directly (seminar flow is covered by Job 08 tests).
	sem := &entity.Seminar{
		ID:          uuid.New(),
		Status:      SeminarStatusScheduled,
		ScheduledAt: ptrTime(futureDays(3)),
		Room:        strPtr("Ruang Seminar A"),
		Thesis: entity.Thesis{
			Title:   "Judul Skripsi",
			Student: entity.User{FullName: "Mahasiswa Satu"},
		},
	}
	semRepo.seminars[sem.ID] = sem

	// The defense must be scheduled too (lead time >= 7 days).
	defenseID := defRepo.byThesis[thesisID]
	examiners := seedDefenseExaminers(t, userRepo, defRepo, 2)
	if _, err := uc.Schedule(context.Background(), defenseID, ScheduleDefenseRequest{
		ScheduledAt: futureDays(10),
		Room:        "Ruang Sidang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule defense: %v", err)
	}
	// The real repo preloads Thesis.Student; mirror that so the upcoming item
	// carries the student name.
	defRepo.defenses[defenseID].Thesis = entity.Thesis{
		Title:   "Judul Skripsi",
		Student: entity.User{FullName: "Mahasiswa Satu"},
	}

	schedules, err := uc.Upcoming(context.Background(), 14)
	if err != nil {
		t.Fatalf("Upcoming returned error: %v", err)
	}
	if len(schedules.Seminars) != 1 {
		t.Errorf("seminars = %d, want 1", len(schedules.Seminars))
	}
	if len(schedules.Defenses) != 1 {
		t.Errorf("defenses = %d, want 1", len(schedules.Defenses))
	}
	if len(schedules.Defenses) > 0 && schedules.Defenses[0].StudentName == "" {
		t.Error("defense upcoming item missing student name")
	}
}

func TestDefenseListScopedByRole(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Owner sees their defense.
	list, total, err := uc.List(context.Background(), repository.DefenseFilter{}, studentID, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 defense for owner, got total=%d len=%d", total, len(list))
	}

	// An unrelated student sees nothing (role scoping actually filters).
	other := seedStudent(t, userRepo, "mahasiswa")
	_, totalOther, err := uc.List(context.Background(), repository.DefenseFilter{}, other, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("List for other: %v", err)
	}
	if totalOther != 0 {
		t.Errorf("expected 0 defenses for unrelated student, got %d", totalOther)
	}
}

func TestGetGradeCategory(t *testing.T) {
	score := func(v float64) *float64 { return &v }
	tests := []struct {
		name  string
		score *float64
		want  string
	}{
		{"nil score", nil, "-"},
		{"A", score(85), "A"},
		{"B+", score(75), "B+"},
		{"B", score(70), "B"},
		{"C", score(60), "C"},
		{"fail", score(50), "Tidak Lulus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGradeCategory(tt.score); got != tt.want {
				t.Errorf("GetGradeCategory(%v) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
