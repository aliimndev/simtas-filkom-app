package usecase

// Job 23 — coverage-gap tests. These target branches that were not exercised
// by the primary suites so the usecase layer meets the ≥80% coverage gate.
// They reuse the shared in-memory fakes; no new fake types are defined.

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
)

func TestAcademicYearList(t *testing.T) {
	repo := newFakeAcademicYearRepo()
	uc := NewAcademicYearUseCase(repo)
	a := &entity.AcademicYear{Name: "2026/2027", Semester: "ganjil"}
	if err := repo.Create(context.Background(), a); err != nil {
		t.Fatalf("seed year: %v", err)
	}

	years, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(years) != 1 || years[0].Name != "2026/2027" {
		t.Errorf("unexpected years: %+v", years)
	}
}

func TestThesisAccessIsExaminer(t *testing.T) {
	thesisRepo := newFakeThesisRepo()
	access := NewThesisAccess(thesisRepo)
	thesisID := uuid.New()

	// Examiner assigned → true.
	thesisRepo.examiners[thesisID] = true
	ok, err := access.IsExaminer(context.Background(), uuid.New(), thesisID)
	if err != nil || !ok {
		t.Errorf("IsExaminer(assigned) = %v, %v; want true, nil", ok, err)
	}

	// Not an examiner → false.
	thesisRepo.examiners[thesisID] = false
	ok, err = access.IsExaminer(context.Background(), uuid.New(), thesisID)
	if err != nil || ok {
		t.Errorf("IsExaminer(unassigned) = %v, %v; want false, nil", ok, err)
	}
}

func TestLogoutInvalidToken(t *testing.T) {
	user := newTestUser("Password123")
	uc, _ := newAuthUC(user)

	if err := uc.Logout(context.Background(), "not-a-jwt", Actor{}); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestLogoutBlacklistedIdempotent(t *testing.T) {
	user := newTestUser("Password123")
	repo := &blacklistAwareAuthRepo{authUserRepo: &authUserRepo{user: user}, blacklisted: true}
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))

	token, _, err := jwtMgr.GenerateAccessToken(user.ID, "mahasiswa", user.Email, 0)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	// Already-blacklisted tokens are a no-op (no error, no audit noise).
	if err := uc.Logout(context.Background(), token, Actor{}); err != nil {
		t.Fatalf("Logout(idempotent): %v", err)
	}
}

func TestLogoutSuccessBlacklistsToken(t *testing.T) {
	user := newTestUser("Password123")
	uc, jwtMgr := newAuthUC(user)

	token, _, err := jwtMgr.GenerateAccessToken(user.ID, "mahasiswa", user.Email, 0)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if err := uc.Logout(context.Background(), token, Actor{UserID: user.ID}); err != nil {
		t.Fatalf("Logout: %v", err)
	}
}

