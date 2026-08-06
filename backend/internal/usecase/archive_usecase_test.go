package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/storage"
)

// fakeArchiveRepo is a minimal in-memory ArchiveRepository for usecase tests.
type fakeArchiveRepo struct {
	archives map[uuid.UUID]*entity.ThesisArchive
	byThesis map[uuid.UUID]uuid.UUID
}

func newFakeArchiveRepo() *fakeArchiveRepo {
	return &fakeArchiveRepo{
		archives: map[uuid.UUID]*entity.ThesisArchive{},
		byThesis: map[uuid.UUID]uuid.UUID{},
	}
}

var _ repository.ArchiveRepository = (*fakeArchiveRepo)(nil)

func (f *fakeArchiveRepo) Create(_ context.Context, archive *entity.ThesisArchive) error {
	archive.ID = uuid.New()
	f.archives[archive.ID] = archive
	f.byThesis[archive.ThesisID] = archive.ID
	return nil
}

func (f *fakeArchiveRepo) FindByThesisID(_ context.Context, thesisID uuid.UUID) (*entity.ThesisArchive, error) {
	id, ok := f.byThesis[thesisID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return f.FindByID(context.Background(), id)
}

func (f *fakeArchiveRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.ThesisArchive, error) {
	a, ok := f.archives[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *a
	// Populate the thesis + student so toArchiveDetail works (mirrors the real
	// repo's Preload("Thesis.Student.Role")). The stored archive only carries
	// ThesisID, so guard on that, not on a (zero) Thesis association.
	if a.ThesisID != uuid.Nil {
		clone.Thesis = entity.Thesis{
			ID:           a.ThesisID,
			Title:        "Judul Skripsi",
			FieldOfStudy: strPtr("Kecerdasan Buatan"),
			Student: entity.User{
				ID:           uuid.New(),
				FullName:     "Mahasiswa Satu",
				StudyProgram: strPtr("Teknik Informatika"),
			},
		}
	}
	return &clone, nil
}

func (f *fakeArchiveRepo) Search(_ context.Context, filter repository.ArchiveFilter) ([]*entity.ThesisArchive, int64, error) {
	var out []*entity.ThesisArchive
	for _, a := range f.archives {
		if filter.GraduationYear > 0 && a.GraduationYear != filter.GraduationYear {
			continue
		}
		if filter.Query != "" {
			// Naive relevance match for tests: title or keywords contain the query.
			joined := strings.Join(a.Keywords, " ") + " " + a.Thesis.Title
			if !strings.Contains(strings.ToLower(joined), strings.ToLower(filter.Query)) {
				continue
			}
		}
		clone := *a
		out = append(out, &clone)
	}
	return out, int64(len(out)), nil
}

func (f *fakeArchiveRepo) Stats(_ context.Context) (*repository.ArchiveStats, error) {
	stats := &repository.ArchiveStats{
		ByYear:         []repository.ArchiveCountByYear{},
		ByField:        []repository.ArchiveCountByField{},
		ByStudyProgram: []repository.ArchiveCountByProgram{},
	}
	stats.TotalArchives = int64(len(f.archives))
	for _, a := range f.archives {
		stats.ByYear = append(stats.ByYear, repository.ArchiveCountByYear{Year: a.GraduationYear, Count: 1})
	}
	return stats, nil
}

// recordingArchiveEmailService records archive email recipients for assertions.
type recordingArchiveEmailService struct {
	createdTo chan string
}

func newRecordingArchiveEmailService() *recordingArchiveEmailService {
	return &recordingArchiveEmailService{createdTo: make(chan string, 16)}
}

func (r *recordingArchiveEmailService) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (r *recordingArchiveEmailService) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
func (r *recordingArchiveEmailService) SendThesisSubmitted(context.Context, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingArchiveEmailService) SendThesisApproved(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingArchiveEmailService) SendThesisRejected(context.Context, string, *entity.Thesis, string) error {
	return nil
}
func (r *recordingArchiveEmailService) SendSupervisorAssigned(context.Context, string, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingArchiveEmailService) SendConsultationCreated(context.Context, []string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingArchiveEmailService) SendConsultationApproved(context.Context, string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingArchiveEmailService) SendDocumentUploaded(context.Context, []string, *entity.Document) error {
	return nil
}
func (r *recordingArchiveEmailService) SendDocumentReviewed(context.Context, string, *entity.Document, string) error {
	return nil
}
func (r *recordingArchiveEmailService) SendSeminarSubmitted(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingArchiveEmailService) SendSeminarScheduled(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingArchiveEmailService) SendSeminarFinalized(context.Context, string, *entity.Seminar) error {
	return nil
}
func (r *recordingArchiveEmailService) SendDefenseSubmitted(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingArchiveEmailService) SendDefenseScheduled(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingArchiveEmailService) SendDefenseFinalized(context.Context, string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingArchiveEmailService) SendGraduated(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingArchiveEmailService) SendArchiveCreated(_ context.Context, to string, _ *entity.ThesisArchive) error {
	r.createdTo <- to
	return nil
}

func (r *recordingArchiveEmailService) SendTitleChangeRequested(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingArchiveEmailService) SendTitleChangeCancelled(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingArchiveEmailService) SendTitleChangeApproved(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingArchiveEmailService) SendTitleChangeRejected(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

// newTestArchiveUseCase wires a fresh archive use case with in-memory fakes.
func newTestArchiveUseCase(t *testing.T) (*ArchiveUseCase, *fakeArchiveRepo, *fakeThesisRepo, *fakeUserRepo, *recordingArchiveEmailService) {
	t.Helper()
	archiveRepo := newFakeArchiveRepo()
	thesisRepo := newFakeThesisRepo()
	userRepo := newFakeUserRepo()
	email := newRecordingArchiveEmailService()

	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	userRepo.roles["admin_fakultas"] = &entity.Role{ID: 1, Name: "admin_fakultas"}
	userRepo.roles["dosen_pembimbing"] = &entity.Role{ID: 4, Name: "dosen_pembimbing"}

	auditSvc := audit.NewAuditService(nil)
	storageSvc := storage.NewStubStorageService(t.TempDir(), "http://test.local")
	uc := NewArchiveUseCase(archiveRepo, thesisRepo, storageSvc, email, auditSvc)
	return uc, archiveRepo, thesisRepo, userRepo, email
}

// seedGraduatedThesis builds a thesis in `graduated` state owned by a student;
// returns thesisID, studentID.
func seedGraduatedThesis(t *testing.T, thesisRepo *fakeThesisRepo, userRepo *fakeUserRepo) (uuid.UUID, uuid.UUID) {
	t.Helper()
	studentID := seedStudent(t, userRepo, "mahasiswa")
	registerFakeStudent(thesisRepo, studentID, userRepo)
	tThesis := &entity.Thesis{
		StudentID:      studentID,
		AcademicYearID: uuid.New(),
		Title:          validTitle(),
		ThesisType:     "skripsi",
		Status:         "graduated",
	}
	if err := thesisRepo.Create(context.Background(), tThesis); err != nil {
		t.Fatalf("seed thesis: %v", err)
	}
	return tThesis.ID, studentID
}

func validArchiveRequest() CreateArchiveRequest {
	return CreateArchiveRequest{
		AbstractID:     strings.Repeat("kata ", 60),
		Keywords:       []string{"machine learning", "deep learning", "computer vision"},
		GraduationYear: time.Now().Year(), // must not exceed the current year
	}
}

func TestCreateArchiveOnGraduatedThesis(t *testing.T) {
	uc, archiveRepo, thesisRepo, userRepo, email := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	detail, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if detail.Title != "Judul Skripsi" {
		t.Errorf("title = %q, want Judul Skripsi", detail.Title)
	}
	if len(archiveRepo.archives) != 1 {
		t.Errorf("expected 1 archive, got %d", len(archiveRepo.archives))
	}
	if len(detail.Keywords) != 3 {
		t.Errorf("keywords = %d, want 3", len(detail.Keywords))
	}
	// Student is notified (async).
	select {
	case to := <-email.createdTo:
		student := userRepo.users[studentID]
		if to != student.Email {
			t.Errorf("archive email to = %q, want %q", to, student.Email)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for archive-created email")
	}
}

func TestCreateArchiveThesisNotGraduated(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	// Thesis is in_progress (not graduated) → gate fails.
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "x.pdf", pdfContent)
	_, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrArchiveThesisNotGrad) {
		t.Errorf("expected ErrArchiveThesisNotGrad, got %v", err)
	}
}

func TestCreateArchiveDuplicateRejected(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	if _, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa); err != nil {
		t.Fatalf("first create: %v", err)
	}
	file2, header2 := makePDFHeader(t, "skripsi_final2.pdf", pdfContent)
	_, err := uc.Create(context.Background(), thesisID, file2, header2, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrArchiveExists) {
		t.Errorf("expected ErrArchiveExists, got %v", err)
	}
}

func TestCreateArchiveByAdmin(t *testing.T) {
	uc, archiveRepo, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, _ := seedGraduatedThesis(t, thesisRepo, userRepo)
	admin := &entity.User{
		Email:    "admin-archive@example.com",
		FullName: "Admin Fakultas",
		Role:     entity.Role{Name: "admin_fakultas"},
		RoleID:   1,
		IsActive: true,
	}
	if err := userRepo.Create(context.Background(), admin); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	adminID := admin.ID

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	_, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: adminID}, ThesisRoleAdminFakultas)
	if err != nil {
		t.Fatalf("Create by admin returned error: %v", err)
	}
	if len(archiveRepo.archives) != 1 {
		t.Errorf("expected 1 archive, got %d", len(archiveRepo.archives))
	}
}

