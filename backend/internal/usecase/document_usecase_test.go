package usecase

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
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

// fakeDocumentRepo is a minimal in-memory DocumentRepository for usecase tests.
type fakeDocumentRepo struct {
	docs          map[uuid.UUID]*entity.Document
	byThesis      map[uuid.UUID][]uuid.UUID
	approvedTypes map[string]bool // "thesisID:type" -> approved
	reviewCalls   []string
}

func newFakeDocumentRepo() *fakeDocumentRepo {
	return &fakeDocumentRepo{
		docs:          map[uuid.UUID]*entity.Document{},
		byThesis:      map[uuid.UUID][]uuid.UUID{},
		approvedTypes: map[string]bool{},
	}
}

func (f *fakeDocumentRepo) Create(_ context.Context, doc *entity.Document) error {
	doc.ID = uuid.New()
	f.docs[doc.ID] = doc
	f.byThesis[doc.ThesisID] = append(f.byThesis[doc.ThesisID], doc.ID)
	return nil
}

func (f *fakeDocumentRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.Document, error) {
	d, ok := f.docs[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *d
	// Mirror the real repo's Preload("Reviewer.Role"): populate the reviewer
	// association when the FK is set.
	if clone.ReviewerID != nil && clone.Reviewer == nil {
		reviewerID := *clone.ReviewerID
		clone.Reviewer = &entity.User{ID: reviewerID, FullName: "Dosen Pembimbing Satu"}
	}
	return &clone, nil
}

// latestPerType mirrors the real repository: newest version per (type, chapter).
func (f *fakeDocumentRepo) latestPerType(thesisID uuid.UUID) map[string]*entity.Document {
	latest := map[string]*entity.Document{}
	for _, id := range f.byThesis[thesisID] {
		d := f.docs[id]
		key := d.DocumentType
		if d.DocumentType == entity.DocTypeDraftChapter && d.ChapterNumber != nil {
			key += "/" + itoaTest(*d.ChapterNumber)
		}
		cur, ok := latest[key]
		if !ok || d.Version > cur.Version {
			latest[key] = d
		}
	}
	return latest
}

func (f *fakeDocumentRepo) FindByThesisID(_ context.Context, thesisID uuid.UUID, filter domainRepo.DocumentFilter) ([]*entity.Document, int64, error) {
	var out []*entity.Document
	for _, d := range f.latestPerType(thesisID) {
		if filter.DocumentType != "" && d.DocumentType != filter.DocumentType {
			continue
		}
		if filter.Status != "" && d.Status != filter.Status {
			continue
		}
		clone := *d
		out = append(out, &clone)
	}
	return out, int64(len(out)), nil
}

var _ domainRepo.DocumentRepository = (*fakeDocumentRepo)(nil)

func (f *fakeDocumentRepo) FindLatestByType(_ context.Context, thesisID uuid.UUID, docType string, chapterNum *int) (*entity.Document, error) {
	var latest *entity.Document
	for _, d := range f.latestPerType(thesisID) {
		if d.DocumentType != docType {
			continue
		}
		if docType == entity.DocTypeDraftChapter {
			if chapterNum == nil || d.ChapterNumber == nil || *d.ChapterNumber != *chapterNum {
				continue
			}
		}
		if latest == nil || d.Version > latest.Version {
			latest = d
		}
	}
	if latest == nil {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *latest
	return &clone, nil
}

func (f *fakeDocumentRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string, reviewerID uuid.UUID, notes string) error {
	d, ok := f.docs[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if d.Status != entity.DocStatusPendingReview {
		return gorm.ErrRecordNotFound
	}
	d.Status = status
	d.ReviewerID = &reviewerID
	d.Reviewer = &entity.User{ID: reviewerID, FullName: "Dosen Pembimbing Satu"}
	if notes != "" {
		d.ReviewerNotes = &notes
	}
	f.reviewCalls = append(f.reviewCalls, status)
	if status == entity.DocStatusApproved {
		f.approvedTypes[d.ThesisID.String()+":"+d.DocumentType] = true
	}
	return nil
}

func (f *fakeDocumentRepo) GetVersionHistory(_ context.Context, thesisID uuid.UUID, docType string, chapterNum *int) ([]*entity.Document, error) {
	var out []*entity.Document
	for _, id := range f.byThesis[thesisID] {
		d := f.docs[id]
		if d.DocumentType != docType {
			continue
		}
		if docType == entity.DocTypeDraftChapter {
			if chapterNum == nil || d.ChapterNumber == nil || *d.ChapterNumber != *chapterNum {
				continue
			}
		}
		clone := *d
		out = append(out, &clone)
	}
	// Newest first (mirror real repo).
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Version > out[i].Version {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

func (f *fakeDocumentRepo) IsDocumentApproved(_ context.Context, thesisID uuid.UUID, docType string) (bool, error) {
	return f.approvedTypes[thesisID.String()+":"+docType], nil
}

func itoaTest(v int) string {
	return string(rune('0' + v))
}

// recordingDocEmailService records document email recipients for assertions.
type recordingDocEmailService struct {
	uploadedTo chan []string
	reviewedTo chan string
}

func newRecordingDocEmailService() *recordingDocEmailService {
	return &recordingDocEmailService{
		uploadedTo: make(chan []string, 16),
		reviewedTo: make(chan string, 16),
	}
}

func (r *recordingDocEmailService) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (r *recordingDocEmailService) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
func (r *recordingDocEmailService) SendThesisSubmitted(context.Context, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingDocEmailService) SendThesisApproved(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingDocEmailService) SendThesisRejected(context.Context, string, *entity.Thesis, string) error {
	return nil
}
func (r *recordingDocEmailService) SendSupervisorAssigned(context.Context, string, []string, *entity.Thesis) error {
	return nil
}
func (r *recordingDocEmailService) SendConsultationCreated(context.Context, []string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingDocEmailService) SendConsultationApproved(context.Context, string, *entity.ConsultationLog) error {
	return nil
}
func (r *recordingDocEmailService) SendDocumentUploaded(_ context.Context, to []string, _ *entity.Document) error {
	r.uploadedTo <- to
	return nil
}
func (r *recordingDocEmailService) SendDocumentReviewed(_ context.Context, to string, _ *entity.Document, _ string) error {
	r.reviewedTo <- to
	return nil
}
func (r *recordingDocEmailService) SendSeminarSubmitted(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingDocEmailService) SendSeminarScheduled(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (r *recordingDocEmailService) SendSeminarFinalized(context.Context, string, *entity.Seminar) error {
	return nil
}
func (r *recordingDocEmailService) SendDefenseSubmitted(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingDocEmailService) SendDefenseScheduled(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingDocEmailService) SendDefenseFinalized(context.Context, string, *entity.ThesisDefense) error {
	return nil
}
func (r *recordingDocEmailService) SendGraduated(context.Context, string, *entity.Thesis) error {
	return nil
}
func (r *recordingDocEmailService) SendArchiveCreated(context.Context, string, *entity.ThesisArchive) error {
	return nil
}

func (r *recordingDocEmailService) SendTitleChangeRequested(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingDocEmailService) SendTitleChangeCancelled(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingDocEmailService) SendTitleChangeApproved(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingDocEmailService) SendTitleChangeRejected(context.Context, []string, *entity.Thesis, *entity.TitleChangeRequest) error {
	return nil
}

func (r *recordingDocEmailService) waitUploaded(t *testing.T) []string {
	t.Helper()
	select {
	case to := <-r.uploadedTo:
		return to
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for document-uploaded email")
		return nil
	}
}

func (r *recordingDocEmailService) waitReviewed(t *testing.T) string {
	t.Helper()
	select {
	case to := <-r.reviewedTo:
		return to
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for document-reviewed email")
		return ""
	}
}

// newTestDocumentUseCase wires a fresh document use case with in-memory fakes.
// The storage service writes to a temp dir so uploads are fully exercised.
func newTestDocumentUseCase(t *testing.T) (*DocumentUseCase, *fakeDocumentRepo, *fakeThesisRepo, *fakeUserRepo, *recordingDocEmailService) {
	t.Helper()
	docRepo := newFakeDocumentRepo()
	thesisRepo := newFakeThesisRepo()
	userRepo := newFakeUserRepo()
	email := newRecordingDocEmailService()

	userRepo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	userRepo.roles["dosen_pembimbing"] = &entity.Role{ID: 4, Name: "dosen_pembimbing"}

	auditSvc := audit.NewAuditService(nil)
	storageSvc := storage.NewStubStorageService(t.TempDir(), "http://test.local")
	uc := NewDocumentUseCase(docRepo, thesisRepo, storageSvc, email, auditSvc, nil)
	return uc, docRepo, thesisRepo, userRepo, email
}

// testFile implements multipart.File over an in-memory buffer.
type testFile struct {
	*bytes.Reader
}

func (testFile) Close() error { return nil }

// makePDFHeader builds a multipart.FileHeader + seekable reader for a PDF.
func makePDFHeader(t *testing.T, name string, content []byte) (multipart.File, *multipart.FileHeader) {
	t.Helper()
	header := &multipart.FileHeader{
		Filename: name,
		Size:     int64(len(content)),
	}
	return testFile{bytes.NewReader(content)}, header
}

var pdfContent = []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\ntrailer\n<<>>\n%%EOF")

func TestUploadDocumentByOwner(t *testing.T) {
	uc, docRepo, thesisRepo, userRepo, email := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	detail, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if detail.Status != entity.DocStatusPendingReview {
		t.Errorf("status = %q, want pending_review", detail.Status)
	}
	if detail.Version != 1 {
		t.Errorf("version = %d, want 1", detail.Version)
	}
	if len(docRepo.docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docRepo.docs))
	}
	// Student uploaded → email goes to the supervisors.
	to := email.waitUploaded(t)
	if len(to) != 1 || to[0] != "supervisor@example.com" {
		t.Errorf("email recipients = %v, want [supervisor@example.com]", to)
	}
}

func TestUploadDocumentVersionIncrements(t *testing.T) {
	uc, docRepo, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	for i := 1; i <= 2; i++ {
		file, header := makePDFHeader(t, "seminar.pdf", pdfContent)
		detail, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
			DocumentType: entity.DocTypeSeminarDoc,
		}, Actor{UserID: studentID})
		if err != nil {
			t.Fatalf("upload #%d: %v", i, err)
		}
		if detail.Version != i {
			t.Errorf("upload #%d version = %d, want %d", i, detail.Version, i)
		}
	}
	if len(docRepo.docs) != 2 {
		t.Errorf("expected 2 versions, got %d", len(docRepo.docs))
	}
	history, err := uc.History(context.Background(), thesisID, entity.DocTypeSeminarDoc, nil, studentID, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("history length = %d, want 2", len(history))
	}
	if history[0].Version != 2 || history[1].Version != 1 {
		t.Errorf("history order = [%d, %d], want [2, 1]", history[0].Version, history[1].Version)
	}
}

func TestUploadDocumentNonPDFRejected(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "dokumen.docx", []byte("not a pdf at all"))
	_, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if !errors.Is(err, utils.ErrNotPDF) {
		t.Errorf("expected ErrNotPDF, got %v", err)
	}
}

func TestUploadDocumentOversizeRejected(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	big := make([]byte, MaxDocumentSizeBytes+1)
	copy(big[:5], "%PDF-")
	file, header := makePDFHeader(t, "besar.pdf", big)
	_, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if !errors.Is(err, utils.ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestUploadDocumentForbiddenForOtherStudent(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	other := seedStudent(t, userRepo, "mahasiswa")

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	_, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: other})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUploadDocumentInvalidType(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "x.pdf", pdfContent)
	_, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: "skripsi_final", // not in ValidDocumentTypes
	}, Actor{UserID: studentID})
	if !errors.Is(err, ErrInvalidDocumentType) {
		t.Errorf("expected ErrInvalidDocumentType, got %v", err)
	}
}

