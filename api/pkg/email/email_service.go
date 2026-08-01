package email

import "context"

// EmailService defines the email operations used by the application.
// The stub implementation in this job (Job 04) logs emails; Job 11 replaces
// it with the real Resend implementation without changing this interface.
type EmailService interface {
	// SendWelcomeEmail sends an account-creation notification with a temporary password.
	SendWelcomeEmail(ctx context.Context, to, fullName, temporaryPassword string) error
	// SendPasswordReset sends a password-reset notification with a new password.
	SendPasswordReset(ctx context.Context, to, fullName, newPassword string) error
}