func TestCreateArchiveForbiddenForOtherStudent(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, _ := seedGraduatedThesis(t, thesisRepo, userRepo)
	other := seedStudent(t, userRepo, "mahasiswa")

	file, header := makePDFHeader(t, "x.pdf", pdfContent)
	_, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: other}, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateArchiveAbstractTooShort(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	req := validArchiveRequest()
	req.AbstractID = "abstrak singkat"
	file, header := makePDFHeader(t, "x.pdf", pdfContent)
	_, err := uc.Create(context.Background(), thesisID, file, header, req, Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrArchiveAbstractShort) {
		t.Errorf("expected ErrArchiveAbstractShort, got %v", err)
	}
}

func TestCreateArchiveTooFewKeywords(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	req := validArchiveRequest()
	req.Keywords = []string{"satu", "dua"}
	file, header := makePDFHeader(t, "x.pdf", pdfContent)
	_, err := uc.Create(context.Background(), thesisID, file, header, req, Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrArchiveKeywordsShort) {
		t.Errorf("expected ErrArchiveKeywordsShort, got %v", err)
	}
}

func TestCreateArchiveInvalidYear(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	req := validArchiveRequest()
	req.GraduationYear = 1999
	file, header := makePDFHeader(t, "x.pdf", pdfContent)
	_, err := uc.Create(context.Background(), thesisID, file, header, req, Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrArchiveInvalidYear) {
		t.Errorf("expected ErrArchiveInvalidYear, got %v", err)
	}
}