func TestUploadDraftChapterRequiresChapterNumber(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	_, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeDraftChapter,
	}, Actor{UserID: studentID})
	if !errors.Is(err, ErrChapterNumberRequired) {
		t.Errorf("expected ErrChapterNumberRequired, got %v", err)
	}
}

func TestReviewDocumentApprovedBySupervisor(t *testing.T) {
	uc, docRepo, thesisRepo, userRepo, email := newTestDocumentUseCase(t)
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "seminar.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeSeminarDoc,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	detail, err := uc.Review(context.Background(), uploaded.ID, DocDecisionApproved, "Dokumen sudah baik", Actor{UserID: supervisorID})
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if detail.Status != entity.DocStatusApproved {
		t.Errorf("status = %q, want approved", detail.Status)
	}
	if detail.Reviewer == nil || detail.Reviewer.ID != supervisorID {
		t.Errorf("reviewer not set correctly")
	}

	// Supervisor reviewed → email goes to the student.
	student := userRepo.users[studentID]
	to := email.waitReviewed(t)
	if to != student.Email {
		t.Errorf("review email to = %q, want %q", to, student.Email)
	}

	// Gate now opens for seminar submission.
	ok, err := uc.CanSubmitSeminar(context.Background(), thesisID)
	if err != nil || !ok {
		t.Errorf("CanSubmitSeminar = %v (err=%v), want true", ok, err)
	}
	_ = docRepo
}

