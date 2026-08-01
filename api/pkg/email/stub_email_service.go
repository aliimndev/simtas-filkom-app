package email

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/api/internal/domain/entity"
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
