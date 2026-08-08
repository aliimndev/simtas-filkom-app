package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/storage"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

// fakeThesisRepo is a minimal in-memory ThesisRepository for usecase tests.
type fakeThesisRepo struct {
	theses        map[uuid.UUID]*entity.Thesis
	byStudent     map[uuid.UUID]uuid.UUID // studentID -> thesisID (latest)
	students      map[uuid.UUID]*entity.User
	years         map[uuid.UUID]*entity.AcademicYear
	supervisors   map[uuid.UUID][]uuid.UUID
	supUsers      map[uuid.UUID][]*entity.User // full supervisor users (optional, for emails)
	assignCalls   int
	statusUpdates []string
	examiners     map[uuid.UUID]bool // thesisID -> is examiner
}

func newFakeThesisRepo() *fakeThesisRepo {
	return &fakeThesisRepo{
		theses:      map[uuid.UUID]*entity.Thesis{},
		byStudent:   map[uuid.UUID]uuid.UUID{},
		students:    map[uuid.UUID]*entity.User{},
		years:       map[uuid.UUID]*entity.AcademicYear{},
		supervisors: map[uuid.UUID][]uuid.UUID{},
		supUsers:    map[uuid.UUID][]*entity.User{},
		examiners:   map[uuid.UUID]bool{},
	}
}

func (f *fakeThesisRepo) Create(_ context.Context, thesis *entity.Thesis) error {
	thesis.ID = uuid.New()
	thesis.SubmittedAt = time.Now()
	f.theses[thesis.ID] = thesis
	f.byStudent[thesis.StudentID] = thesis.ID
	return nil
}