func TestReviewDocumentByNonSupervisorForbidden(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	otherDosen := seedOtherDosen(t, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	_, err = uc.Review(context.Background(), uploaded.ID, DocDecisionApproved, "ok", Actor{UserID: otherDosen})
	if !errors.Is(err, ErrNotDocumentReviewer) {
		t.Errorf("expected ErrNotDocumentReviewer, got %v", err)
	}
}

func TestReviewDocumentAlreadyReviewedRejected(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if _, err := uc.Review(context.Background(), uploaded.ID, DocDecisionApproved, "ok", Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("first review: %v", err)
	}

	_, err = uc.Review(context.Background(), uploaded.ID, DocDecisionApproved, "lagi", Actor{UserID: supervisorID})
	if !errors.Is(err, ErrDocumentNotPending) {
		t.Errorf("expected ErrDocumentNotPending, got %v", err)
	}
}

func TestListDocumentsAdminAccess(t *testing.T) {
	uc, docRepo, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	if _, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID}); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	docs, total, err := uc.List(context.Background(), thesisID, domainRepo.DocumentFilter{}, uuid.New(), ThesisRoleAdminFakultas)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(docs) != 1 {
		t.Errorf("expected 1 doc, got total=%d len=%d", total, len(docs))
	}
	_ = docRepo
}