func TestArchiveSearchWithFilters(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	if _, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Full-text query matches the keywords.
	results, total, err := uc.Search(context.Background(), repository.ArchiveFilter{Query: "machine learning"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Errorf("query search: total=%d len=%d, want 1/1", total, len(results))
	}

	// Non-matching query returns nothing.
	_, totalNoMatch, err := uc.Search(context.Background(), repository.ArchiveFilter{Query: "quantum computing"})
	if err != nil {
		t.Fatalf("Search no-match: %v", err)
	}
	if totalNoMatch != 0 {
		t.Errorf("no-match query: total=%d, want 0", totalNoMatch)
	}

	// Year filter works.
	resultsYear, totalYear, err := uc.Search(context.Background(), repository.ArchiveFilter{GraduationYear: time.Now().Year()})
	if err != nil {
		t.Fatalf("Search by year: %v", err)
	}
	if totalYear != 1 || len(resultsYear) != 1 {
		t.Errorf("year filter: total=%d len=%d, want 1/1", totalYear, len(resultsYear))
	}
}

func TestArchiveDownloadByOwner(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	detail, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	result, err := uc.Download(context.Background(), detail.ID, studentID, ThesisRoleMahasiswa, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Download by owner returned error: %v", err)
	}
	if result.DownloadURL == "" || result.ExpiresIn != ArchiveDownloadExpirySecs {
		t.Errorf("download result incomplete: %+v", result)
	}
}

func TestArchiveDownloadByOtherStudentForbidden(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	detail, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	other := seedStudent(t, userRepo, "mahasiswa")

	_, err = uc.Download(context.Background(), detail.ID, other, ThesisRoleMahasiswa, Actor{UserID: other})
	if !errors.Is(err, ErrArchiveDownloadDenied) {
		t.Errorf("expected ErrArchiveDownloadDenied, got %v", err)
	}
}

func TestArchiveDownloadBySupervisorAllowed(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	detail, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// A dosen_pembimbing (not the owner) can download any archive.
	dosen := seedOtherDosen(t, userRepo)

	result, err := uc.Download(context.Background(), detail.ID, dosen, ThesisRoleDosenPembimbing, Actor{UserID: dosen})
	if err != nil {
		t.Fatalf("Download by supervisor returned error: %v", err)
	}
	if result.DownloadURL == "" {
		t.Error("expected a presigned URL")
	}
}

func TestArchiveGetByThesisID(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	created, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	detail, err := uc.GetByThesisID(context.Background(), thesisID)
	if err != nil {
		t.Fatalf("GetByThesisID: %v", err)
	}
	if detail.ID != created.ID {
		t.Errorf("archive id = %v, want %v", detail.ID, created.ID)
	}
}

func TestArchiveGetByThesisIDNotFound(t *testing.T) {
	uc, _, _, _, _ := newTestArchiveUseCase(t)
	_, err := uc.GetByThesisID(context.Background(), uuid.New())
	if !errors.Is(err, ErrArchiveNotFound) {
		t.Errorf("expected ErrArchiveNotFound, got %v", err)
	}
}

func TestArchiveStats(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	if _, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa); err != nil {
		t.Fatalf("create: %v", err)
	}

	stats, err := uc.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalArchives != 1 {
		t.Errorf("total archives = %d, want 1", stats.TotalArchives)
	}
	if len(stats.ByYear) != 1 || stats.ByYear[0].Year != time.Now().Year() {
		t.Errorf("by_year = %+v, want year %d count 1", stats.ByYear, time.Now().Year())
	}
}
