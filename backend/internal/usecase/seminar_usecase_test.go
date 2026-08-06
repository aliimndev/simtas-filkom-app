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
)

// fakeSeminarRepo is a minimal in-memory SeminarRepository for usecase tests.
type fakeSeminarRepo struct {
	seminars       map[uuid.UUID]*entity.Seminar
	byThesis       map[uuid.UUID]uuid.UUID // thesisID -> seminarID (latest)
	examiners      map[uuid.UUID][]uuid.UUID
	examinerUser   map[uuid.UUID]*entity.User // examinerID -> user (for role checks)
	scores         map[uuid.UUID][]*entity.SeminarScore
	scheduled      []entity.Seminar
	notes          map[uuid.UUID]string
	thesisStudents map[uuid.UUID]uuid.UUID // thesisID -> studentID (for FindAll scoping)
}

func newFakeSeminarRepo() *fakeSeminarRepo {
	return &fakeSeminarRepo{
		seminars:       map[uuid.UUID]*entity.Seminar{},
		byThesis:       map[uuid.UUID]uuid.UUID{},
		examiners:      map[uuid.UUID][]uuid.UUID{},
		examinerUser:   map[uuid.UUID]*entity.User{},
		scores:         map[uuid.UUID][]*entity.SeminarScore{},
		notes:          map[uuid.UUID]string{},
		thesisStudents: map[uuid.UUID]uuid.UUID{},
	}
}

func (f *fakeSeminarRepo) Create(_ context.Context, seminar *entity.Seminar) error {
	seminar.ID = uuid.New()
	f.seminars[seminar.ID] = seminar
	f.byThesis[seminar.ThesisID] = seminar.ID
	return nil
}

var _ repository.SeminarRepository = (*fakeSeminarRepo)(nil)

