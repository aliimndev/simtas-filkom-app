package email

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// StubEmailService is a placeholder email implementation used until Job 11.
// It logs the email to the console and records it in the email_logs table
// with status "sent" so the audit trail stays complete.
type StubEmailService struct {
	db *gorm.DB
}

func NewStubEmailService(db *gorm.DB) *StubEmailService {
	return &StubEmailService{db: db}
}

func (s *StubEmailService) SendWelcomeEmail(ctx context.Context, to, fullName, temporaryPassword string) error {
	subject := "[SIMTAS] Akun Anda Telah Dibuat"
	log.Printf("[email-stub][welcome] to=%s subject=%q (password sementara dikirim)", to, subject)
	return s.record(ctx, to, "welcome", subject)
}

func (s *StubEmailService) SendPasswordReset(ctx context.Context, to, fullName, newPassword string) error {
	subject := "[SIMTAS] Password Anda Telah Direset"
	log.Printf("[email-stub][password-reset] to=%s subject=%q (password baru dikirim)", to, subject)
	return s.record(ctx, to, "password_reset", subject)
}

func (s *StubEmailService) SendThesisSubmitted(ctx context.Context, to []string, thesis *entity.Thesis) error {
	subject := "[SIMTAS] Pengajuan Judul Skripsi Baru"
	for _, recipient := range to {
		log.Printf("[email-stub][thesis-submitted] to=%s subject=%q thesis=%s title=%q", recipient, subject, thesis.ID, thesis.Title)
		if err := s.record(ctx, recipient, "thesis_submitted", subject); err != nil {
			return err
		}
	}
	return nil
}

func (s *StubEmailService) SendThesisApproved(ctx context.Context, to string, thesis *entity.Thesis) error {
	subject := "[SIMTAS] Judul Skripsi Disetujui"
	log.Printf("[email-stub][thesis-approved] to=%s subject=%q thesis=%s title=%q", to, subject, thesis.ID, thesis.Title)
	return s.record(ctx, to, "thesis_approved", subject)
}

func (s *StubEmailService) SendThesisRejected(ctx context.Context, to string, thesis *entity.Thesis, notes string) error {
	subject := "[SIMTAS] Judul Skripsi Ditolak"
	log.Printf("[email-stub][thesis-rejected] to=%s subject=%q thesis=%s title=%q notes=%q", to, subject, thesis.ID, thesis.Title, notes)
	return s.record(ctx, to, "thesis_rejected", subject)
}

func (s *StubEmailService) SendSupervisorAssigned(ctx context.Context, studentEmail string, supervisorEmails []string, thesis *entity.Thesis) error {
	subject := "[SIMTAS] Dosen Pembimbing Telah Ditunjuk"
	log.Printf("[email-stub][supervisor-assigned] student=%s supervisors=%v thesis=%s title=%q", studentEmail, supervisorEmails, thesis.ID, thesis.Title)
	if err := s.record(ctx, studentEmail, "supervisor_assigned", subject); err != nil {
		return err
	}
	for _, recipient := range supervisorEmails {
		if err := s.record(ctx, recipient, "supervisor_assigned", subject); err != nil {
			return err
		}
	}
	return nil
}

// record writes an email_logs entry when a DB handle is available.
func (s *StubEmailService) record(ctx context.Context, to, eventType, subject string) error {
	if s.db == nil {
		return nil
	}
	entry := &entity.EmailLog{
		RecipientEmail: to,
		EventType:      eventType,
		Subject:        &subject,
		Status:         "sent",
		Provider:       "stub",
	}
	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		// Never fail the request because logging failed
		log.Printf("[email-stub] failed to record email_logs entry: %v", err)
	}
	return nil
}