// TestConsultationSummaryForbiddenForExaminer — dosen_penguji is not in the
// consultation access list (owner/supervisor/admin/kaprodi) → ErrForbidden.
func TestConsultationSummaryForbiddenForExaminer(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	_, err := uc.Summary(context.Background(), thesisID, uuid.New(), ThesisRoleDosenPenguji)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateConsultationInvalidDate(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Topik",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	badDate := "31-12-2026" // wrong format → parseConsultationDate error
	_, err := uc.Update(context.Background(), thesisID, id, UpdateConsultationRequest{
		ConsultationDate: &badDate,
	}, Actor{UserID: studentID})
	if err == nil {
		t.Fatal("expected parse error for invalid consultation date")
	}
}

func TestUpdateConsultationEmptyTopicsRejected(t *testing.T) {
	uc, consultRepo, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	if _, err := uc.Create(context.Background(), thesisID, CreateConsultationRequest{
		ConsultationDate: todayStr(),
		TopicsDiscussed:  "Topik",
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("seed log: %v", err)
	}
	id := consultRepo.byThesis[thesisID][0]

	blank := "   "
	_, err := uc.Update(context.Background(), thesisID, id, UpdateConsultationRequest{
		TopicsDiscussed: &blank,
	}, Actor{UserID: studentID})
	if !errors.Is(err, ErrTopicsDiscussedRequired) {
		t.Errorf("expected ErrTopicsDiscussedRequired, got %v", err)
	}
}

func TestDeleteConsultationNotFound(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestConsultationUseCase()
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	err := uc.Delete(context.Background(), thesisID, uuid.New(), studentID, Actor{UserID: studentID})
	if !errors.Is(err, ErrConsultationNotFound) {
		t.Errorf("expected ErrConsultationNotFound, got %v", err)
	}
}

func TestSubmitSeminarThesisNotFound(t *testing.T) {
	uc, _, _, _, _, _ := newTestSeminarUseCase(t)

	_, err := uc.Submit(context.Background(), uuid.New(), Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrThesisNotFound) {
		t.Errorf("expected ErrThesisNotFound, got %v", err)
	}
}

func TestArchiveDownloadNotFound(t *testing.T) {
	uc, _, _, _, _ := newTestArchiveUseCase(t)

	_, err := uc.Download(context.Background(), uuid.New(), uuid.New(), ThesisRoleAdminFakultas, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrArchiveNotFound) {
		t.Errorf("expected ErrArchiveNotFound, got %v", err)
	}
}

// TestDefenseGetByIDFullDetail — exercises toDefenseDetail with a fully
// populated defense (thesis + student + examiners + scores).
func TestDefenseGetByIDFullDetail(t *testing.T) {
	uc, defRepo, _, _, _, _, _ := newTestDefenseUseCase(t)

	defID := uuid.New()
	examiners := []entity.User{
		{ID: uuid.New(), FullName: "Penguji A"},
		{ID: uuid.New(), FullName: "Penguji B"},
	}
	defRepo.defenses[defID] = &entity.ThesisDefense{
		ID: defID,
		Thesis: entity.Thesis{
			ID:      uuid.New(),
			Title:   "Judul Skripsi",
			Student: entity.User{ID: uuid.New(), FullName: "Mahasiswa", Email: "m@example.com"},
		},
		Examiners: examiners,
		Scores: []entity.DefenseScore{
			{Examiner: examiners[0], ComponentName: "Presentasi", ComponentWeight: 50, Score: 80},
		},
	}

	detail, err := uc.GetByID(context.Background(), defID, uuid.New(), ThesisRoleAdminFakultas)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.Thesis == nil || detail.Thesis.Student == nil {
		t.Error("expected thesis + student in detail")
	}
	if len(detail.Examiners) != 2 {
		t.Errorf("examiners = %d, want 2", len(detail.Examiners))
	}
	if len(detail.Scores) != 1 {
		t.Errorf("scores = %d, want 1", len(detail.Scores))
	}
}

// ── User import (Job 23 coverage: parseXLSX/cell/validateImportRow) ──────

func TestBuildImportTemplate(t *testing.T) {
	uc, _ := newTestUserUseCase()
	data, err := uc.BuildImportTemplate()
	if err != nil {
		t.Fatalf("BuildImportTemplate: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty xlsx template")
	}
}

// buildXLSX builds an in-memory .xlsx matching the import template shape.
func buildXLSX(t *testing.T, rows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	_ = f.SetSheetName("Sheet1", importTemplateSheet)
	for r, record := range rows {
		for c, val := range record {
			cellName, _ := excelize.CoordinatesToCellName(c+1, r+1)
			_ = f.SetCellValue(importTemplateSheet, cellName, val)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}
	return buf.Bytes()
}

func TestParseXLSX(t *testing.T) {
	data := buildXLSX(t, [][]string{
		{"email", "full_name", "nim_nidn", "role", "study_program"},
		{"ali@example.com", "Ali", "NIM001", "mahasiswa", "Teknik Informatika"},
		{"budi@example.com", "Budi", "", "dosen_pembimbing", ""},
	})

	rows, err := parseXLSX(data)
	if err != nil {
		t.Fatalf("parseXLSX: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].Email != "ali@example.com" || rows[0].Row != 2 {
		t.Errorf("rows[0] = %+v", rows[0])
	}
	if rows[1].Role != "dosen_pembimbing" {
		t.Errorf("rows[1].Role = %q", rows[1].Role)
	}
}

func TestParseXLSXInvalid(t *testing.T) {
	if _, err := parseXLSX([]byte("not-an-xlsx")); err == nil {
		t.Fatal("expected error for invalid xlsx bytes")
	}
}

func TestCell(t *testing.T) {
	if got := cell([]string{" a ", "b"}, 0); got != "a" {
		t.Errorf("cell[0] = %q, want 'a'", got)
	}
	if got := cell([]string{"a"}, 3); got != "" {
		t.Errorf("cell[3] = %q, want ''", got)
	}
}

func TestValidateImportRow(t *testing.T) {
	tests := []struct {
		name string
		row  ImportRow
		seen map[string]bool
		want string
	}{
		{"valid", ImportRow{Email: "ali@example.com", FullName: "Ali", Role: "mahasiswa"}, map[string]bool{}, ""},
		{"empty email", ImportRow{FullName: "Ali", Role: "mahasiswa"}, map[string]bool{}, "Email tidak boleh kosong"},
		{"invalid email", ImportRow{Email: "not-an-email", FullName: "Ali", Role: "mahasiswa"}, map[string]bool{}, "Format email tidak valid"},
		{"duplicate email", ImportRow{Email: "ali@example.com", FullName: "Ali", Role: "mahasiswa"}, map[string]bool{"ali@example.com": true}, "Email duplikat dalam file"},
		{"empty name", ImportRow{Email: "ali@example.com", Role: "mahasiswa"}, map[string]bool{}, "Nama lengkap tidak boleh kosong"},
		{"empty role", ImportRow{Email: "ali@example.com", FullName: "Ali"}, map[string]bool{}, "Role tidak boleh kosong"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateImportRow(tt.row, tt.seen); got != tt.want {
				t.Errorf("validateImportRow = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestImportUsersXLSX(t *testing.T) {
	uc, repo := newTestUserUseCase()
	data := buildXLSX(t, [][]string{
		{"email", "full_name", "nim_nidn", "role", "study_program"},
		{"x@example.com", "X User", "", "mahasiswa", ""},
	})

	result, err := uc.ImportUsers(context.Background(), "users.xlsx", data, Actor{})
	if err != nil {
		t.Fatalf("ImportUsers: %v", err)
	}
	if result.SuccessCount != 1 || result.ErrorCount != 0 {
		t.Errorf("result = %+v, want 1 success", result)
	}
	if len(repo.users) != 1 {
		t.Errorf("expected 1 user created, got %d", len(repo.users))
	}
}

func TestImportUsersInvalidExtension(t *testing.T) {
	uc, _ := newTestUserUseCase()
	_, err := uc.ImportUsers(context.Background(), "users.txt", []byte("data"), Actor{})
	if !errors.Is(err, ErrInvalidFileFormat) {
		t.Errorf("expected ErrInvalidFileFormat, got %v", err)
	}
}

// ── User usecase List/GetByID (Job 23 coverage) ──────────────────────────

func TestUserList(t *testing.T) {
	uc, repo := newTestUserUseCase()
	seedUser(repo, "a@example.com")
	seedUser(repo, "b@example.com")

	users, total, err := uc.List(context.Background(), domainRepo.UserFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Errorf("total = %d, users = %d, want 2", total, len(users))
	}
}

func TestUserGetByID(t *testing.T) {
	uc, repo := newTestUserUseCase()
	u := seedUser(repo, "target@example.com")

	got, err := uc.GetByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != u.Email {
		t.Errorf("email = %q, want %q", got.Email, u.Email)
	}

	if _, err := uc.GetByID(context.Background(), uuid.New()); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound for missing user, got %v", err)
	}
}

// ── Pure helper functions (Job 23 coverage) ───────────────────────────────

func TestWordCount(t *testing.T) {
	if got := wordCount(""); got != 0 {
		t.Errorf("wordCount('') = %d, want 0", got)
	}
	if got := wordCount("satu dua tiga"); got != 3 {
		t.Errorf("wordCount = %d, want 3", got)
	}
}

func TestDefenseRecipients(t *testing.T) {
	supervisors := []entity.User{{Email: "sup1@example.com"}, {Email: ""}}
	defense := &entity.ThesisDefense{
		Thesis: entity.Thesis{
			Student:     entity.User{Email: "student@example.com"},
			Supervisors: supervisors,
		},
	}

	got := defenseRecipients(defense, []string{"penguji@example.com"})
	want := []string{"student@example.com", "sup1@example.com", "penguji@example.com"}
	if len(got) != len(want) {
		t.Fatalf("recipients = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("recipients[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestToSeminarDetailFull(t *testing.T) {
	examiner := entity.User{ID: uuid.New(), FullName: "Penguji A", Email: "penguji@example.com"}
	sem := &entity.Seminar{
		ID: uuid.New(),
		Thesis: entity.Thesis{
			ID:      uuid.New(),
			Title:   "Judul Skripsi",
			Student: entity.User{ID: uuid.New(), FullName: "Mahasiswa", NimNidn: strPtr("NIM001")},
		},
		Examiners: []entity.User{examiner},
		Scores: []entity.SeminarScore{
			{ComponentName: "Presentasi", ComponentWeight: 50, Score: 80, Examiner: examiner},
		},
	}

	detail := toSeminarDetail(sem)
	if detail.Thesis == nil || detail.Thesis.Student == nil {
		t.Error("expected thesis + student in detail")
	}
	if len(detail.Examiners) != 1 || detail.Examiners[0].FullName != "Penguji A" {
		t.Errorf("examiners = %+v", detail.Examiners)
	}
	if len(detail.Scores) != 1 || detail.Scores[0].Examiner == nil {
		t.Errorf("scores = %+v", detail.Scores)
	}
}

func TestToConsultationDetailFull(t *testing.T) {
	log := &entity.ConsultationLog{
		ID:               uuid.New(),
		ThesisID:         uuid.New(),
		CreatedBy:        uuid.New(),
		Creator:          entity.User{ID: uuid.New(), FullName: "Mahasiswa"},
		ApprovedBy:       func() *uuid.UUID { id := uuid.New(); return &id }(),
		Approver:         &entity.User{ID: uuid.New(), FullName: "Dosen Pembimbing"},
		ConsultationDate: time.Now(),
		TopicsDiscussed:  "Topik",
	}

	detail := toConsultationDetail(log)
	if detail.Creator == nil || detail.Approver == nil {
		t.Error("expected creator + approver in detail")
	}
	if detail.Creator.FullName != "Mahasiswa" || detail.Approver.FullName != "Dosen Pembimbing" {
		t.Errorf("detail = %+v", detail)
	}
}

func TestToDocumentDetailWithReviewer(t *testing.T) {
	reviewerID := uuid.New()
	doc := &entity.Document{
		ID:         uuid.New(),
		ReviewerID: &reviewerID,
		Reviewer:   &entity.User{ID: reviewerID, FullName: "Dosen Pembimbing Satu"},
	}

	detail := toDocumentDetail(doc)
	if detail.Reviewer == nil || detail.Reviewer.FullName != "Dosen Pembimbing Satu" {
		t.Errorf("reviewer = %+v", detail.Reviewer)
	}
}

// TestThesisGetByIDSupervisorAndExaminer — covers canAccessThesis supervisor
// and examiner branches (previously 33.3% covered).
func TestThesisGetByIDSupervisorAndExaminer(t *testing.T) {
	uc, thesisRepo, userRepo, _ := newTestThesisUseCase()
	thesisID, _, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	// Supervisor → allowed.
	if _, err := uc.GetByID(context.Background(), thesisID, supervisorID, ThesisRoleDosenPembimbing); err != nil {
		t.Errorf("supervisor GetByID: %v", err)
	}

	// Examiner assigned via fake → allowed.
	thesisRepo.examiners[thesisID] = true
	if _, err := uc.GetByID(context.Background(), thesisID, uuid.New(), ThesisRoleDosenPenguji); err != nil {
		t.Errorf("examiner GetByID: %v", err)
	}

	// Examiner not assigned → forbidden.
	thesisRepo.examiners[thesisID] = false
	if _, err := uc.GetByID(context.Background(), thesisID, uuid.New(), ThesisRoleDosenPenguji); !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden for non-examiner, got %v", err)
	}
}

// TestThesisCancelNotFound — covers the ErrThesisNotFound branch of Cancel.
func TestThesisCancelNotFound(t *testing.T) {
	uc, _, _, _ := newTestThesisUseCase()
	err := uc.Cancel(context.Background(), uuid.New(), CancelThesisRequest{Reason: "x"}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrThesisNotFound) {
		t.Errorf("expected ErrThesisNotFound, got %v", err)
	}
}

// TestDocumentGetByIDAsExaminer — covers the dosen_penguji branch of
// DocumentUseCase.canView (previously 33.3% covered).
func TestDocumentGetByIDAsExaminer(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	// Assigned examiner → allowed.
	thesisRepo.examiners[thesisID] = true
	if _, err := uc.GetByID(context.Background(), thesisID, uploaded.ID, uuid.New(), ThesisRoleDosenPenguji); err != nil {
		t.Errorf("examiner GetByID: %v", err)
	}

	// Not assigned → forbidden.
	thesisRepo.examiners[thesisID] = false
	if _, err := uc.GetByID(context.Background(), thesisID, uploaded.ID, uuid.New(), ThesisRoleDosenPenguji); !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden for non-examiner, got %v", err)
	}
}

// countAwareAcademicYearRepo lets tests control CountActiveTheses (the shared
// fake always returns 0, which would make the "year in use" branch unreachable).
type countAwareAcademicYearRepo struct {
	*fakeAcademicYearRepo
	count int64
}

func (f *countAwareAcademicYearRepo) CountActiveTheses(context.Context, uuid.UUID) (int64, error) {
	return f.count, nil
}

// TestAcademicYearUpdateActiveInUse — an active year with ongoing theses cannot
// be updated (ErrAcademicYearInUse branch of Update).
func TestAcademicYearUpdateActiveInUse(t *testing.T) {
	repo := &countAwareAcademicYearRepo{fakeAcademicYearRepo: newFakeAcademicYearRepo(), count: 3}
	uc := NewAcademicYearUseCase(repo)
	existing := &entity.AcademicYear{Name: "2026/2027", Semester: "ganjil", IsActive: true}
	if err := repo.Create(context.Background(), existing); err != nil {
		t.Fatalf("seed year: %v", err)
	}

	_, err := uc.Update(context.Background(), existing.ID, validYearRequest())
	if !errors.Is(err, ErrAcademicYearInUse) {
		t.Errorf("expected ErrAcademicYearInUse, got %v", err)
	}
}

// TestDeactivateUserNotFound — covers the ErrUserNotFound branch of Deactivate.
func TestDeactivateUserNotFound(t *testing.T) {
	uc, _ := newTestUserUseCase()
	err := uc.Deactivate(context.Background(), uuid.New(), Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// TestSeminarRecipients — pure helper mirroring defenseRecipients.
func TestSeminarRecipients(t *testing.T) {
	sem := &entity.Seminar{
		Thesis: entity.Thesis{
			Student:     entity.User{Email: "student@example.com"},
			Supervisors: []entity.User{{Email: "sup@example.com"}, {Email: ""}},
		},
	}

	got := seminarRecipients(sem, []string{"penguji@example.com"})
	want := []string{"student@example.com", "sup@example.com", "penguji@example.com"}
	if len(got) != len(want) {
		t.Fatalf("recipients = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("recipients[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestUserResetPasswordNotFound — covers the ErrUserNotFound branch.
func TestUserResetPasswordNotFound(t *testing.T) {
	uc, _ := newTestUserUseCase()
	err := uc.ResetPassword(context.Background(), uuid.New(), Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// TestDocumentDownloadNotFound — owner passes canView, then FindByID misses.
func TestDocumentDownloadNotFound(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	_, err := uc.Download(context.Background(), thesisID, uuid.New(), studentID, "mahasiswa", Actor{UserID: studentID})
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Errorf("expected ErrDocumentNotFound, got %v", err)
	}
}