func (f *fakeSeminarRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.Seminar, error) {
	s, ok := f.seminars[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *s
	// Populate examiners + thesis student so toSeminarDetail works.
	for _, eid := range f.examiners[id] {
		if u, ok := f.examinerUser[eid]; ok {
			clone.Examiners = append(clone.Examiners, *u)
		} else {
			clone.Examiners = append(clone.Examiners, entity.User{ID: eid})
		}
	}
	return &clone, nil
}

func (f *fakeSeminarRepo) FindByThesisID(_ context.Context, thesisID uuid.UUID) (*entity.Seminar, error) {
	id, ok := f.byThesis[thesisID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return f.FindByID(context.Background(), id)
}

func (f *fakeSeminarRepo) FindAll(_ context.Context, filter repository.SeminarFilter) ([]*entity.Seminar, int64, error) {
	var out []*entity.Seminar
	for _, s := range f.seminars {
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		// Role scoping: student sees seminars of their own theses only.
		if filter.StudentID != uuid.Nil && f.thesisStudents[s.ThesisID] != filter.StudentID {
			continue
		}
		clone := *s
		out = append(out, &clone)
	}
	return out, int64(len(out)), nil
}

func (f *fakeSeminarRepo) UpdateSchedule(_ context.Context, id uuid.UUID, scheduledAt time.Time, room string) error {
	s, ok := f.seminars[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	s.ScheduledAt = &scheduledAt
	s.Room = &room
	return nil
}

func (f *fakeSeminarRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	s, ok := f.seminars[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	s.Status = status
	return nil
}

func (f *fakeSeminarRepo) UpdateFinalScore(_ context.Context, id uuid.UUID, score float64) error {
	s, ok := f.seminars[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	s.FinalScore = &score
	return nil
}

func (f *fakeSeminarRepo) UpdateNotes(_ context.Context, id uuid.UUID, notes string) error {
	if _, ok := f.seminars[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	f.notes[id] = notes
	return nil
}

func (f *fakeSeminarRepo) AssignExaminer(_ context.Context, seminarID, examinerID, _ uuid.UUID) error {
	if _, ok := f.seminars[seminarID]; !ok {
		return gorm.ErrRecordNotFound
	}
	f.examiners[seminarID] = append(f.examiners[seminarID], examinerID)
	return nil
}

func (f *fakeSeminarRepo) RemoveAllExaminers(_ context.Context, seminarID uuid.UUID) error {
	f.examiners[seminarID] = nil
	return nil
}

func (f *fakeSeminarRepo) GetExaminers(_ context.Context, seminarID uuid.UUID) ([]*entity.User, error) {
	var out []*entity.User
	for _, eid := range f.examiners[seminarID] {
		if u, ok := f.examinerUser[eid]; ok {
			out = append(out, u)
		} else {
			out = append(out, &entity.User{ID: eid, FullName: "Penguji"})
		}
	}
	return out, nil
}

func (f *fakeSeminarRepo) AddScore(_ context.Context, score *entity.SeminarScore) error {
	score.ID = uuid.New()
	f.scores[score.SeminarID] = append(f.scores[score.SeminarID], score)
	return nil
}

func (f *fakeSeminarRepo) GetAllScores(_ context.Context, seminarID uuid.UUID) ([]*entity.SeminarScore, error) {
	return f.scores[seminarID], nil
}

func (f *fakeSeminarRepo) HasExaminerScored(_ context.Context, seminarID, examinerID uuid.UUID) (bool, error) {
	for _, s := range f.scores[seminarID] {
		if s.ExaminerID == examinerID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeSeminarRepo) CountDistinctScoredExaminers(_ context.Context, seminarID uuid.UUID) (int, error) {
	seen := map[uuid.UUID]bool{}
	for _, s := range f.scores[seminarID] {
		seen[s.ExaminerID] = true
	}
	return len(seen), nil
}

func (f *fakeSeminarRepo) CheckScheduleConflict(_ context.Context, room string, _ time.Time, examinerIDs []uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	for _, s := range f.scheduled {
		if excludeID != nil && s.ID == *excludeID {
			continue
		}
		if s.Room != nil && *s.Room == room {
			return true, nil
		}
	}
	for _, s := range f.scheduled {
		if excludeID != nil && s.ID == *excludeID {
			continue
		}
		for _, eid := range examinerIDs {
			for _, assigned := range f.examiners[s.ID] {
				if assigned == eid {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// registerExaminerUser stores a dosen_penguji user so role validation passes.
func (f *fakeSeminarRepo) registerExaminerUser(u *entity.User) {
	f.examinerUser[u.ID] = u
}

// recordingSeminarEmailService records seminar email recipients for assertions.
type recordingSeminarEmailService struct {
	submittedTo chan []string
	scheduledTo chan []string
	finalizedTo chan string
}

func newRecordingSeminarEmailService() *recordingSeminarEmailService {
	return &recordingSeminarEmailService{
		submittedTo: make(chan []string, 16),
		scheduledTo: make(chan []string, 16),
		finalizedTo: make(chan string, 16),
	}
}

func (r *recordingSeminarEmailService) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (r *recordingSeminarEmailService) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
func (r *recordingSeminarEmailService) SendThesisSubmitted(context.Context, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingSeminarEmailService) SendThesisApproved(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingSeminarEmailService) SendThesisRejected(context.Context, string, *entity.Thesis, string) error {
	return nil
}
func (r *recordingSeminarEmailService) SendSupervisorAssigned(context.Context, string, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingSeminarEmailService) SendConsultationCreated(context.Context, []string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingSeminarEmailService) SendConsultationApproved(context.Context, string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingSeminarEmailService) SendDocumentUploaded(context.Context, []string, *entity.Document) error {
	return nil
}
func (r *recordingSeminarEmailService) SendDocumentReviewed(context.Context, string, *entity.Document, string) error {
	return nil
}
func (r *recordingSeminarEmailService) SendSeminarSubmitted(_ context.Context, to []string, _ *entity.Seminar) error {
	r.submittedTo <- to
	return nil
}
func (r *recordingSeminarEmailService) SendSeminarScheduled(_ context.Context, to []string, _ *entity.Seminar) error {
	r.scheduledTo <- to
	return nil
}
func (r *recordingSeminarEmailService) SendSeminarFinalized(_ context.Context, to string, _ *entity.Seminar) error {
	r.finalizedTo <- to
	return nil
}
func (r *recordingSeminarEmailService) SendDefenseSubmitted(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingSeminarEmailService) SendDefenseScheduled(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingSeminarEmailService) SendDefenseFinalized(context.Context, string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingSeminarEmailService) SendGraduated(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingSeminarEmailService) SendArchiveCreated(context.Context, string, *entity.ThesisArchive) error {
	return nil
}

func (r *recordingSeminarEmailService) SendTitleChangeRequested(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingSeminarEmailService) SendTitleChangeCancelled(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingSeminarEmailService) SendTitleChangeApproved(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingSeminarEmailService) SendTitleChangeRejected(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

// newTestSeminarUseCase wires a fresh seminar use case with in-memory fakes.
func newTestSeminarUseCase(t *testing.T) (*SeminarUseCase, *fakeSeminarRepo, *fakeDocumentRepo, *fakeThesisRepo, *fakeUserRepo, *recordingSeminarEmailService) {
	t.Helper()
	semRepo := newFakeSeminarRepo()
	docRepo := newFakeDocumentRepo()
	thesisRepo := newFakeThesisRepo()
	userRepo := newFakeUserRepo()
	email := newRecordingSeminarEmailService()

	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	userRepo.roles["dosen_pembimbing"] = &entity.Role{ID: 4, Name: "dosen_pembimbing"}
	userRepo.roles["dosen_penguji"] = &entity.Role{ID: 5, Name: "dosen_penguji"}
	userRepo.roles["kaprodi"] = &entity.Role{ID: 2, Name: "kaprodi"}

	auditSvc := audit.NewAuditService(nil)
	documentUC := NewDocumentUseCase(docRepo, thesisRepo, nil, email, auditSvc)
	uc := NewSeminarUseCase(semRepo, thesisRepo, userRepo, documentUC, email, auditSvc)
	return uc, semRepo, docRepo, thesisRepo, userRepo, email
}

// seedSeminarReadyThesis builds an in_progress thesis with an approved
// seminar_doc so the seminar gate passes; returns thesisID, studentID, supervisorID.
func seedSeminarReadyThesis(t *testing.T, semRepo *fakeSeminarRepo, docRepo *fakeDocumentRepo, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)
	// Register the thesis→student mapping so FindAll role-scoping works.
	semRepo.thesisStudents[thesisID] = studentID
	// Simulate an approved seminar_doc via the doc repo's approvedTypes map.
	docRepo.approvedTypes[thesisID.String()+":"+entity.DocTypeSeminarDoc] = true
	return thesisID, studentID, supervisorID
}

// seedExaminers registers N active dosen_penguji and returns their IDs.
func seedExaminers(t *testing.T, userRepo *fakeUserRepo, semRepo *fakeSeminarRepo, n int) []uuid.UUID {
	t.Helper()
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		u := &entity.User{
			// Unique email so multiple calls (e.g. adding a non-examiner) never collide.
			Email:    "penguji-" + uuid.New().String()[:8] + "@example.com",
			FullName: "Penguji " + string(rune('A'+i)),
			Role:     entity.Role{Name: "dosen_penguji"},
			RoleID:   5,
			IsActive: true,
		}
		if err := userRepo.Create(context.Background(), u); err != nil {
			t.Fatalf("seed penguji: %v", err)
		}
		ids = append(ids, u.ID)
		semRepo.registerExaminerUser(u)
	}
	return ids
}

func futureDays(days int) time.Time {
	return time.Now().Add(time.Duration(days) * 24 * time.Hour).Truncate(time.Second).UTC()
}

func TestSubmitSeminarGatePasses(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)

	detail, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if detail.Status != SeminarStatusPending {
		t.Errorf("status = %q, want pending", detail.Status)
	}
	if len(semRepo.seminars) != 1 {
		t.Errorf("expected 1 seminar, got %d", len(semRepo.seminars))
	}
	// Thesis moved to seminar_ready.
	th := thesisRepo.theses[thesisID]
	if th.Status != "seminar_ready" {
		t.Errorf("thesis status = %q, want seminar_ready", th.Status)
	}
}

func TestSubmitSeminarGateNotMet(t *testing.T) {
	uc, _, _, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	// No approved seminar_doc → gate fails.
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	_, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if !errors.Is(err, ErrSeminarGateNotMet) {
		t.Errorf("expected ErrSeminarGateNotMet, got %v", err)
	}
}

func TestSubmitSeminarForbiddenForOtherStudent(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, _, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	other := seedStudent(t, userRepo, "mahasiswa")

	_, err := uc.Submit(context.Background(), thesisID, Actor{UserID: other})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestSubmitSeminarActiveExists(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)

	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	_, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if !errors.Is(err, ErrSeminarActiveExists) {
		t.Errorf("expected ErrSeminarActiveExists, got %v", err)
	}
}

func TestScheduleSeminar(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)

	detail, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang Sidang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Schedule returned error: %v", err)
	}
	if detail.Status != SeminarStatusScheduled {
		t.Errorf("status = %q, want scheduled", detail.Status)
	}
	if len(semRepo.examiners[seminarID]) != 2 {
		t.Errorf("expected 2 examiners, got %d", len(semRepo.examiners[seminarID]))
	}
}

func TestScheduleSeminarLeadTimeTooShort(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)

	_, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(1), // < 3 days
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrSeminarScheduleLeadTime) {
		t.Errorf("expected ErrSeminarScheduleLeadTime, got %v", err)
	}
}

func TestScheduleSeminarMinExaminers(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)

	_, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners[:1], // only 1
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrSeminarMinExaminers) {
		t.Errorf("expected ErrSeminarMinExaminers, got %v", err)
	}
}

func TestScheduleSeminarRoomConflict(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)

	// Mark an existing seminar occupying "Ruang Sidang A".
	existing := &entity.Seminar{ID: uuid.New(), Status: SeminarStatusScheduled}
	room := "Ruang Sidang A"
	existing.Room = &room
	semRepo.scheduled = append(semRepo.scheduled, *existing)

	_, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang Sidang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrSeminarRoomConflict) {
		t.Errorf("expected ErrSeminarRoomConflict, got %v", err)
	}
}

func TestSubmitScoresAndFinalize(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, email := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)
	if _, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Examiner 1 submits 80 across all components.
	result1, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 80),
	}, Actor{UserID: examiners[0]})
	if err != nil {
		t.Fatalf("submit scores #1: %v", err)
	}
	if result1.IsComplete {
		t.Error("result should not be complete after one examiner")
	}

	// Examiner 2 submits 90 → finalize triggered.
	result2, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 90),
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
	if result2.Status != SeminarStatusPassed {
		t.Errorf("status = %q, want passed", result2.Status)
	}

	// Thesis moved to seminar_done.
	th := thesisRepo.theses[thesisID]
	if th.Status != "seminar_done" {
		t.Errorf("thesis status = %q, want seminar_done", th.Status)
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

func TestSubmitScoresFailedSeminar(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)
	if _, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Both examiners score 50 → final = 50 < 60 → failed.
	if _, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 50),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("scores #1: %v", err)
	}
	result, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 50),
	}, Actor{UserID: examiners[1]})
	if err != nil {
		t.Fatalf("scores #2: %v", err)
	}
	if result.Status != SeminarStatusFailed {
		t.Errorf("status = %q, want failed", result.Status)
	}
	// Thesis stays seminar_ready on failure.
	th := thesisRepo.theses[thesisID]
	if th.Status != "seminar_ready" {
		t.Errorf("thesis status = %q, want seminar_ready (kept on failure)", th.Status)
	}
}

