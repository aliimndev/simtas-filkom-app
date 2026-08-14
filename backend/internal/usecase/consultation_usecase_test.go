package usecase

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
)

// fakeConsultationRepo is a minimal in-memory ConsultationRepository for usecase tests.
type fakeConsultationRepo struct {
	logs        map[uuid.UUID]*entity.ConsultationLog
	byThesis    map[uuid.UUID][]uuid.UUID // thesisID -> ordered log IDs
	approvedCnt map[uuid.UUID]int
	deleted     []uuid.UUID
}

func newFakeConsultationRepo() *fakeConsultationRepo {
	return &fakeConsultationRepo{
		logs:        map[uuid.UUID]*entity.ConsultationLog{},
		byThesis:    map[uuid.UUID][]uuid.UUID{},
		approvedCnt: map[uuid.UUID]int{},
	}
}

func (f *fakeConsultationRepo) Create(_ context.Context, log *entity.ConsultationLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now()
	log.UpdatedAt = time.Now()
	f.logs[log.ID] = log
	f.byThesis[log.ThesisID] = append(f.byThesis[log.ThesisID], log.ID)
	return nil
}

func (f *fakeConsultationRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.ConsultationLog, error) {
	l, ok := f.logs[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *l
	return &clone, nil
}

func (f *fakeConsultationRepo) FindByThesisID(_ context.Context, thesisID uuid.UUID, filter domainRepo.ConsultationFilter) ([]*entity.ConsultationLog, int64, error) {
	var out []*entity.ConsultationLog
	for _, id := range f.byThesis[thesisID] {
		l := f.logs[id]
		if filter.Status != "" && l.Status != filter.Status {
			continue
		}
		if filter.DateFrom != "" {
			from, _ := time.Parse("2006-01-02", filter.DateFrom)
			if l.ConsultationDate.Before(from) {
				continue
			}
		}
		if filter.DateTo != "" {
			to, _ := time.Parse("2006-01-02", filter.DateTo)
			if l.ConsultationDate.After(to) {
				continue
			}
		}
		clone := *l
		out = append(out, &clone)
	}
	// Newest first (mirror real repo: ORDER BY consultation_date DESC, created_at DESC).
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].ConsultationDate.Equal(out[j].ConsultationDate) {
			return out[i].ConsultationDate.After(out[j].ConsultationDate)
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	total := int64(len(out))
	// PerPage == 0 means "no pagination" (mirror real repo, used by summary).
	if filter.PerPage <= 0 {
		return out, total, nil
	}
	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	start := (page - 1) * perPage
	if start > len(out) {
		start = len(out)
	}
	end := start + perPage
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], total, nil
}