func TestListDocumentsForbiddenForOtherStudent(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	other := seedStudent(t, userRepo, "mahasiswa")

	_, _, err := uc.List(context.Background(), thesisID, domainRepo.DocumentFilter{}, other, ThesisRoleMahasiswa)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestDownloadDocumentGeneratesURL(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	result, err := uc.Download(context.Background(), thesisID, uploaded.ID, studentID, ThesisRoleMahasiswa, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.DownloadURL == "" {
		t.Error("download_url should not be empty")
	}
	if result.ExpiresIn != DocumentDownloadExpirySecs {
		t.Errorf("expires_in = %d, want %d", result.ExpiresIn, DocumentDownloadExpirySecs)
	}
}

func TestCanSubmitDefenseGate(t *testing.T) {
	uc, docRepo, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, supervisorID := seedInProgressThesis(t, thesisRepo, userRepo)

	ok, err := uc.CanSubmitDefense(context.Background(), thesisID)
	if err != nil || ok {
		t.Errorf("CanSubmitDefense before approval = %v (err=%v), want false", ok, err)
	}

	file, header := makePDFHeader(t, "sidang.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeDefenseDoc,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if _, err := uc.Review(context.Background(), uploaded.ID, DocDecisionApproved, "ok", Actor{UserID: supervisorID}); err != nil {
		t.Fatalf("Review: %v", err)
	}

	ok, err = uc.CanSubmitDefense(context.Background(), thesisID)
	if err != nil || !ok {
		t.Errorf("CanSubmitDefense after approval = %v (err=%v), want true", ok, err)
	}
	_ = docRepo
}