func TestSubmitScoresByNonExaminer(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)
	if _, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// A third dosen_penguji who was NOT assigned.
	other := seedExaminers(t, userRepo, semRepo, 1)[0]

	_, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 80),
	}, Actor{UserID: other})
	if !errors.Is(err, ErrSeminarNotExaminer) {
		t.Errorf("expected ErrSeminarNotExaminer, got %v", err)
	}
}

func TestSubmitScoresAlreadyScored(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)
	if _, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	if _, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 80),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	_, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 85),
	}, Actor{UserID: examiners[0]})
	if !errors.Is(err, ErrSeminarAlreadyScored) {
		t.Errorf("expected ErrSeminarAlreadyScored, got %v", err)
	}
}

func TestSubmitScoresIncompleteComponents(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)
	if _, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Only one component supplied.
	_, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: []ComponentScoreInput{{ComponentName: "Presentasi", Score: 80}},
	}, Actor{UserID: examiners[0]})
	if !errors.Is(err, ErrSeminarIncompleteScore) {
		t.Errorf("expected ErrSeminarIncompleteScore, got %v", err)
	}
}

func TestResultBreakdown(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)
	if _, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Examiner A scores 80 in Presentasi (30%), 80 elsewhere → 80.0.
	if _, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 80),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("scores #1: %v", err)
	}

	// Admin sees a partial breakdown with 1 examiner scored.
	result, err := uc.Result(context.Background(), seminarID, uuid.New(), ThesisRoleAdminFakultas)
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

