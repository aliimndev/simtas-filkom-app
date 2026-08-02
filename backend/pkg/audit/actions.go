package audit

// Audit action constants — auth actions (Job 13).
const (
	ActionUserLogin       = "USER_LOGIN"
	ActionUserLoginFailed = "USER_LOGIN_FAILED"
	ActionUserLogout      = "USER_LOGOUT"
)

// Audit action constants — user management actions used by Job 04.
const (
	ActionUserCreated       = "USER_CREATED"
	ActionUserUpdated       = "USER_UPDATED"
	ActionUserDeleted       = "USER_DELETED"
	ActionUserActivated     = "USER_ACTIVATED"
	ActionUserDeactivated   = "USER_DEACTIVATED"
	ActionUserPasswordReset = "USER_PASSWORD_RESET"
	ActionUserBulkImported  = "USER_BULK_IMPORTED"
)

// Audit action constants — thesis workflow actions used by Job 05.
const (
	ActionThesisSubmitted    = "THESIS_SUBMITTED"
	ActionThesisApproved     = "THESIS_APPROVED"
	ActionThesisRejected     = "THESIS_REJECTED"
	ActionThesisCancelled    = "THESIS_CANCELLED"
	ActionSupervisorAssigned = "SUPERVISOR_ASSIGNED"
)

// Audit action constants — consultation actions used by Job 06.
const (
	ActionConsultationCreated  = "CONSULTATION_CREATED"
	ActionConsultationUpdated  = "CONSULTATION_UPDATED"
	ActionConsultationApproved = "CONSULTATION_APPROVED"
	ActionConsultationDeleted  = "CONSULTATION_DELETED"
)

// Audit action constants — document actions used by Job 07.
const (
	ActionDocumentUploaded   = "DOCUMENT_UPLOADED"
	ActionDocumentApproved   = "DOCUMENT_APPROVED"
	ActionDocumentRevision   = "DOCUMENT_REVISION_REQUESTED"
	ActionDocumentDownloaded = "DOCUMENT_DOWNLOADED"
)

// Audit action constants — seminar actions used by Job 08.
const (
	ActionSeminarSubmitted    = "SEMINAR_SUBMITTED"
	ActionSeminarScheduled    = "SEMINAR_SCHEDULED"
	ActionSeminarRescheduled  = "SEMINAR_RESCHEDULED"
	ActionSeminarScoreSubmit  = "SEMINAR_SCORE_SUBMITTED"
	ActionSeminarFinalized    = "SEMINAR_FINALIZED"
	ActionSeminarRevisionNote = "SEMINAR_REVISION_NOTED"
)

// Audit action constants — defense actions used by Job 09.
const (
	ActionDefenseSubmitted    = "DEFENSE_SUBMITTED"
	ActionDefenseScheduled    = "DEFENSE_SCHEDULED"
	ActionDefenseRescheduled  = "DEFENSE_RESCHEDULED"
	ActionDefenseScoreSubmit  = "DEFENSE_SCORE_SUBMITTED"
	ActionDefenseFinalized    = "DEFENSE_FINALIZED"
	ActionDefenseRevisionNote = "DEFENSE_REVISION_NOTED"
	ActionThesisGraduated     = "THESIS_GRADUATED"
)

// Audit action constants — archive actions used by Job 10.
const (
	ActionArchiveCreated    = "ARCHIVE_CREATED"
	ActionArchiveDownloaded = "ARCHIVE_DOWNLOADED"
)