func (f *fakeConsultationRepo) Update(_ context.Context, log *entity.ConsultationLog) error {
	if _, ok := f.logs[log.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	log.UpdatedAt = time.Now()
	f.logs[log.ID] = log
	return nil
}

func (f *fakeConsultationRepo) Approve(_ context.Context, id uuid.UUID, approvedBy uuid.UUID) error {
	l, ok := f.logs[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if l.Status != "pending" {
		return errors.New("status must be pending")
	}
	l.Status = "approved"
	l.ApprovedBy = &approvedBy
	now := time.Now()
	l.ApprovedAt = &now
	f.approvedCnt[l.ThesisID]++
	return nil
}

func (f *fakeConsultationRepo) CountApprovedByThesisID(_ context.Context, thesisID uuid.UUID) (int, error) {
	return f.approvedCnt[thesisID], nil
}

func (f *fakeConsultationRepo) Delete(_ context.Context, id uuid.UUID) error {
	l, ok := f.logs[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(f.logs, id)
	ids := f.byThesis[l.ThesisID]
	for i, x := range ids {
		if x == id {
			f.byThesis[l.ThesisID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	f.deleted = append(f.deleted, id)
	return nil
}

// recordingEmailService records consultation email recipients for assertions.
// Channel-backed so tests can deterministically wait for the (async) email.
type recordingEmailService struct {
	createdTo  chan []string
	approvedTo chan string
}

func newRecordingEmailService() *recordingEmailService {
	return &recordingEmailService{
		createdTo:  make(chan []string, 16),
		approvedTo: make(chan string, 16),
	}
}

func (r *recordingEmailService) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (r *recordingEmailService) SendPasswordResetLink(context.Context, string, string, string) error {
	return nil
}
func (r *recordingEmailService) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
func (r *recordingEmailService) SendThesisSubmitted(context.Context, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingEmailService) SendThesisApproved(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingEmailService) SendThesisRejected(context.Context, string, *entity.Thesis, string) error {
	return nil
}
func (r *recordingEmailService) SendSupervisorAssigned(context.Context, string, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingEmailService) SendConsultationCreated(_ context.Context, to []string, _ *entity.ConsultationLog) error {
	r.createdTo <- to
	return nil
}
func (r *recordingEmailService) SendConsultationApproved(_ context.Context, to string, _ *entity.ConsultationLog) error {
	r.approvedTo <- to
	return nil
}
func (r *recordingEmailService) SendDocumentUploaded(context.Context, []string, *entity.Document) error {
	return nil
}
func (r *recordingEmailService) SendDocumentReviewed(context.Context, string, *entity.Document, string) error {
	return nil
}
func (r *recordingEmailService) SendSeminarSubmitted(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingEmailService) SendSeminarScheduled(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingEmailService) SendSeminarFinalized(context.Context, string, *entity.Seminar) error {
	return nil
}
func (r *recordingEmailService) SendDefenseSubmitted(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingEmailService) SendDefenseScheduled(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingEmailService) SendDefenseFinalized(context.Context, string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingEmailService) SendGraduated(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingEmailService) SendArchiveCreated(context.Context, string, *entity.ThesisArchive) error {
	return nil
}

func (r *recordingEmailService) SendTitleChangeRequested(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingEmailService) SendTitleChangeCancelled(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingEmailService) SendTitleChangeApproved(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingEmailService) SendTitleChangeRejected(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

// waitCreated blocks until one consultation-created email arrives (with timeout).
func (r *recordingEmailService) waitCreated(t *testing.T) []string {
	t.Helper()
	select {
	case to := <-r.createdTo:
		return to
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for consultation-created email")
		return nil
	}
}

// waitApproved blocks until one consultation-approved email arrives (with timeout).
func (r *recordingEmailService) waitApproved(t *testing.T) string {
	t.Helper()
	select {
	case to := <-r.approvedTo:
		return to
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for consultation-approved email")
		return ""
	}
}

// newTestConsultationUseCase wires a fresh use case with in-memory fakes.
func newTestConsultationUseCase() (*ConsultationUseCase, *fakeConsultationRepo, *fakeThesisRepo, *fakeUserRepo, *recordingEmailService) {
	consultRepo := newFakeConsultationRepo()
	thesisRepo := newFakeThesisRepo()
	userRepo := newFakeUserRepo()
	email := newRecordingEmailService()

	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	userRepo.roles["dosen_pembimbing"] = &entity.Role{ID: 4, Name: "dosen_pembimbing"}
	userRepo.roles["dosen_penguji"] = &entity.Role{ID: 5, Name: "dosen_penguji"}

	auditSvc := audit.NewAuditService(nil)
	uc := NewConsultationUseCase(consultRepo, thesisRepo, email, auditSvc, nil)
	return uc, consultRepo, thesisRepo, userRepo, email
}

// seedInProgressThesis builds a thesis in `in_progress` state owned by a student
// and supervised by a dosen; returns thesisID, studentID, supervisorID.
func seedInProgressThesis(t *testing.T, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	supervisor := &entity.User{
		Email:    "supervisor@example.com",
		FullName: "Dosen Pembimbing Satu",
		Role:     entity.Role{Name: "dosen_pembimbing"},
		RoleID:   4,
		IsActive: true,
	}
	if err := userRepo.Create(context.Background(), supervisor); err != nil {
		t.Fatalf("seed supervisor: %v", err)
	}
	registerFakeStudent(thesisRepo, studentID, userRepo)

	thesis := &entity.Thesis{
		StudentID:      studentID,
		AcademicYearID: uuid.New(),
		Title:          validTitle(),
		ThesisType:     "skripsi",
		Status:         "in_progress",
		SubmittedAt:    time.Now(),
	}
	if err := thesisRepo.Create(context.Background(), thesis); err != nil {
		t.Fatalf("seed thesis: %v", err)
	}
	thesisRepo.supervisors[thesis.ID] = []uuid.UUID{supervisor.ID}
	thesisRepo.supUsers[thesis.ID] = []*entity.User{supervisor}
	return thesis.ID, studentID, supervisor.ID
}

// seedOtherDosen creates a dosen_pembimbing who is NOT assigned to the thesis.
func seedOtherDosen(t *testing.T, userRepo *fakeUserRepo) uuid.UUID {
	t.Helper()
	d := &entity.User{
		Email:    "other-dosen@example.com",
		FullName: "Dosen Lain",
		Role:     entity.Role{Name: "dosen_pembimbing"},
		RoleID:   4,
		IsActive: true,
	}
	if err := userRepo.Create(context.Background(), d); err != nil {
		t.Fatalf("seed other dosen: %v", err)
	}
	return d.ID
}

func todayStr() string {
	return time.Now().Format("2006-01-02")
}

func TestCreateConsultationByOwner(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, email := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	detail, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Pembahasan BAB 2 tinjauan pustaka",
		Notes:            strPtr("Perlu tambahkan referensi terbaru"),
		FollowUp:         strPtr("Upload draft revisi minggu depan"),
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if detail.Status != ConsultationStatusPending {
		t.Errorf("status = %q, want pending", detail.Status)
	}
	if detail.CreatedBy != studentID {
		t.Errorf("created_by = %v, want student", detail.CreatedBy)
	}
	if len(consultRepo.logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(consultRepo.logs))
	}
	// Student created → email goes to the supervisor.
	to := email.waitCreated(t)
	if len(to) != 1 || to[0] != "supervisor@example.com" {
		t.Errorf("email recipients = %v, want [supervisor@example.com]", to)
	}
	_ = supervisorID
}

func TestCreateConsultationBySupervisor(t *testing.T) {
	uc, _, thesisRepo, userRepo, email := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	_, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Pembahasan progress bab 3",
	}, Actor{UserID: supervisorID})
	if err != nil {
		t.Fatalf("Create by supervisor returned error: %v", err)
	}
	// Supervisor created → email goes to the student.
	to := email.waitCreated(t)
	if len(to) != 1 {
		t.Fatalf("expected student email, got %v", to)
	}
	student := userRepo.users[studentID]
	if to[0] != student.Email {
		t.Errorf("email recipient = %q, want %q", to[0], student.Email)
	}
}

func TestCreateConsultationForbiddenForOtherStudent(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	other := seedStudent(t, userRepo, "mahasiswa")

	_, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Pembahasan",
	}, Actor{UserID: other})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateConsultationFutureDateRejected(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	future := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	_, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: future,
		TopicsDiscussed:  "Pembahasan",
	}, Actor{UserID: studentID})
	if !errors.Is(err, ErrConsultationDateFuture) {
		t.Errorf("expected ErrConsultationDateFuture, got %v", err)
	}
}

func TestCreateConsultationEmptyTopicsRejected(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	_, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "   ",
	}, Actor{UserID: studentID})
	if !errors.Is(err, ErrTopicsDiscussedRequired) {
		t.Errorf("expected ErrTopicsDiscussedRequired, got %v", err)
	}
}

