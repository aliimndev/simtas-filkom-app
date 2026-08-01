package email

import (
	"context"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// EmailService defines the email operations used by the application.
// The stub implementation (Job 04) logs emails; Job 11 replaces it with the
// real Resend implementation without changing this interface.
type EmailService interface {
	// SendWelcomeEmail sends an account-creation notification with a temporary password.
	SendWelcomeEmail(ctx context.Context, to, fullName, temporaryPassword string) error
	// SendPasswordReset sends a password-reset notification with a new password.
	SendPasswordReset(ctx context.Context, to, fullName, newPassword string) error

	// ── Thesis notifications (Job 05) ────────────────────────────────────
	// SendThesisSubmitted notifies all kaprodi about a new thesis submission.
	SendThesisSubmitted(ctx context.Context, to []string, thesis *entity.Thesis) error
	// SendThesisApproved notifies the student that their title was approved.
	SendThesisApproved(ctx context.Context, to string, thesis *entity.Thesis) error
	// SendThesisRejected notifies the student that their title was rejected.
	SendThesisRejected(ctx context.Context, to string, thesis *entity.Thesis, notes string) error
	// SendSupervisorAssigned notifies the student and the newly assigned supervisors.
	SendSupervisorAssigned(ctx context.Context, studentEmail string, supervisorEmails []string, thesis *entity.Thesis) error
}