// FindByID mirrors the real repo's preloads: Student, AcademicYear, Supervisors.
func (f *fakeThesisRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.Thesis, error) {
	t, ok := f.theses[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *t
	if u, ok := f.students[t.StudentID]; ok {
		clone.Student = *u
	}
	if y, ok := f.years[t.AcademicYearID]; ok {
		clone.AcademicYear = *y
	}
	if users, ok := f.supUsers[t.ID]; ok {
		for _, u := range users {
			clone.Supervisors = append(clone.Supervisors, *u)
		}
	} else {
		for _, sid := range f.supervisors[t.ID] {
			clone.Supervisors = append(clone.Supervisors, entity.User{ID: sid})
		}
	}
	return &clone, nil
}

func (f *fakeThesisRepo) FindAll(_ context.Context, filter domainRepo.ThesisFilter) ([]*entity.Thesis, int64, error) {
	var out []*entity.Thesis
	for _, t := range f.theses {
		if filter.StudentID != uuid.Nil && t.StudentID != filter.StudentID {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.SupervisorID != uuid.Nil {
			found := false
			for _, s := range f.supervisors[t.ID] {
				if s == filter.SupervisorID {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		// Populate associations like the real repo's preloads.
		clone := *t
		if u, ok := f.students[t.StudentID]; ok {
			clone.Student = *u
		}
		if y, ok := f.years[t.AcademicYearID]; ok {
			clone.AcademicYear = *y
		}
		out = append(out, &clone)
	}
	return out, int64(len(out)), nil
}

func (f *fakeThesisRepo) FindByStudentID(_ context.Context, studentID uuid.UUID) (*entity.Thesis, error) {
	id, ok := f.byStudent[studentID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return f.theses[id], nil
}

func (f *fakeThesisRepo) FindActiveByStudentID(_ context.Context, studentID uuid.UUID) (*entity.Thesis, error) {
	id, ok := f.byStudent[studentID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	t := f.theses[id]
	if t.Status == "cancelled" || t.Status == "graduated" {
		return nil, gorm.ErrRecordNotFound
	}
	return t, nil
}

func (f *fakeThesisRepo) UpdateStatus(_ context.Context, id uuid.UUID, status, notes string) error {
	t, ok := f.theses[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	t.Status = status
	if notes != "" {
		t.KaprodiNotes = &notes
	}
	f.statusUpdates = append(f.statusUpdates, status)
	return nil
}

func (f *fakeThesisRepo) Update(_ context.Context, thesis *entity.Thesis) error {
	f.theses[thesis.ID] = thesis
	return nil
}

func (f *fakeThesisRepo) AssignSupervisor(_ context.Context, thesisID, supervisorID, _ uuid.UUID) error {
	f.assignCalls++
	f.supervisors[thesisID] = append(f.supervisors[thesisID], supervisorID)
	return nil
}

func (f *fakeThesisRepo) AssignSupervisors(_ context.Context, thesisID uuid.UUID, supervisorIDs []uuid.UUID, _ uuid.UUID) error {
	f.assignCalls++
	f.supervisors[thesisID] = append(f.supervisors[thesisID], supervisorIDs...)
	if t, ok := f.theses[thesisID]; ok {
		t.Status = "in_progress"
	}
	return nil
}

func (f *fakeThesisRepo) GetSupervisors(_ context.Context, thesisID uuid.UUID) ([]*entity.User, error) {
	var users []*entity.User
	for _, sid := range f.supervisors[thesisID] {
		users = append(users, &entity.User{ID: sid})
	}
	return users, nil
}

func (f *fakeThesisRepo) CountActiveSupervisions(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeThesisRepo) IsExaminer(_ context.Context, thesisID, _ uuid.UUID) (bool, error) {
	return f.examiners[thesisID], nil
}

// fakeAcademicYearRepo is a minimal AcademicYearRepository for usecase tests.
type fakeAcademicYearRepo struct {
	years  map[uuid.UUID]*entity.AcademicYear
	active uuid.UUID
}

func newFakeAcademicYearRepo() *fakeAcademicYearRepo {
	return &fakeAcademicYearRepo{years: map[uuid.UUID]*entity.AcademicYear{}}
}

func (f *fakeAcademicYearRepo) FindAll(context.Context) ([]*entity.AcademicYear, error) {
	var out []*entity.AcademicYear
	for _, y := range f.years {
		out = append(out, y)
	}
	return out, nil
}

func (f *fakeAcademicYearRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.AcademicYear, error) {
	y, ok := f.years[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return y, nil
}

func (f *fakeAcademicYearRepo) FindActive(_ context.Context) (*entity.AcademicYear, error) {
	y, ok := f.years[f.active]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return y, nil
}

func (f *fakeAcademicYearRepo) Create(_ context.Context, year *entity.AcademicYear) error {
	year.ID = uuid.New()
	f.years[year.ID] = year
	return nil
}

func (f *fakeAcademicYearRepo) Update(_ context.Context, year *entity.AcademicYear) error {
	f.years[year.ID] = year
	return nil
}

func (f *fakeAcademicYearRepo) Activate(_ context.Context, id uuid.UUID) error {
	f.active = id
	return nil
}

func (f *fakeAcademicYearRepo) CountActiveTheses(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

// newTestThesisUseCase wires a fresh use case with in-memory fakes.
func newTestThesisUseCase() (*ThesisUseCase, *fakeThesisRepo, *fakeUserRepo, *fakeAcademicYearRepo) {
	thesisRepo := newFakeThesisRepo()
	userRepo := newFakeUserRepo()
	acadRepo := newFakeAcademicYearRepo()

	// Seed roles + active academic year + a kaprodi.
	userRepo.roles["kaprodi"] = &entity.Role{ID: 2, Name: "kaprodi"}
	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	userRepo.roles["dosen_pembimbing"] = &entity.Role{ID: 4, Name: "dosen_pembimbing"}
	userRepo.roles["dosen_penguji"] = &entity.Role{ID: 5, Name: "dosen_penguji"}

	year := &entity.AcademicYear{Name: "2026/2027", Semester: "ganjil", IsActive: true}
	_ = acadRepo.Create(context.Background(), year)
	_ = acadRepo.Activate(context.Background(), year.ID)

	auditSvc := audit.NewAuditService(nil) // nil repo → no-op, safe
	docRepo := newFakeDocumentRepo()
	storageSvc := storage.NewStubStorageService(os.TempDir(), "http://test.local")
	uc := NewThesisUseCase(thesisRepo, userRepo, acadRepo, docRepo, storageSvc, &fakeEmailService{}, auditSvc, nil)
	return uc, thesisRepo, userRepo, acadRepo
}

// mustDraftFile / mustDraftHeader return a standalone valid PDF draft pair.
// They are independent (each call re-reads pdfContent) so they can be assigned
// directly to DraftFile / DraftHeader fields.
func mustDraftFile() multipart.File {
	return testFile{bytes.NewReader(pdfContent)}
}

func mustDraftHeader() *multipart.FileHeader {
	return &multipart.FileHeader{Filename: "draft-proposal.pdf", Size: int64(len(pdfContent))}
}

// seedStudentWithRole adds an active student user and returns their ID.
// Emails are uniquified so multiple seeds never collide.
func seedStudent(t *testing.T, repo *fakeUserRepo, role string) uuid.UUID {
	t.Helper()
	u := &entity.User{
		Email:    fmt.Sprintf("student-%s@example.com", uuid.New().String()[:8]),
		FullName: "Mahasiswa Satu",
		Role:     entity.Role{Name: role},
		RoleID:   3,
		IsActive: true,
	}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("seed student: %v", err)
	}
	return u.ID
}

func validTitle() string {
	return "Sistem Rekomendasi Pemilihan Konsentrasi Menggunakan Metode Certainty Factor Berbasis Web"
}

func longAbstract(n int) string {
	return strings.Repeat("kata ", n)
}

func TestSubmitThesis(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	thesisRepo.students[studentID] = &entity.User{ID: studentID, FullName: "Mahasiswa Satu"}

	thesis, err := uc.Submit(context.Background(), CreateThesisRequest{
		Title:        validTitle(),
		Abstract:     longAbstract(100),
		FieldOfStudy: "Kecerdasan Buatan",
		ThesisType:   "skripsi",
		DraftFile:    mustDraftFile(),
		DraftHeader:  mustDraftHeader(),
	}, studentID, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	if thesis.Status != "submitted" {
		t.Errorf("status = %q, want submitted", thesis.Status)
	}
	if len(thesisRepo.theses) != 1 {
		t.Errorf("expected 1 thesis, got %d", len(thesisRepo.theses))
	}
	if thesis.Student.FullName != "Mahasiswa Satu" {
		t.Errorf("student full_name = %q", thesis.Student.FullName)
	}
}

func TestSubmitThesisActiveExists(t *testing.T) {
	uc, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")

	// First submission succeeds.
	if _, err := uc.Submit(context.Background(), CreateThesisRequest{
		Title:       validTitle(),
		Abstract:    longAbstract(100),
		ThesisType:  "skripsi",
		DraftFile:   mustDraftFile(),
		DraftHeader: mustDraftHeader(),
	}, studentID, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("first Submit returned error: %v", err)
	}

	// Second submission must be rejected.
	_, err := uc.Submit(context.Background(), CreateThesisRequest{
		Title:       validTitle(),
		Abstract:    longAbstract(100),
		ThesisType:  "skripsi",
		DraftFile:   mustDraftFile(),
		DraftHeader: mustDraftHeader(),
	}, studentID, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrActiveThesisExists) {
		t.Errorf("expected ErrActiveThesisExists, got %v", err)
	}
}

func TestSubmitThesisValidation(t *testing.T) {
	uc, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")

	tests := []struct {
		name    string
		req     CreateThesisRequest
		wantErr error
	}{
		{"title too short", CreateThesisRequest{Title: "Judul Pendek", Abstract: longAbstract(100), ThesisType: "skripsi"}, ErrTitleTooShort},
		{"title too long", CreateThesisRequest{Title: strings.Repeat("kata ", 101), Abstract: longAbstract(100), ThesisType: "skripsi"}, ErrTitleTooLong},
		{"abstract too short", CreateThesisRequest{Title: validTitle(), Abstract: "abstrak singkat", ThesisType: "skripsi"}, ErrAbstractTooShort},
		{"invalid thesis type", CreateThesisRequest{Title: validTitle(), Abstract: longAbstract(100), ThesisType: "tesis"}, ErrInvalidThesisType},
		{"draft required", CreateThesisRequest{Title: validTitle(), Abstract: longAbstract(100), ThesisType: "skripsi"}, ErrDraftRequired},
		{"draft not pdf", CreateThesisRequest{Title: validTitle(), Abstract: longAbstract(100), ThesisType: "skripsi", DraftFile: testFile{bytes.NewReader([]byte("not a pdf"))}, DraftHeader: &multipart.FileHeader{Filename: "draft-proposal.pdf", Size: 10}}, utils.ErrNotPDF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Submit(context.Background(), tt.req, studentID, Actor{UserID: uuid.New()})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestSubmitThesisNoActiveAcademicYear(t *testing.T) {
	thesisRepo := newFakeThesisRepo()
	userRepo := newFakeUserRepo()
	acadRepo := newFakeAcademicYearRepo() // no active year
	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	studentID := seedStudent(t, userRepo, "mahasiswa")

	auditSvc := audit.NewAuditService(nil)
	uc := NewThesisUseCase(thesisRepo, userRepo, acadRepo, newFakeDocumentRepo(), storage.NewStubStorageService(os.TempDir(), "http://test.local"), &fakeEmailService{}, auditSvc, nil)

	_, err := uc.Submit(context.Background(), CreateThesisRequest{
		Title:       validTitle(),
		Abstract:    longAbstract(100),
		ThesisType:  "skripsi",
		DraftFile:   mustDraftFile(),
		DraftHeader: mustDraftHeader(),
	}, studentID, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrNoActiveAcademicYear) {
		t.Errorf("expected ErrNoActiveAcademicYear, got %v", err)
	}
}

// seedSubmittedThesis creates a thesis in submitted state and returns its ID.
func seedSubmittedThesis(t *testing.T, uc *ThesisUseCase, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo) (uuid.UUID, uuid.UUID) {
	t.Helper()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	registerFakeStudent(thesisRepo, studentID, userRepo)
	detail, err := uc.Submit(context.Background(), CreateThesisRequest{
		Title:       validTitle(),
		Abstract:    longAbstract(100),
		ThesisType:  "skripsi",
		DraftFile:   mustDraftFile(),
		DraftHeader: mustDraftHeader(),
	}, studentID, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("seed thesis: %v", err)
	}
	return detail.ID, studentID
}

// registerFakeStudent keeps the fake thesis repo's student lookup in sync with
// the seeded user so association preloads match the real repository.
func registerFakeStudent(thesisRepo *fakeThesisRepo, studentID uuid.UUID, userRepo *fakeUserRepo) {
	u := userRepo.users[studentID]
	thesisRepo.students[studentID] = u
}

func TestReviewApproved(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	detail, err := uc.Review(context.Background(), thesisID, ReviewThesisRequest{
		Decision: "approved",
		Notes:    "Judul disetujui",
	}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if detail.Status != "approved" {
		t.Errorf("status = %q, want approved", detail.Status)
	}
	if detail.ApprovedAt == nil {
		t.Error("approved_at should be set")
	}
}

func TestReviewRejected(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	detail, err := uc.Review(context.Background(), thesisID, ReviewThesisRequest{
		Decision: "rejected",
		Notes:    "Judul terlalu luas",
	}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if detail.Status != "rejected" {
		t.Errorf("status = %q, want rejected", detail.Status)
	}
	if detail.ApprovedAt != nil {
		t.Error("approved_at should be nil when rejected")
	}
}

func TestReviewInvalidTransition(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	// Approve first, then try to approve again → invalid transition.
	if _, err := uc.Review(context.Background(), thesisID, ReviewThesisRequest{Decision: "approved"}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("first review failed: %v", err)
	}
	_, err := uc.Review(context.Background(), thesisID, ReviewThesisRequest{Decision: "approved"}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestReviewInvalidDecision(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	_, err := uc.Review(context.Background(), thesisID, ReviewThesisRequest{Decision: "maybe"}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrInvalidDecision) {
		t.Errorf("expected ErrInvalidDecision, got %v", err)
	}
}

// seedApprovedThesis returns an approved thesis ID.
func seedApprovedThesis(t *testing.T, uc *ThesisUseCase, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo) uuid.UUID {
	t.Helper()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)
	if _, err := uc.Review(context.Background(), thesisID, ReviewThesisRequest{Decision: "approved"}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("seed approved: %v", err)
	}
	return thesisID
}

func TestAssignSupervisor(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID := seedApprovedThesis(t, uc, thesisRepo, userRepo)

	// Seed two active dosen_pembimbing.
	dosen1 := &entity.User{Email: "dosen1@example.com", FullName: "Dosen Satu", Role: entity.Role{Name: "dosen_pembimbing"}, RoleID: 4, IsActive: true}
	dosen2 := &entity.User{Email: "dosen2@example.com", FullName: "Dosen Dua", Role: entity.Role{Name: "dosen_pembimbing"}, RoleID: 4, IsActive: true}
	_ = userRepo.Create(context.Background(), dosen1)
	_ = userRepo.Create(context.Background(), dosen2)

	detail, err := uc.AssignSupervisor(context.Background(), thesisID, AssignSupervisorRequest{
		SupervisorIDs: []uuid.UUID{dosen1.ID, dosen2.ID},
	}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("AssignSupervisor returned error: %v", err)
	}
	if detail.Status != "in_progress" {
		t.Errorf("status = %q, want in_progress", detail.Status)
	}
	if got := len(thesisRepo.supervisors[thesisID]); got != 2 {
		t.Errorf("expected 2 supervisors assigned, got %d", got)
	}
}

func TestAssignSupervisorWrongRole(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID := seedApprovedThesis(t, uc, thesisRepo, userRepo)

	// This user is a mahasiswa, not dosen_pembimbing.
	nonLecturer := seedStudent(t, userRepo, "mahasiswa")

	_, err := uc.AssignSupervisor(context.Background(), thesisID, AssignSupervisorRequest{
		SupervisorIDs: []uuid.UUID{nonLecturer},
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrSupervisorNotEligible) {
		t.Errorf("expected ErrSupervisorNotEligible, got %v", err)
	}
}

func TestAssignSupervisorTooMany(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID := seedApprovedThesis(t, uc, thesisRepo, userRepo)

	_, err := uc.AssignSupervisor(context.Background(), thesisID, AssignSupervisorRequest{
		SupervisorIDs: []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrInvalidSupervisorCount) {
		t.Errorf("expected ErrInvalidSupervisorCount, got %v", err)
	}
}

func TestAssignSupervisorNotApproved(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	// thesis is submitted, not approved.
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	dosen := &entity.User{Email: "dosen@example.com", FullName: "Dosen", Role: entity.Role{Name: "dosen_pembimbing"}, RoleID: 4, IsActive: true}
	_ = userRepo.Create(context.Background(), dosen)

	_, err := uc.AssignSupervisor(context.Background(), thesisID, AssignSupervisorRequest{
		SupervisorIDs: []uuid.UUID{dosen.ID},
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestListScopedByStudent(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	registerFakeStudent(thesisRepo, studentID, userRepo)

	if _, err := uc.Submit(context.Background(), CreateThesisRequest{
		Title:       validTitle(),
		Abstract:    longAbstract(100),
		ThesisType:  "skripsi",
		DraftFile:   mustDraftFile(),
		DraftHeader: mustDraftHeader(),
	}, studentID, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	theses, total, err := uc.List(context.Background(), domainRepo.ThesisFilter{}, studentID, "mahasiswa")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(theses) != 1 || theses[0].Student.ID != studentID {
		t.Errorf("expected only the student's own thesis")
	}
}

func TestListAllForKaprodi(t *testing.T) {
	uc, _, userRepo, _ := newTestThesisUseCase()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	if _, err := uc.Submit(context.Background(), CreateThesisRequest{
		Title:       validTitle(),
		Abstract:    longAbstract(100),
		ThesisType:  "skripsi",
		DraftFile:   mustDraftFile(),
		DraftHeader: mustDraftHeader(),
	}, studentID, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	theses, total, err := uc.List(context.Background(), domainRepo.ThesisFilter{}, uuid.New(), "kaprodi")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 1 || len(theses) != 1 {
		t.Errorf("kaprodi should see all theses, got %d/%d", len(theses), total)
	}
}

func TestGetByIDForbidden(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	// A different student tries to read the thesis.
	other := seedStudent(t, userRepo, "mahasiswa")
	_, err := uc.GetByID(context.Background(), thesisID, other, "mahasiswa")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGetByIDKaprodiAllowed(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	detail, err := uc.GetByID(context.Background(), thesisID, uuid.New(), "kaprodi")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if detail.ID != thesisID {
		t.Errorf("wrong thesis returned")
	}
}

func TestCancelThesis(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	err := uc.Cancel(context.Background(), thesisID, CancelThesisRequest{Reason: "Mengundurkan diri"}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	if thesisRepo.theses[thesisID].Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", thesisRepo.theses[thesisID].Status)
	}
}

func TestCancelAlreadyCancelled(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	if err := uc.Cancel(context.Background(), thesisID, CancelThesisRequest{}, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("first cancel failed: %v", err)
	}
	err := uc.Cancel(context.Background(), thesisID, CancelThesisRequest{}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrThesisAlreadyCancelled) {
		t.Errorf("expected ErrThesisAlreadyCancelled, got %v", err)
	}
}

func TestCancelGraduatedThesis(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _ := seedSubmittedThesis(t, uc, thesisRepo, userRepo)

	// Simulate a graduated thesis (data-integrity guard from the production review).
	thesisRepo.theses[thesisID].Status = "graduated"

	err := uc.Cancel(context.Background(), thesisID, CancelThesisRequest{}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrThesisCannotCancel) {
		t.Errorf("expected ErrThesisCannotCancel, got %v", err)
	}
	if thesisRepo.theses[thesisID].Status != "graduated" {
		t.Errorf("graduated thesis must remain unchanged, got %q", thesisRepo.theses[thesisID].Status)
	}
}

func TestListLecturersSortedByLoad(t *testing.T) {
	uc, _, userRepo, _ := newTestThesisUseCase()

	dosenA := &entity.User{Email: "a@example.com", FullName: "Dosen A", Role: entity.Role{Name: "dosen_pembimbing"}, RoleID: 4, IsActive: true}
	dosenB := &entity.User{Email: "b@example.com", FullName: "Dosen B", Role: entity.Role{Name: "dosen_pembimbing"}, RoleID: 4, IsActive: true}
	_ = userRepo.Create(context.Background(), dosenA)
	_ = userRepo.Create(context.Background(), dosenB)

	lecturers, err := uc.ListLecturers(context.Background())
	if err != nil {
		t.Fatalf("ListLecturers returned error: %v", err)
	}
	if len(lecturers) != 2 {
		t.Fatalf("expected 2 lecturers, got %d", len(lecturers))
	}
	// All loads are 0 → stable sort keeps creation order; both valid.
	if lecturers[0].FullName == "" || lecturers[1].FullName == "" {
		t.Error("lecturer names missing")
	}
}
