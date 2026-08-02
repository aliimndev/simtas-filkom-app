package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// newGateDocUseCase wires a DocumentUseCase with the shared document fakes.
func newGateDocUseCase(t *testing.T) (*DocumentUseCase, *fakeDocumentRepo, *fakeThesisRepo, *fakeUserRepo) {
	t.Helper()
	uc, docRepo, thesisRepo, userRepo, _ := newTestDocumentUseCase(t)
	return uc, docRepo, thesisRepo, userRepo
}

// TestCanSubmitSeminar_NotApproved — Job 23: seminar_doc belum approved → false.
func TestCanSubmitSeminar_NotApproved(t *testing.T) {
	uc, _, thesisRepo, userRepo := newGateDocUseCase(t)
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	ok, err := uc.CanSubmitSeminar(context.Background(), thesisID)
	if err != nil {
		t.Fatalf("CanSubmitSeminar: %v", err)
	}
	if ok {
		t.Error("CanSubmitSeminar should be false when seminar_doc is not approved")
	}
}

// TestCanSubmitSeminar_Approved — Job 23: seminar_doc approved → true.
func TestCanSubmitSeminar_Approved(t *testing.T) {
	uc, docRepo, thesisRepo, userRepo := newGateDocUseCase(t)
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)

	docRepo.approvedTypes[thesisID.String()+":"+entity.DocTypeSeminarDoc] = true

	ok, err := uc.CanSubmitSeminar(context.Background(), thesisID)
	if err != nil {
		t.Fatalf("CanSubmitSeminar: %v", err)
	}
	if !ok {
		t.Error("CanSubmitSeminar should be true when seminar_doc is approved")
	}
}

// TestCanSubmitDefense_SeminarNotDone — Job 23: seminar belum passed (thesis
// belum seminar_done) → false dengan reason "seminar".
func TestCanSubmitDefense_SeminarNotDone(t *testing.T) {
	uc, _, _, _, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	// thesis stays "in_progress" — seminar milestone not reached.

	canSubmit, reason, err := uc.CanSubmitDefense(context.Background(), thesisID)
	if err != nil {
		t.Fatalf("CanSubmitDefense: %v", err)
	}
	if canSubmit {
		t.Error("CanSubmitDefense should be false before seminar_done")
	}
	if reason != "seminar" {
		t.Errorf("reason = %q, want %q", reason, "seminar")
	}
}

// TestCanSubmitDefense_SeminarDoneDocNotApproved — thesis seminar_done tapi
// defense_doc belum approved → false dengan reason "defense_doc".
func TestCanSubmitDefense_SeminarDoneDocNotApproved(t *testing.T) {
	uc, _, _, _, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, _, _ := seedInProgressThesis(t, thesisRepo, userRepo)
	thesisRepo.theses[thesisID].Status = "seminar_done"

	canSubmit, reason, err := uc.CanSubmitDefense(context.Background(), thesisID)
	if err != nil {
		t.Fatalf("CanSubmitDefense: %v", err)
	}
	if canSubmit {
		t.Error("CanSubmitDefense should be false when defense_doc not approved")
	}
	if reason != "defense_doc" {
		t.Errorf("reason = %q, want %q", reason, "defense_doc")
	}
}

// TestCanSubmitDefense_AllGatesMet — Job 23: seminar passed + defense_doc
// approved → true.
func TestCanSubmitDefense_AllGatesMet(t *testing.T) {
	uc, defRepo, _, docRepo, thesisRepo, userRepo, _ := newTestDefenseUseCase(t)
	thesisID, _, _ := seedDefenseReadyThesis(t, defRepo, docRepo, thesisRepo, userRepo)

	canSubmit, reason, err := uc.CanSubmitDefense(context.Background(), thesisID)
	if err != nil {
		t.Fatalf("CanSubmitDefense: %v", err)
	}
	if !canSubmit {
		t.Error("CanSubmitDefense should be true when all gates are met")
	}
	if reason != "" {
		t.Errorf("reason = %q, want empty", reason)
	}
}

// TestCanSubmitDefense_ThesisNotFound — missing thesis surfaces the sentinel.
func TestCanSubmitDefense_ThesisNotFound(t *testing.T) {
	uc, _, _, _, _, _, _ := newTestDefenseUseCase(t)

	_, _, err := uc.CanSubmitDefense(context.Background(), uuid.New())
	if !errors.Is(err, ErrThesisNotFound) {
		t.Errorf("expected ErrThesisNotFound, got %v", err)
	}
}