func TestCreateConsultationThesisNotInProgress(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	registerFakeStudent(thesisRepo, studentID, userRepo)

	thesis := &entity.Thesis{
		StudentID:      studentID,
		AcademicYearID: uuid.New(),
		Title:          validTitle(),
		ThesisType:     "skripsi",
		Status:         "submitted", // not eligible
		SubmittedAt:    time.Now(),
	}
	if err := thesisRepo.Create(context.Background(), thesis); err != nil {
		t.Fatalf("seed thesis: %v", err)
	}

	_, err := uc.Create(context.Background(), thesis.ID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Pembahasan",
	}, Actor{UserID: studentID})
	if !errors.Is(err, ErrThesisNotInProgress) {
		t.Errorf("expected ErrThesisNotInProgress, got %v", err)
	}
}

func TestCreateConsultationThesisNotFound(t *testing.T) {
	uc, _, _, _, _ := newTestConsultationUseCase()
	_, err := uc.Create(context.Background(), uuid.New(), CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Pembahasan",
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrThesisNotFound) {
		t.Errorf("expected ErrThesisNotFound, got %v", err)
	}
}

func TestListConsultationsWithSummary(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	// Two logs by the student, one approved by the supervisor.
	for i := 0; i < 2; i++ {
		if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
			ConsultationDate: time.Now().AddDate(0, 0, -i).Format("2006-01-02"),
			TopicsDiscussed:  "Topik " + string(rune('A'+i)),
		}, Actor{UserID: studentID}); err != nil {
			t.Fatalf("seed log: %v", err)
		}
	}
	// Approve the most recent one.
	ids := consultRepo.byThesis[thesisID]
	if _, err := uc.Approve(context.Background(), thesisID, ids[0], supervisorID, Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	result, total, err := uc.List(context.Background(), thesisID, domainRepo.ConsultationFilter{}, studentID, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 2 || len(result.Consultations) != 2 {
		t.Errorf("expected 2 logs, got total=%d len=%d", total, len(result.Consultations))
	}
	if result.Summary.TotalConsultations != 2 {
		t.Errorf("summary total = %d, want 2", result.Summary.TotalConsultations)
	}
	if result.Summary.ApprovedCount != 1 || result.Summary.PendingCount != 1 {
		t.Errorf("summary approved/pending = %d/%d, want 1/1", result.Summary.ApprovedCount, result.Summary.PendingCount)
	}
	if result.Summary.LastConsultationDate == nil {
		t.Error("last_consultation_date should be set")
	}
}

func TestListConsultationsForbiddenForOtherStudent(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	other := seedStudent(t, userRepo, "mahasiswa")

	_, _, err := uc.List(context.Background(), thesisID, domainRepo.ConsultationFilter{}, other, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGetConsultation(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Detail pembahasan",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	detail, err := uc.GetByID(context.Background(), thesisID, id, studentID, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if detail.ID != id {
		t.Errorf("wrong log returned")
	}
}

func TestGetConsultationNotFound(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	_, err := uc.GetByID(context.Background(), thesisID, uuid.New(), studentID, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrConsultationNotFound) {
		t.Errorf("expected ErrConsultationNotFound, got %v", err)
	}
}

func TestUpdateConsultationWhilePending(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Sebelum",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	detail, err := uc.Update(context.Background(), thesisID, id, UpdateConsultationRequest{
		TopicsDiscussed: strPtr("Sesudah revisi"),
		FollowUp:        strPtr("Kumpulkan minggu depan"),
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if detail.TopicsDiscussed != "Sesudah revisi" {
		t.Errorf("topics = %q, want updated value", detail.TopicsDiscussed)
	}
	if detail.FollowUp == nil || *detail.FollowUp != "Kumpulkan minggu depan" {
		t.Errorf("follow_up not updated")
	}
}

func TestUpdateConsultationByNonCreator(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Sebelum",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	// Supervisor is not the creator → cannot update.
	_, err := uc.Update(context.Background(), thesisID, id, UpdateConsultationRequest{
		TopicsDiscussed: strPtr("Changed"),
	}, Actor{UserID: supervisorID})
	if !errors.Is(err, ErrNotConsultationCreator) {
		t.Errorf("expected ErrNotConsultationCreator, got %v", err)
	}
}

func TestUpdateConsultationAfterApproved(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Sebelum",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]
	if _, err := uc.Approve(context.Background(), thesisID, id, supervisorID, Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	_, err := uc.Update(context.Background(), thesisID, id, UpdateConsultationRequest{
		TopicsDiscussed: strPtr("Changed"),
	}, Actor{UserID: studentID})
	if !errors.Is(err, ErrConsultationAlreadyDone) {
		t.Errorf("expected ErrConsultationAlreadyDone, got %v", err)
	}
}

func TestApproveConsultationBySupervisor(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, email := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Bimbingan bab 2",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	detail, err := uc.Approve(context.Background(), thesisID, id, supervisorID, Actor{UserID: supervisorID})
	if err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}
	if detail.Status != ConsultationStatusApproved {
		t.Errorf("status = %q, want approved", detail.Status)
	}
	if detail.ApprovedBy == nil || *detail.ApprovedBy != supervisorID {
		t.Errorf("approved_by not set correctly")
	}
	if detail.ApprovedAt == nil {
		t.Error("approved_at should be set")
	}
	// Supervisor approved → email goes to the student.
	studentEmail := email.waitApproved(t)
	student := userRepo.users[studentID]
	if studentEmail != student.Email {
		t.Errorf("approval email to = %q, want %q", studentEmail, student.Email)
	}
	count, err := consultRepo.CountApprovedByThesisID(context.Background(), thesisID)
	if err != nil || count != 1 {
		t.Errorf("approved count = %d (err=%v), want 1", count, err)
	}
}

func TestApproveConsultationByNonSupervisor(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	otherDosen := seedOtherDosen(t, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Bimbingan",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	_, err := uc.Approve(context.Background(), thesisID, id, otherDosen, Actor{UserID: otherDosen})
	if !errors.Is(err, ErrNotSupervisorOfThesis) {
		t.Errorf("expected ErrNotSupervisorOfThesis, got %v", err)
	}
}

func TestApproveConsultationAlreadyApproved(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Bimbingan",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]
	if _, err := uc.Approve(context.Background(), thesisID, id, supervisorID, Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("first approve: %v", err)
	}

	_, err := uc.Approve(context.Background(), thesisID, id, supervisorID, Actor{UserID: supervisorID})
	if !errors.Is(err, ErrConsultationAlreadyDone) {
		t.Errorf("expected ErrConsultationAlreadyDone, got %v", err)
	}
}

func TestCountApprovedByThesisID(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	for i := 0; i < 3; i++ {
		if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
			ConsultationDate: time.Now().AddDate(0, 0, -i).Format("2006-01-02"),
			TopicsDiscussed:  "Topik",
		}, Actor{UserID: studentID}); err != nil {
			t.Fatalf("seed log: %v", err)
		}
	}
	for _, id := range consultRepo.byThesis[thesisID] {
		if _, err := uc.Approve(context.Background(), thesisID, id, supervisorID, Actor{UserID: supervisorID}); err != nil {
			t.Fatalf("approve: %v", err)
		}
	}
	count, err := consultRepo.CountApprovedByThesisID(context.Background(), thesisID)
	if err != nil || count != 3 {
		t.Errorf("approved count = %d (err=%v), want 3", count, err)
	}
}

func TestDeleteConsultationWhilePending(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Akan dihapus",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	if err := uc.Delete(context.Background(), thesisID, id, studentID, Actor{UserID: studentID}); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if len(consultRepo.logs) != 0 {
		t.Errorf("expected log to be deleted")
	}
}

