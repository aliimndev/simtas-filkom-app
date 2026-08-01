package audit

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
