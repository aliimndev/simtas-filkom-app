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