func TestDeleteConsultationAfterApprovedRejected(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Tidak bisa dihapus",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]
	if _, err := uc.Approve(context.Background(), thesisID, id, supervisorID, Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	err := uc.Delete(context.Background(), thesisID, id, studentID, Actor{UserID: studentID})
	if !errors.Is(err, ErrConsultationAlreadyDone) {
		t.Errorf("expected ErrConsultationAlreadyDone, got %v", err)
	}
}

func TestDeleteConsultationByNonCreator(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Bimbingan",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	err := uc.Delete(context.Background(), thesisID, id, supervisorID, Actor{UserID: supervisorID})
	if !errors.Is(err, ErrNotConsultationCreator) {
		t.Errorf("expected ErrNotConsultationCreator, got %v", err)
	}
}

func TestSummaryEndpoint(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	// Two logs with a 7-day gap.
	for i := 0; i < 2; i++ {
		if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
			ConsultationDate: time.Now().AddDate(0, 0, -7*i).Format("2006-01-02"),
			TopicsDiscussed:  "Topik",
		}, Actor{UserID: studentID}); err != nil {
			t.Fatalf("seed log: %v", err)
		}
	}
	if _, err := uc.Approve(context.Background(), thesisID, consultRepo.byThesis[thesisID][0], supervisorID, Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("approve: %v", err)
	}

	summary, err := uc.Summary(context.Background(), thesisID, studentID, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if summary.TotalConsultations != 2 {
		t.Errorf("total = %d, want 2", summary.TotalConsultations)
	}
	if summary.ApprovedCount != 1 || summary.PendingCount != 1 {
		t.Errorf("approved/pending = %d/%d, want 1/1", summary.ApprovedCount, summary.PendingCount)
	}
	if summary.AverageIntervalDays == nil || *summary.AverageIntervalDays != 7 {
		t.Errorf("average interval = %v, want 7", summary.AverageIntervalDays)
	}
	if summary.LastConsultationDate == nil {
		t.Error("last_consultation_date should be set")
	}
}

func TestSummaryAdminAccess(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Topik",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}

	// Admin + Kaprodi can view without being linked to the thesis.
	summary, err := uc.Summary(context.Background(), thesisID, uuid.New(), ThesisRoleAdminFakultas)
	if err != nil {
		t.Fatalf("Summary for admin returned error: %v", err)
	}
	if summary.TotalConsultations != 1 {
		t.Errorf("total = %d, want 1", summary.TotalConsultations)
	}
	_ = consultRepo
}

func strPtr(s string) *string {
	return &s
}
