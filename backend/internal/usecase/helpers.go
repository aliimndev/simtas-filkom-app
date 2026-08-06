package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

// ── Ownership & role helpers (Job 06) ───────────────────────────────────
//
// ThesisAccess bundles reusable access checks that operate on a thesis. It is
// shared by modules that work with a thesis: consultation logs (Job 06) now,
// seminars (Job 08) and defenses (Job 09) later.

// ThesisAccess provides thesis ownership/supervisor/examiner checks.
type ThesisAccess struct {
	thesisRepo domainRepo.ThesisRepository
}

// NewThesisAccess builds a ThesisAccess backed by the thesis repository.
func NewThesisAccess(thesisRepo domainRepo.ThesisRepository) *ThesisAccess {
	return &ThesisAccess{thesisRepo: thesisRepo}
}

// IsThesisOwner reports whether userID is the mahasiswa who owns the thesis.
func (a *ThesisAccess) IsThesisOwner(ctx context.Context, userID, thesisID uuid.UUID) (bool, error) {
	thesis, err := a.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		return false, err
	}
	return isThesisOwner(thesis, userID), nil
}

// IsSupervisor reports whether userID is one of the thesis's dosen pembimbing.
func (a *ThesisAccess) IsSupervisor(ctx context.Context, userID, thesisID uuid.UUID) (bool, error) {
	thesis, err := a.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		return false, err
	}
	return isSupervisor(thesis, userID), nil
}

// IsExaminer reports whether userID is an examiner of the thesis (linked via
// seminar_examiners or defense_examiners). Thesis-scoped for now. It is not
// used by Job 06 (consultation access is owner/supervisor/admin/kaprodi only)
// but is reserved for Jobs 08/09 (seminar & defense modules) per the job spec.
func (a *ThesisAccess) IsExaminer(ctx context.Context, userID, thesisID uuid.UUID) (bool, error) {
	return a.thesisRepo.IsExaminer(ctx, thesisID, userID)
}

// isThesisOwner is a pure predicate over an already-loaded thesis.
func isThesisOwner(thesis *entity.Thesis, userID uuid.UUID) bool {
	return thesis != nil && thesis.StudentID == userID
}

// userIDs extracts the IDs from a slice of users (used to fan out notifications).
func userIDs(users []entity.User) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids
}

// documentTypeLabel maps a document type code to a human-friendly Indonesian label.
func documentTypeLabel(t string) string {
	switch t {
	case entity.DocTypeProposal:
		return "Proposal"
	case entity.DocTypeDraftChapter:
		return "Draft Bab"
	case entity.DocTypeSeminarDoc:
		return "Dokumen Seminar"
	case entity.DocTypeDefenseDoc:
		return "Dokumen Sidang"
	case entity.DocTypeFinalThesis:
		return "Skripsi Final"
	case entity.DocTypeRevisionSheet:
		return "Lembar Revisi"
	case entity.DocTypeEndorsementLetter:
		return "Surat Pengesahan"
	default:
		return t
	}
}

// isSupervisor is a pure predicate over an already-loaded thesis.
func isSupervisor(thesis *entity.Thesis, userID uuid.UUID) bool {
	if thesis == nil {
		return false
	}
	for _, s := range thesis.Supervisors {
		if s.ID == userID {
			return true
		}
	}
	return false
}
