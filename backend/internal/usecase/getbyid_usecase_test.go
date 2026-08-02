package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// ── Document GetByID ──────────────────────────────────────────────────────

func TestDocumentGetByID_Owner(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	detail, err := uc.GetByID(context.Background(), thesisID, uploaded.ID, studentID, "mahasiswa")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.ID != uploaded.ID {
		t.Errorf("id = %v, want %v", detail.ID, uploaded.ID)
	}
}

func TestDocumentGetByID_NotFound(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	// Owner passes canView; then FindByID misses → ErrDocumentNotFound.
	_, err := uc.GetByID(context.Background(), thesisID, uuid.New(), studentID, "mahasiswa")
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Errorf("expected ErrDocumentNotFound, got %v", err)
	}
}

func TestDocumentGetByID_NotOwnerForbidden(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	thesisID, studentID, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "bab1.pdf", pdfContent)
	uploaded, err := uc.Upload(context.Background(), thesisID, file, header, UploadDocumentRequest{
		DocumentType: entity.DocTypeProposal,
	}, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	_, err = uc.GetByID(context.Background(), thesisID, uploaded.ID, uuid.New(), "mahasiswa")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// ── Seminar GetByID ──────────────────────────────────────────────────────

func TestSeminarGetByID(t *testing.T) {
	uc, semRepo, docRepo, thesisRepo, userRepo, _ := newTestSeminarUseCase(t)
	thesisID, studentID, _ := seedSeminarReadyThesis(t, semRepo, docRepo, thesisRepo, userRepo)

	sem, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	detail, err := uc.GetByID(context.Background(), sem.ID, studentID, "mahasiswa")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.ID != sem.ID {
		t.Errorf("id = %v, want %v", detail.ID, sem.ID)
	}
}

func TestSeminarGetByID_NotFound(t *testing.T) {
	uc, _, _, _, _, _ := newTestSeminarUseCase(t)
	_, err := uc.GetByID(context.Background(), uuid.New(), uuid.New(), "mahasiswa")
	if !errors.Is(err, ErrSeminarNotFound) {
		t.Errorf("expected ErrSeminarNotFound, got %v", err)
	}
}

// ── Archive GetByID ──────────────────────────────────────────────────────

func TestArchiveGetByID(t *testing.T) {
	uc, _, thesisRepo, userRepo, _ := newTestArchiveUseCase(t)
	thesisID, studentID := seedGraduatedThesis(t, thesisRepo, userRepo)

	file, header := makePDFHeader(t, "skripsi_final.pdf", pdfContent)
	arch, err := uc.Create(context.Background(), thesisID, file, header, validArchiveRequest(), Actor{UserID: studentID}, ThesisRoleMahasiswa)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := uc.GetByID(context.Background(), arch.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.ID != arch.ID {
		t.Errorf("id = %v, want %v", detail.ID, arch.ID)
	}
}

func TestArchiveGetByID_NotFound(t *testing.T) {
	uc, _, _, _, _ := newTestArchiveUseCase(t)
	_, err := uc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrArchiveNotFound) {
		t.Errorf("expected ErrArchiveNotFound, got %v", err)
	}
}

// ── Defense GetByID ──────────────────────────────────────────────────────

func TestDefenseGetByID(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, studentID, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)

	def, err := uc.Submit(context.Background(), thesisID, Actor{UserID: studentID})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	detail, err := uc.GetByID(context.Background(), def.ID, studentID, "mahasiswa")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.ID != def.ID {
		t.Errorf("id = %v, want %v", detail.ID, def.ID)
	}
}

func TestDefenseGetByID_NotFound(t *testing.T) {
	uc, _, _, _, _, _, _ := newTestDefenseUseCase(t)
	_, err := uc.GetByID(context.Background(), uuid.New(), uuid.New(), "mahasiswa")
	if !errors.Is(err, ErrDefenseNotFound) {
		t.Errorf("expected ErrDefenseNotFound, got %v", err)
	}
}
