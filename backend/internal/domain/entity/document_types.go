package entity

// Document type constants (Job 07). Kept in sync with the DB check constraint
// in 000007_create_documents_table.up.sql.
const (
	DocTypeProposal          = "proposal"
	DocTypeDraftChapter      = "draft_chapter" // chapter_number: 1-5
	DocTypeSeminarDoc        = "seminar_doc"
	DocTypeDefenseDoc        = "defense_doc"
	DocTypeFinalThesis       = "final_thesis"
	DocTypeRevisionSheet     = "revision_sheet"
	DocTypeEndorsementLetter = "endorsement_letter"
)

// Document status values (Job 07). Kept in sync with the DB check constraint.
const (
	DocStatusPendingReview    = "pending_review"
	DocStatusApproved         = "approved"
	DocStatusRevisionRequired = "revision_required"
)

// ValidDocumentTypes lists every accepted document type.
var ValidDocumentTypes = []string{
	DocTypeProposal,
	DocTypeDraftChapter,
	DocTypeSeminarDoc,
	DocTypeDefenseDoc,
	DocTypeFinalThesis,
	DocTypeRevisionSheet,
	DocTypeEndorsementLetter,
}

// ValidDocumentType reports whether t is a known document type.
func ValidDocumentType(t string) bool {
	for _, v := range ValidDocumentTypes {
		if v == t {
			return true
		}
	}
	return false
}

// SeminarGate lists the documents that must be approved before a student may
// submit a seminar proposal.
var SeminarGate = []string{DocTypeSeminarDoc}

// DefenseGate lists the documents that must be approved before a student may
// submit a defense request.
var DefenseGate = []string{DocTypeDefenseDoc}
