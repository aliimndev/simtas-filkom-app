package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
)

// fakeTitleChangeRequestRepo is a minimal in-memory TitleChangeRequestRepository.
type fakeTitleChangeRequestRepo struct {
	reqs       map[uuid.UUID]*entity.TitleChangeRequest
	byThesis   map[uuid.UUID][]uuid.UUID
	thesisRepo *fakeThesisRepo // optional, used to mirror the real repo's Thesis preloads
}

func newFakeTitleChangeRequestRepo() *fakeTitleChangeRequestRepo {
	return &fakeTitleChangeRequestRepo{
		reqs:     map[uuid.UUID]*entity.TitleChangeRequest{},
		byThesis: map[uuid.UUID][]uuid.UUID{},
	}
}

func (f *fakeTitleChangeRequestRepo) Create(_ context.Context, r *entity.TitleChangeRequest) error {
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	f.reqs[r.ID] = r
	f.byThesis[r.ThesisID] = append(f.byThesis[r.ThesisID], r.ID)
	return nil
}

func (f *fakeTitleChangeRequestRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.TitleChangeRequest, error) {
	r, ok := f.reqs[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	c := *r
	// Mirror the real repo's Thesis preloads so supervisor checks work.
	if f.thesisRepo != nil {
		if t, err := f.thesisRepo.FindByID(ctx, c.ThesisID); err == nil {
			c.Thesis = *t
		}
	}
	// Mirror the real repo's user preloads (RequestedBy/ReviewedBy/CancelledBy).
	if c.RequestedByID != uuid.Nil && c.RequestedBy == nil {
		c.RequestedBy = &entity.User{ID: c.RequestedByID, FullName: "Mahasiswa Satu"}
	}
	if c.ReviewedByID != nil && c.ReviewedBy == nil {
		c.ReviewedBy = &entity.User{ID: *c.ReviewedByID, FullName: "Dosen Pembimbing"}
	}
	if c.CancelledByID != nil && c.CancelledBy == nil {
		c.CancelledBy = &entity.User{ID: *c.CancelledByID, FullName: "Mahasiswa Satu"}
	}
	return &c, nil
}

func (f *fakeTitleChangeRequestRepo) FindByThesisID(_ context.Context, thesisID uuid.UUID) ([]*entity.TitleChangeRequest, error) {
	var out []*entity.TitleChangeRequest
	for _, id := range f.byThesis[thesisID] {
		c := *f.reqs[id]
		out = append(out, &c)
	}
	return out, nil
}

func (f *fakeTitleChangeRequestRepo) FindPendingByThesisID(_ context.Context, thesisID uuid.UUID) (*entity.TitleChangeRequest, error) {
	for _, id := range f.byThesis[thesisID] {
		if f.reqs[id].Status == TitleChangeStatusPending {
			c := *f.reqs[id]
			return &c, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeTitleChangeRequestRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string, reviewedByID *uuid.UUID, reviewedAt *time.Time, reviewNotes *string, cancelledByID *uuid.UUID, cancelledAt *time.Time) error {
	r, ok := f.reqs[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	// Mirror the real repo's PENDING guard (concurrent transition wins the race).
	if r.Status != TitleChangeStatusPending {
		return gorm.ErrRecordNotFound
	}
	r.Status = status
	if reviewedByID != nil {
		r.ReviewedByID = reviewedByID
	}
	if reviewedAt != nil {
		r.ReviewedAt = reviewedAt
	}
	if reviewNotes != nil {
		r.ReviewNotes = reviewNotes
	}
	if cancelledByID != nil {
		r.CancelledByID = cancelledByID
	}
	if cancelledAt != nil {
		r.CancelledAt = cancelledAt
	}
	r.UpdatedAt = time.Now()
	return nil
}

// FindPendingBySupervisorID mirrors the real repo's review-queue query: it
// returns PENDING requests only for theses supervised by supervisorID.
func (f *fakeTitleChangeRequestRepo) FindPendingBySupervisorID(_ context.Context, supervisorID uuid.UUID) ([]*entity.TitleChangeRequest, error) {
	var out []*entity.TitleChangeRequest
	for _, r := range f.reqs {
		if r.Status != TitleChangeStatusPending {
			continue
		}
		if f.thesisRepo != nil {
			assigned := false
			for _, sid := range f.thesisRepo.supervisors[r.ThesisID] {
				if sid == supervisorID {
					assigned = true
					break
				}
			}
			if !assigned {
				continue
			}
		}
		clone := *r
		out = append(out, &clone)
	}
	return out, nil
}

// Approve mirrors the real repo's atomic approve (request + thesis title).
// It needs the thesis repo to update the title, mirroring the DB transaction.
func (f *fakeTitleChangeRequestRepo) Approve(_ context.Context, id uuid.UUID, reviewedByID uuid.UUID, reviewedAt time.Time, reviewNotes *string, newTitle string) error {
	r, ok := f.reqs[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if r.Status != TitleChangeStatusPending {
		return gorm.ErrRecordNotFound
	}
	r.Status = TitleChangeStatusApproved
	r.ReviewedByID = &reviewedByID
	r.ReviewedAt = &reviewedAt
	r.ReviewNotes = reviewNotes
	r.UpdatedAt = time.Now()
	return nil
}

// tcrRecordingEmail captures title-change email sends. It implements the full
// EmailService interface (other events are no-ops) so it can back the use case.
type tcrRecordingEmail struct {
	requestedTo chan []string
	cancelledTo chan []string
	approvedTo  chan []string
	rejectedTo  chan []string
}

func newTCRRecordingEmail() *tcrRecordingEmail {
	return &tcrRecordingEmail{
		requestedTo: make(chan []string, 8),
		cancelledTo: make(chan []string, 8),
		approvedTo:  make(chan []string, 8),
		rejectedTo:  make(chan []string, 8),
	}
}

func (r *tcrRecordingEmail) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (r *tcrRecordingEmail) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
func (r *tcrRecordingEmail) SendThesisSubmitted(context.Context, []string, *entity.Thesis) error {
	return nil
}
func (r *tcrRecordingEmail) SendThesisApproved(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *tcrRecordingEmail) SendThesisRejected(context.Context, string, *entity.Thesis, string) error {
	return nil
}
func (r *tcrRecordingEmail) SendSupervisorAssigned(context.Context, string, []string, *entity.Thesis) error {
	return nil
}
func (r *tcrRecordingEmail) SendConsultationCreated(context.Context, []string, *entity.ConsultationLog) error {
	return nil
}
func (r *tcrRecordingEmail) SendConsultationApproved(context.Context, string, *entity.ConsultationLog) error {
	return nil
}
func (r *tcrRecordingEmail) SendDocumentUploaded(context.Context, []string, *entity.Document) error {
	return nil
}
func (r *tcrRecordingEmail) SendDocumentReviewed(context.Context, string, *entity.Document, string) error {
	return nil
}
func (r *tcrRecordingEmail) SendSeminarSubmitted(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *tcrRecordingEmail) SendSeminarScheduled(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *tcrRecordingEmail) SendSeminarFinalized(context.Context, string, *entity.Seminar) error {
	return nil
}
func (r *tcrRecordingEmail) SendDefenseSubmitted(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *tcrRecordingEmail) SendDefenseScheduled(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *tcrRecordingEmail) SendDefenseFinalized(context.Context, string, *entity.ThesisDefense) error {
	return nil
}
func (r *tcrRecordingEmail) SendGraduated(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *tcrRecordingEmail) SendArchiveCreated(context.Context, string, *entity.ThesisArchive) error {
	return nil
}
func (r *tcrRecordingEmail) SendTitleChangeRequested(_ context.Context, to []string, _ *entity.Thesis, _ *entity.TitleChangeRequest) error {
	r.requestedTo <- to
	return nil
}
func (r *tcrRecordingEmail) SendTitleChangeCancelled(_ context.Context, to []string, _ *entity.Thesis, _ *entity.TitleChangeRequest) error {
	r.cancelledTo <- to
	return nil
}
func (r *tcrRecordingEmail) SendTitleChangeApproved(_ context.Context, to []string, _ *entity.Thesis, _ *entity.TitleChangeRequest) error {
	r.approvedTo <- to
	return nil
}
func (r *tcrRecordingEmail) SendTitleChangeRejected(_ context.Context, to []string, _ *entity.Thesis, _ *entity.TitleChangeRequest) error {
	r.rejectedTo <- to
	return nil
}

// seedTCRThesis builds an approved/in_progress thesis with one supervisor and
// returns thesisID, studentID, and supervisorID.
func seedTCRThesis(t *testing.T, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo, studentID uuid.UUID, status string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	registerFakeStudent(thesisRepo, studentID, userRepo)

	// Create the thesis directly in the passed fake repo (the real use case
	// would create it in its own fresh repo, which the tests would never see).
	thesis := &entity.Thesis{
		StudentID:      studentID,
		AcademicYearID: uuid.New(),
		Title:          validTitle(),
		ThesisType:     "skripsi",
		Status:         status,
		SubmittedAt:    time.Now(),
	}
	if err := thesisRepo.Create(context.Background(), thesis); err != nil {
		t.Fatalf("seed thesis: %v", err)
	}
	thesisID := thesis.ID

	// Attach a supervisor (with email) to the thesis.
	supervisorID := uuid.New()
	supervisor := &entity.User{ID: supervisorID, Email: "sup@example.com", FullName: "Dosen Pembimbing", RoleID: 4}
	thesisRepo.supUsers[thesisID] = []*entity.User{supervisor}
	thesisRepo.supervisors[thesisID] = []uuid.UUID{supervisorID}

	return thesisID, supervisorID
}

// newTestTCRUseCase wires a title-change use case backed by fakes.
func newTestTCRUseCase(thesisRepo *fakeThesisRepo) (*TitleChangeRequestUseCase, *fakeTitleChangeRequestRepo, *fakeUserRepo, *tcrRecordingEmail) {
	tcrRepo := newFakeTitleChangeRequestRepo()
	tcrRepo.thesisRepo = thesisRepo
	emailRec := newTCRRecordingEmail()
	auditSvc := audit.NewAuditService(nil)
	uc := NewTitleChangeRequestUseCase(tcrRepo, thesisRepo, emailRec, auditSvc)
	userRepo := newFakeUserRepo()
	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	userRepo.roles["dosen_pembimbing"] = &entity.Role{ID: 4, Name: "dosen_pembimbing"}
	return uc, tcrRepo, userRepo, emailRec
}

func newTCRTitle() string {
	return "Sistem Pendukung Keputusan Pemilihan Dosen Pembimbing Menggunakan Metode TOPSIS Dan AHP Berbasis Web"
}

func TestTitleChangeSubmitSuccess(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	thesisID, _ := seedTCRThesis(t, thesisRepo, userRepo, studentID, "approved")

	uc, tcrRepo, _, emailRec := newTestTCRUseCase(thesisRepo)

	reason := "Judul diperbaiki agar lebih spesifik"
	detail, err := uc.Submit(context.Background(), thesisID, CreateTitleChangeRequest{
		RequestedTitle: newTCRTitle(),
		Reason:         &reason,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}
	if detail.Status != TitleChangeStatusPending {
		t.Errorf("status = %q, want PENDING", detail.Status)
	}
	if detail.PreviousTitle != validTitle() {
		t.Errorf("previous_title mismatch")
	}
	if len(tcrRepo.reqs) != 1 {
		t.Errorf("expected 1 request, got %d", len(tcrRepo.reqs))
	}
	select {
	case to := <-emailRec.requestedTo:
		if len(to) != 2 {
			t.Errorf("expected 2 email recipients (student+supervisor), got %v", to)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for requested email")
	}
}

func TestTitleChangeSubmitNotOwner(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	thesisID, _ := seedTCRThesis(t, thesisRepo, userRepo, studentID, "approved")

	uc, _, _, _ := newTestTCRUseCase(thesisRepo)
	_, err := uc.Submit(context.Background(), thesisID, CreateTitleChangeRequest{RequestedTitle: newTCRTitle()}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrTitleChangeForbidden) {
		t.Fatalf("expected ErrTitleChangeForbidden, got %v", err)
	}
}

func TestTitleChangeSubmitNotEligible(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	thesisID, _ := seedTCRThesis(t, thesisRepo, userRepo, studentID, "submitted")

	uc, _, _, _ := newTestTCRUseCase(thesisRepo)
	_, err := uc.Submit(context.Background(), thesisID, CreateTitleChangeRequest{RequestedTitle: newTCRTitle()}, Actor{UserID: studentID})
	if !errors.Is(err, ErrTitleChangeNotEligible) {
		t.Fatalf("expected ErrTitleChangeNotEligible, got %v", err)
	}
}

func TestTitleChangeSubmitNoSupervisor(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	registerFakeStudent(thesisRepo, studentID, userRepo)

	thesis := &entity.Thesis{
		StudentID:      studentID,
		AcademicYearID: uuid.New(),
		Title:          validTitle(),
		ThesisType:     "skripsi",
		Status:         "approved",
		SubmittedAt:    time.Now(),
	}
	if err := thesisRepo.Create(context.Background(), thesis); err != nil {
		t.Fatalf("seed thesis: %v", err)
	}
	// no supervisor attached

	uc, _, _, _ := newTestTCRUseCase(thesisRepo)
	_, err := uc.Submit(context.Background(), thesis.ID, CreateTitleChangeRequest{RequestedTitle: newTCRTitle()}, Actor{UserID: studentID})
	if !errors.Is(err, ErrNoSupervisorAssigned) {
		t.Fatalf("expected ErrNoSupervisorAssigned, got %v", err)
	}
}

// seedPendingTCR submits a title change request into a freshly wired use case and
// returns the request ID, thesis ID, supervisor ID, and the use case itself so
// tests act on the exact repo that holds the seeded request.
func seedPendingTCR(t *testing.T, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo, studentID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID, *TitleChangeRequestUseCase) {
	t.Helper()
	thesisID, supervisorID := seedTCRThesis(t, thesisRepo, userRepo, studentID, "approved")
	uc, _, _, _ := newTestTCRUseCase(thesisRepo)
	detail, err := uc.Submit(context.Background(), thesisID, CreateTitleChangeRequest{
		RequestedTitle: newTCRTitle(),
		Reason:         strPtr("Memperluas fokus penelitian"),
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("seed pending TCR: %v", err)
	}
	return detail.ID, thesisID, supervisorID, uc
}

func TestTitleChangeApproveSuccess(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, supervisorID, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	notes := "Disetujui, lanjutkan bimbingan"
	detail, err := uc.Approve(context.Background(), requestID, ReviewTitleChangeRequest{ReviewNotes: &notes}, Actor{UserID: supervisorID})
	if err != nil {
		t.Fatalf("Approve error: %v", err)
	}
	if detail.Status != TitleChangeStatusApproved {
		t.Errorf("status = %q, want APPROVED", detail.Status)
	}
	if detail.ReviewedBy == nil || detail.ReviewedBy.ID != supervisorID {
		t.Errorf("reviewed_by not set to supervisor")
	}
	if detail.ReviewNotes == nil || *detail.ReviewNotes != notes {
		t.Errorf("review_notes mismatch")
	}
	select {
	case to := <-uc.emailSvc.(*tcrRecordingEmail).approvedTo:
		if len(to) != 1 {
			t.Errorf("expected 1 student email recipient, got %v", to)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for approve email")
	}
}

func TestTitleChangeApproveNotSupervisor(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, _, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	_, err := uc.Approve(context.Background(), requestID, ReviewTitleChangeRequest{}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrTitleChangeNotSupervisor) {
		t.Fatalf("expected ErrTitleChangeNotSupervisor, got %v", err)
	}
}

func TestTitleChangeApproveNotPending(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, supervisorID, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	if _, err := uc.Approve(context.Background(), requestID, ReviewTitleChangeRequest{}, Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("first approve failed: %v", err)
	}
	_, err := uc.Approve(context.Background(), requestID, ReviewTitleChangeRequest{}, Actor{UserID: supervisorID})
	if !errors.Is(err, ErrTitleChangeNotPending) {
		t.Fatalf("expected ErrTitleChangeNotPending, got %v", err)
	}
}

func TestTitleChangeCancelSuccess(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, _, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	detail, err := uc.Cancel(context.Background(), requestID, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Cancel error: %v", err)
	}
	if detail.Status != TitleChangeStatusCancelled {
		t.Errorf("status = %q, want CANCELLED", detail.Status)
	}
	if detail.CancelledAt == nil {
		t.Error("cancelled_at should be set")
	}
	// Supervisors are notified for transparency.
	select {
	case to := <-uc.emailSvc.(*tcrRecordingEmail).cancelledTo:
		if len(to) != 1 || to[0] != "sup@example.com" {
			t.Errorf("expected supervisor email [sup@example.com], got %v", to)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for cancel email")
	}
}

func TestTitleChangeCancelNotOwner(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, _, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	_, err := uc.Cancel(context.Background(), requestID, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrTitleChangeForbidden) {
		t.Fatalf("expected ErrTitleChangeForbidden, got %v", err)
	}
}

func TestTitleChangeCancelNotPending(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, supervisorID, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	if _, err := uc.Approve(context.Background(), requestID, ReviewTitleChangeRequest{}, Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("approve before cancel: %v", err)
	}
	_, err := uc.Cancel(context.Background(), requestID, Actor{UserID: studentID})
	if !errors.Is(err, ErrTitleChangeNotPending) {
		t.Fatalf("expected ErrTitleChangeNotPending, got %v", err)
	}
}

func TestTitleChangeApproveNotFound(t *testing.T) {
	thesisRepo := newFakeThesisRepo()
	uc, _, _, _ := newTestTCRUseCase(thesisRepo)
	_, err := uc.Approve(context.Background(), uuid.New(), ReviewTitleChangeRequest{}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrTitleChangeNotFound) {
		t.Fatalf("expected ErrTitleChangeNotFound, got %v", err)
	}
}

func TestTitleChangeRejectSuccess(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, supervisorID, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	notes := "Judul belum spesifik"
	detail, err := uc.Reject(context.Background(), requestID, ReviewTitleChangeRequest{ReviewNotes: &notes}, Actor{UserID: supervisorID})
	if err != nil {
		t.Fatalf("Reject error: %v", err)
	}
	if detail.Status != TitleChangeStatusRejected {
		t.Errorf("status = %q, want REJECTED", detail.Status)
	}
	if detail.ReviewNotes == nil || *detail.ReviewNotes != notes {
		t.Errorf("review_notes mismatch")
	}
	select {
	case to := <-uc.emailSvc.(*tcrRecordingEmail).rejectedTo:
		if len(to) != 1 {
			t.Errorf("expected 1 student email recipient, got %v", to)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for reject email")
	}
}

func TestTitleChangeRejectRequiresNotes(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, supervisorID, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	_, err := uc.Reject(context.Background(), requestID, ReviewTitleChangeRequest{}, Actor{UserID: supervisorID})
	if !errors.Is(err, ErrTitleChangeReviewNotesReq) {
		t.Fatalf("expected ErrTitleChangeReviewNotesReq, got %v", err)
	}
}

func TestTitleChangeRejectNotSupervisor(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	requestID, _, _, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	notes := "Ditolak"
	_, err := uc.Reject(context.Background(), requestID, ReviewTitleChangeRequest{ReviewNotes: &notes}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrTitleChangeNotSupervisor) {
		t.Fatalf("expected ErrTitleChangeNotSupervisor, got %v", err)
	}
}

func TestTitleChangeListPendingForSupervisor(t *testing.T) {
	_, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo := newFakeThesisRepo()
	_, _, supervisorID, uc := seedPendingTCR(t, thesisRepo, userRepo, studentID)

	list, err := uc.ListPendingForSupervisor(context.Background(), supervisorID)
	if err != nil {
		t.Fatalf("ListPendingForSupervisor error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 pending request, got %d", len(list))
	}
	if list[0].Status != TitleChangeStatusPending {
		t.Errorf("status = %q, want PENDING", list[0].Status)
	}

	// A supervisor not assigned to the thesis must not see the request.
	otherSupervisor := uuid.New()
	empty, err := uc.ListPendingForSupervisor(context.Background(), otherSupervisor)
	if err != nil {
		t.Fatalf("ListPendingForSupervisor (other) error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 pending requests for unassigned supervisor, got %d", len(empty))
	}
}
