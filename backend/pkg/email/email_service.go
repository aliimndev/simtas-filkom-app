package email

import (
	"context"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// EmailService defines the email operations used by the application.
// Implementations: ResendEmailService (production, Job 11) and the recording
// fakes used in unit tests. EMAIL_DEV_MODE routes through ResendEmailService
// in console-only mode.
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

	// ── Consultation notifications (Job 06) ──────────────────────────────
	// SendConsultationCreated notifies the recipient(s) about a new consultation log
	// (supervisors when a student creates; the student when a supervisor creates).
	SendConsultationCreated(ctx context.Context, to []string, log *entity.ConsultationLog) error
	// SendConsultationApproved notifies the student that a consultation log was approved.
	SendConsultationApproved(ctx context.Context, to string, log *entity.ConsultationLog) error

	// ── Document notifications (Job 07) ─────────────────────────────────
	// SendDocumentUploaded notifies the supervisors that a student uploaded a document.
	SendDocumentUploaded(ctx context.Context, to []string, doc *entity.Document) error
	// SendDocumentReviewed notifies the student about the review decision.
	SendDocumentReviewed(ctx context.Context, to string, doc *entity.Document, decision string) error

	// ── Seminar notifications (Job 08) ──────────────────────────────────
	// SendSeminarSubmitted notifies kaprodi + admin about a seminar proposal submission.
	SendSeminarSubmitted(ctx context.Context, to []string, seminar *entity.Seminar) error
	// SendSeminarScheduled notifies student, supervisors, and examiners about a schedule.
	SendSeminarScheduled(ctx context.Context, to []string, seminar *entity.Seminar) error
	// SendSeminarFinalized notifies the student about the final seminar result.
	SendSeminarFinalized(ctx context.Context, to string, seminar *entity.Seminar) error

	// ── Defense notifications (Job 09) ──────────────────────────────────
	// SendDefenseSubmitted notifies kaprodi + admin about a defense submission.
	SendDefenseSubmitted(ctx context.Context, to []string, defense *entity.ThesisDefense) error
	// SendDefenseScheduled notifies student, supervisors, and examiners about a schedule.
	SendDefenseScheduled(ctx context.Context, to []string, defense *entity.ThesisDefense) error
	// SendDefenseFinalized notifies the student about the final defense result.
	SendDefenseFinalized(ctx context.Context, to string, defense *entity.ThesisDefense) error
	// SendGraduated notifies the student that their thesis was declared graduated.
	SendGraduated(ctx context.Context, to string, thesis *entity.Thesis) error

	// ── Archive notifications (Job 10) ──────────────────────────────────
	// SendArchiveCreated notifies the student that their thesis archive is available.
	SendArchiveCreated(ctx context.Context, to string, archive *entity.ThesisArchive) error
}