func TestSetRevisionNotes(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]
	examiners := seedExaminers(t, userRepo, semRepo, 2)
	if _, err := uc.Schedule(context.Background(), seminarID, ScheduleSeminarRequest{
		ScheduledAt: futureDays(5),
		Room:        "Ruang A",
		ExaminerIDs: examiners,
	}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if _, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 80),
	}, Actor{UserID: examiners[0]}); err != nil {
		t.Fatalf("scores #1: %v", err)
	}
	if _, err := uc.SubmitScores(context.Background(), seminarID, SubmitScoreRequest{
		Scores: fullScoreInput(entity.SeminarGradingComponents, 90),
	}, Actor{UserID: examiners[1]}); err != nil {
		t.Fatalf("scores #2: %v", err)
	}

	detail, err := uc.SetRevisionNotes(context.Background(), seminarID, "Perbaiki rumusan masalah di BAB 1", Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("SetRevisionNotes: %v", err)
	}
	if detail.Notes == nil || *detail.Notes != "Perbaiki rumusan masalah di BAB 1" {
		t.Errorf("notes not persisted in response")
	}
	if semRepo.notes[seminarID] != "Perbaiki rumusan masalah di BAB 1" {
		t.Errorf("notes not persisted in repo")
	}
}

func TestSetRevisionNotesBeforePassedRejected(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	seminarID := semRepo.byThesis[thesisID]

	_, err := uc.SetRevisionNotes(context.Background(), seminarID, "catatan", Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrSeminarNotPassed) {
		t.Errorf("expected ErrSeminarNotPassed, got %v", err)
	}
}

func TestSeminarListScopedByRole(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)
	if _, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Owner sees their seminar.
	list, total, err := uc.List(context.Background(), repository.SeminarFilter{}, studentID, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 seminar for owner, got total=%d len=%d", total, len(list))
	}

	// An unrelated student sees nothing (role scoping actually filters).
	other := seedStudent(t, userRepo, "mahasiswa")
	_, totalOther, err := uc.List(context.Background(), repository.SeminarFilter{}, other, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("List for other: %v", err)
	}
	if totalOther != 0 {
		t.Errorf("expected 0 seminars for unrelated student, got %d", totalOther)
	}
}

// fullScoreInput builds a score input for every component with the given value.
func fullScoreInput(components []entity.GradingComponent, score float64) []ComponentScoreInput {
	out := make([]ComponentScoreInput, 0, len(components))
	for _, c := range components {
		out = append(out, ComponentScoreInput{ComponentName: c.Name, Score: score})
	}
	return out
}
