package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/resend/resend-go/v2"
	"gorm.io/gorm"
)

//go:embed templates/*.html
var templateFS embed.FS

// TemplateData carries the dynamic values rendered into the email templates.
type TemplateData struct {
	AppName     string
	FrontendURL string
	Year        int
	Title       string   // heading + HTML <title>
	Greeting    string   // opening line, e.g. "Halo Budi,"
	Message     string   // body paragraph
	Status      string   // badge text, e.g. DISETUJUI / PERLU REVISI / LULUS
	StatusGood  bool     // badge color (green when true, red otherwise)
	Details     []Detail // label/value rows
	Notes       string   // optional notes block
	CTA         *CTA     // action button
}

// Detail is a single label/value row in the details table.
type Detail struct {
	Label string
	Value string
}

// CTA is the call-to-action button shown in the email.
type CTA struct {
	Label string
	URL   string
}

// ResendEmailService implements EmailService using the Resend provider.
// When devMode is true no email is sent — the rendered message is logged to
// the console and recorded in email_logs with provider "dev".
type ResendEmailService struct {
	client      *resend.Client
	fromEmail   string
	fromName    string
	frontendURL string
	db          *gorm.DB
	devMode     bool
	templates   map[string]*template.Template
}

// NewResendEmailService builds a Resend-backed service. Pass an empty apiKey
// together with devMode=true to run in console-only mode without a client.
// Templates are parsed once at construction and cached for subsequent sends.
func NewResendEmailService(apiKey, fromEmail, fromName, frontendURL string, db *gorm.DB, devMode bool) *ResendEmailService {
	s := &ResendEmailService{
		fromEmail:   fromEmail,
		fromName:    fromName,
		frontendURL: frontendURL,
		db:          db,
		devMode:     devMode,
		templates:   loadTemplates(),
	}
	if !devMode && apiKey != "" {
		s.client = resend.NewClient(apiKey)
	}
	return s
}

// ── Account notifications ───────────────────────────────────────────────

func (s *ResendEmailService) SendWelcomeEmail(ctx context.Context, to, fullName, temporaryPassword string) error {
	subject := "[SIMTAS] Akun Anda Telah Dibuat"
	data := s.baseData("Akun Anda Telah Dibuat")
	data.Greeting = "Halo " + fullName + ","
	data.Message = "Akun SIMTAS Anda telah dibuat. Gunakan kredensial berikut untuk masuk:"
	data.Details = []Detail{
		{Label: "Email", Value: to},
		{Label: "Password Sementara", Value: temporaryPassword},
	}
	data.Notes = "Untuk keamanan, silakan ganti password Anda setelah login pertama."
	data.CTA = &CTA{Label: "Masuk ke SIMTAS", URL: s.frontendURL + "/login"}
	return s.send([]string{to}, subject, "event.html", "welcome", data)
}

func (s *ResendEmailService) SendPasswordReset(ctx context.Context, to, fullName, newPassword string) error {
	subject := "[SIMTAS] Password Anda Telah Direset"
	data := s.baseData("Password Anda Telah Direset")
	data.Greeting = "Halo " + fullName + ","
	data.Message = "Password akun SIMTAS Anda telah direset. Gunakan password baru berikut untuk masuk:"
	data.Details = []Detail{
		{Label: "Email", Value: to},
		{Label: "Password Baru", Value: newPassword},
	}
	data.Notes = "Segera ganti password ini setelah login."
	data.CTA = &CTA{Label: "Masuk ke SIMTAS", URL: s.frontendURL + "/login"}
	return s.send([]string{to}, subject, "event.html", "password_reset", data)
}

// ── Thesis notifications (Job 05) ───────────────────────────────────────

func (s *ResendEmailService) SendThesisSubmitted(ctx context.Context, to []string, thesis *entity.Thesis) error {
	subject := fmt.Sprintf("[SIMTAS] Pengajuan Judul Baru — %s", thesis.Student.FullName)
	data := s.baseData("Pengajuan Judul Skripsi Baru")
	data.Details = []Detail{
		{Label: "Nama Mahasiswa", Value: thesis.Student.FullName},
		{Label: "NIM", Value: deref(thesis.Student.NimNidn)},
		{Label: "Judul", Value: thesis.Title},
		{Label: "Tanggal Pengajuan", Value: thesis.SubmittedAt.Format("02 Jan 2006")},
	}
	if thesis.Abstract != nil {
		data.Notes = truncate(*thesis.Abstract, 220)
	}
	data.CTA = &CTA{Label: "Review Pengajuan", URL: s.frontendURL + "/kaprodi/theses/" + thesis.ID.String()}
	return s.send(to, subject, "thesis_submitted.html", "thesis_submitted", data)
}

func (s *ResendEmailService) SendThesisApproved(ctx context.Context, to string, thesis *entity.Thesis) error {
	subject := "[SIMTAS] Judul Skripsi Anda Disetujui"
	data := s.baseData("Judul Skripsi Disetujui")
	data.Status = "DISETUJUI"
	data.StatusGood = true
	data.Greeting = "Halo " + thesis.Student.FullName + ","
	data.Message = "Selamat! Judul skripsi Anda telah disetujui oleh Kaprodi."
	data.Details = []Detail{{Label: "Judul", Value: thesis.Title}}
	if thesis.KaprodiNotes != nil {
		data.Notes = *thesis.KaprodiNotes
	}
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/mahasiswa/theses/" + thesis.ID.String()}
	return s.send([]string{to}, subject, "thesis_reviewed.html", "thesis_approved", data)
}

func (s *ResendEmailService) SendThesisRejected(ctx context.Context, to string, thesis *entity.Thesis, notes string) error {
	subject := "[SIMTAS] Judul Skripsi Anda Perlu Revisi"
	data := s.baseData("Judul Skripsi Perlu Revisi")
	data.Status = "PERLU REVISI"
	data.StatusGood = false
	data.Greeting = "Halo " + thesis.Student.FullName + ","
	data.Message = "Judul skripsi Anda belum dapat disetujui. Silakan lakukan perbaikan sesuai catatan Kaprodi."
	data.Details = []Detail{{Label: "Judul", Value: thesis.Title}}
	data.Notes = notes
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/mahasiswa/theses/" + thesis.ID.String()}
	return s.send([]string{to}, subject, "thesis_reviewed.html", "thesis_rejected", data)
}

func (s *ResendEmailService) SendSupervisorAssigned(ctx context.Context, studentEmail string, supervisorEmails []string, thesis *entity.Thesis) error {
	subject := "[SIMTAS] Dosen Pembimbing Telah Ditetapkan"

	var supNames, supContacts []string
	for _, sup := range thesis.Supervisors {
		supNames = append(supNames, sup.FullName)
		supContacts = append(supContacts, fmt.Sprintf("%s (%s / %s)", sup.FullName, deref(sup.NimNidn), sup.Email))
	}

	// → Mahasiswa: daftar pembimbing + kontak
	studentData := s.baseData("Dosen Pembimbing Telah Ditetapkan")
	studentData.Greeting = "Halo " + thesis.Student.FullName + ","
	studentData.Message = "Kaprodi telah menetapkan dosen pembimbing untuk skripsi Anda:"
	studentData.Details = []Detail{
		{Label: "Dosen Pembimbing", Value: strings.Join(supNames, ", ")},
		{Label: "Judul Skripsi", Value: thesis.Title},
	}
	if len(supContacts) > 0 {
		studentData.Notes = "Kontak dosen pembimbing: " + strings.Join(supContacts, "; ")
	}
	studentData.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/mahasiswa/theses/" + thesis.ID.String()}
	if err := s.send([]string{studentEmail}, subject, "supervisor_assigned.html", "supervisor_assigned", studentData); err != nil {
		return err
	}

	// → Dosen pembimbing: info mahasiswa
	supervisorData := s.baseData("Dosen Pembimbing Telah Ditetapkan")
	supervisorData.Greeting = "Halo,"
	supervisorData.Message = "Anda ditetapkan sebagai dosen pembimbing untuk mahasiswa berikut:"
	supervisorData.Details = []Detail{
		{Label: "Nama Mahasiswa", Value: thesis.Student.FullName},
		{Label: "NIM", Value: deref(thesis.Student.NimNidn)},
		{Label: "Judul Skripsi", Value: thesis.Title},
	}
	supervisorData.Notes = "Mulai lakukan bimbingan dan catat log konsultasi di SIMTAS."
	supervisorData.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/dosen/theses/" + thesis.ID.String()}
	return s.send(supervisorEmails, subject, "supervisor_assigned.html", "supervisor_assigned", supervisorData)
}

// ── Consultation notifications (Job 06) ─────────────────────────────────

func (s *ResendEmailService) SendConsultationCreated(ctx context.Context, to []string, entry *entity.ConsultationLog) error {
	subject := "[SIMTAS] Log Konsultasi Baru Tercatat"
	data := s.baseData("Log Konsultasi Baru")
	data.Message = "Sebuah log konsultasi baru telah dicatat dan menunggu persetujuan."
	data.Details = []Detail{
		{Label: "Tanggal Konsultasi", Value: entry.ConsultationDate.Format("02 Jan 2006")},
		{Label: "Topik", Value: entry.TopicsDiscussed},
		{Label: "Catatan", Value: deref(entry.Notes)},
	}
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/theses/" + entry.ThesisID.String() + "/consultations"}
	return s.send(to, subject, "event.html", "consultation_created", data)
}

func (s *ResendEmailService) SendConsultationApproved(ctx context.Context, to string, entry *entity.ConsultationLog) error {
	subject := "[SIMTAS] Log Konsultasi Disetujui"
	data := s.baseData("Log Konsultasi Disetujui")
	data.Status = "DISETUJUI"
	data.StatusGood = true
	data.Message = "Log konsultasi Anda telah disetujui oleh dosen pembimbing."
	data.Details = []Detail{
		{Label: "Tanggal Konsultasi", Value: entry.ConsultationDate.Format("02 Jan 2006")},
		{Label: "Topik", Value: entry.TopicsDiscussed},
	}
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/theses/" + entry.ThesisID.String() + "/consultations"}
	return s.send([]string{to}, subject, "event.html", "consultation_approved", data)
}

// ── Document notifications (Job 07) ─────────────────────────────────────

func (s *ResendEmailService) SendDocumentUploaded(ctx context.Context, to []string, doc *entity.Document) error {
	subject := "[SIMTAS] Dokumen Baru Diunggah"
	data := s.baseData("Dokumen Baru Diunggah")
	data.Message = "Sebuah dokumen baru telah diunggah dan menunggu review Anda."
	data.Details = []Detail{
		{Label: "Jenis Dokumen", Value: doc.DocumentType},
		{Label: "Versi", Value: strconv.Itoa(doc.Version)},
		{Label: "Nama File", Value: doc.FileName},
	}
	data.CTA = &CTA{Label: "Tinjau Dokumen", URL: s.frontendURL + "/theses/" + doc.ThesisID.String() + "/documents"}
	return s.send(to, subject, "event.html", "document_uploaded", data)
}

func (s *ResendEmailService) SendDocumentReviewed(ctx context.Context, to string, doc *entity.Document, decision string) error {
	approved := decision == "approved"
	subject := "[SIMTAS] Dokumen Disetujui"
	if !approved {
		subject = "[SIMTAS] Dokumen Perlu Revisi"
	}
	data := s.baseData("Review Dokumen")
	data.Status = "DISETUJUI"
	data.StatusGood = true
	if !approved {
		data.Status = "PERLU REVISI"
		data.StatusGood = false
	}
	data.Message = "Dokumen Anda telah direview oleh dosen pembimbing."
	data.Details = []Detail{
		{Label: "Jenis Dokumen", Value: doc.DocumentType},
		{Label: "Versi", Value: strconv.Itoa(doc.Version)},
	}
	data.Notes = deref(doc.ReviewerNotes)
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/theses/" + doc.ThesisID.String() + "/documents"}
	return s.send([]string{to}, subject, "event.html", "document_reviewed", data)
}

// ── Seminar notifications (Job 08) ──────────────────────────────────────

func (s *ResendEmailService) SendSeminarSubmitted(ctx context.Context, to []string, seminar *entity.Seminar) error {
	subject := "[SIMTAS] Pengajuan Seminar Proposal Baru"
	data := s.baseData("Pengajuan Seminar Proposal Baru")
	data.Details = []Detail{
		{Label: "Nama Mahasiswa", Value: seminar.Thesis.Student.FullName},
		{Label: "Judul", Value: seminar.Thesis.Title},
	}
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/seminars/" + seminar.ID.String()}
	return s.send(to, subject, "event.html", "seminar_submitted", data)
}

func (s *ResendEmailService) SendSeminarScheduled(ctx context.Context, to []string, seminar *entity.Seminar) error {
	subject := fmt.Sprintf("[SIMTAS] Jadwal Seminar Proposal — %s", seminar.Thesis.Student.FullName)
	data := s.baseData("Jadwal Seminar Proposal")
	data.Greeting = "Halo,"
	data.Message = "Jadwal seminar proposal telah ditetapkan:"
	data.Details = []Detail{
		{Label: "Mahasiswa", Value: seminar.Thesis.Student.FullName},
		{Label: "Judul", Value: seminar.Thesis.Title},
		{Label: "Tanggal & Waktu", Value: formatTime(seminar.ScheduledAt)},
		{Label: "Ruangan", Value: deref(seminar.Room)},
		{Label: "Dosen Penguji", Value: joinUserNames(seminar.Examiners)},
	}
	data.Notes = deref(seminar.Notes)
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/seminars/" + seminar.ID.String()}
	return s.send(to, subject, "seminar_scheduled.html", "seminar_scheduled", data)
}

func (s *ResendEmailService) SendSeminarFinalized(ctx context.Context, to string, seminar *entity.Seminar) error {
	passed := seminar.Status == "passed"
	subject := "[SIMTAS] Hasil Seminar Proposal"
	if passed {
		subject = "[SIMTAS] Selamat! Anda Lulus Seminar Proposal"
	}
	data := s.baseData("Hasil Seminar Proposal")
	data.Greeting = "Halo " + seminar.Thesis.Student.FullName + ","
	data.Status = "LULUS"
	data.StatusGood = true
	if !passed {
		data.Status = "TIDAK LULUS"
		data.StatusGood = false
	}
	data.Message = "Hasil seminar proposal Anda telah dirilis:"
	data.Details = []Detail{{Label: "Nilai Akhir", Value: formatScore(seminar.FinalScore)}}
	if !passed {
		data.Notes = "Silakan lakukan revisi dan persiapkan diri untuk mengikuti seminar kembali."
	}
	data.CTA = &CTA{Label: "Lihat Hasil", URL: s.frontendURL + "/seminars/" + seminar.ID.String()}
	return s.send([]string{to}, subject, "seminar_result.html", "seminar_finalized", data)
}

// ── Defense notifications (Job 09) ──────────────────────────────────────

func (s *ResendEmailService) SendDefenseSubmitted(ctx context.Context, to []string, defense *entity.ThesisDefense) error {
	subject := "[SIMTAS] Pengajuan Sidang Skripsi Baru"
	data := s.baseData("Pengajuan Sidang Skripsi Baru")
	data.Details = []Detail{
		{Label: "Nama Mahasiswa", Value: defense.Thesis.Student.FullName},
		{Label: "Judul", Value: defense.Thesis.Title},
	}
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/defenses/" + defense.ID.String()}
	return s.send(to, subject, "event.html", "defense_submitted", data)
}

func (s *ResendEmailService) SendDefenseScheduled(ctx context.Context, to []string, defense *entity.ThesisDefense) error {
	subject := fmt.Sprintf("[SIMTAS] Jadwal Sidang Skripsi — %s", defense.Thesis.Student.FullName)
	data := s.baseData("Jadwal Sidang Skripsi")
	data.Greeting = "Halo,"
	data.Message = "Jadwal sidang skripsi Anda telah ditetapkan:"
	data.Details = []Detail{
		{Label: "Mahasiswa", Value: defense.Thesis.Student.FullName},
		{Label: "Judul", Value: defense.Thesis.Title},
		{Label: "Tanggal & Waktu", Value: formatTime(defense.ScheduledAt)},
		{Label: "Ruangan", Value: deref(defense.Room)},
		{Label: "Dosen Penguji", Value: joinUserNames(defense.Examiners)},
	}
	data.Notes = deref(defense.RevisionNotes)
	data.CTA = &CTA{Label: "Lihat Detail", URL: s.frontendURL + "/defenses/" + defense.ID.String()}
	return s.send(to, subject, "defense_scheduled.html", "defense_scheduled", data)
}

func (s *ResendEmailService) SendDefenseFinalized(ctx context.Context, to string, defense *entity.ThesisDefense) error {
	passed := defense.Status == "passed"
	subject := "[SIMTAS] Hasil Sidang Skripsi"
	if passed {
		subject = "[SIMTAS] Selamat! Anda Lulus Sidang Skripsi"
	}
	data := s.baseData("Hasil Sidang Skripsi")
	data.Greeting = "Halo " + defense.Thesis.Student.FullName + ","
	data.Status = "LULUS"
	data.StatusGood = true
	if !passed {
		data.Status = "TIDAK LULUS"
		data.StatusGood = false
	}
	data.Message = "Hasil sidang skripsi Anda telah dirilis:"
	data.Details = []Detail{{Label: "Nilai Akhir", Value: formatScore(defense.FinalScore)}}
	data.Notes = deref(defense.RevisionNotes)
	data.CTA = &CTA{Label: "Lihat Hasil", URL: s.frontendURL + "/defenses/" + defense.ID.String()}
	return s.send([]string{to}, subject, "defense_result.html", "defense_finalized", data)
}

func (s *ResendEmailService) SendGraduated(ctx context.Context, to string, thesis *entity.Thesis) error {
	subject := "[SIMTAS] Selamat! Skripsi Anda Dinyatakan Lulus"
	data := s.baseData("Selamat! Skripsi Anda Dinyatakan Lulus")
	data.Greeting = "Halo " + thesis.Student.FullName + ","
	data.Message = "Selamat! Skripsi Anda telah dinyatakan LULUS pada yudisium."
	data.Details = []Detail{
		{Label: "Judul Skripsi", Value: thesis.Title},
		{Label: "Tanggal Yudisium", Value: formatTime(thesis.GraduatedAt)},
	}
	data.Notes = "Langkah selanjutnya: unggah skripsi final Anda untuk diarsipkan."
	data.CTA = &CTA{Label: "Unggah Skripsi Final", URL: s.frontendURL + "/mahasiswa/theses/" + thesis.ID.String() + "/archive"}
	return s.send([]string{to}, subject, "graduation.html", "graduated", data)
}

// ── Archive notifications (Job 10) ──────────────────────────────────────

func (s *ResendEmailService) SendArchiveCreated(ctx context.Context, to string, archive *entity.ThesisArchive) error {
	subject := "[SIMTAS] Arsip Skripsi Anda Tersedia"
	data := s.baseData("Arsip Skripsi Anda Tersedia")
	data.Greeting = "Halo,"
	data.Message = "Skripsi Anda telah diarsipkan di perpustakaan digital SIMTAS."
	data.Details = []Detail{
		{Label: "Judul", Value: archive.Thesis.Title},
		{Label: "Tahun Kelulusan", Value: strconv.Itoa(archive.GraduationYear)},
	}
	data.CTA = &CTA{Label: "Lihat Arsip", URL: s.frontendURL + "/archives/" + archive.ID.String()}
	return s.send([]string{to}, subject, "event.html", "archive_created", data)
}

// ── Internal helpers ────────────────────────────────────────────────────

func (s *ResendEmailService) baseData(title string) TemplateData {
	return TemplateData{
		AppName:     "SIMTAS FILKOM",
		FrontendURL: s.frontendURL,
		Year:        time.Now().Year(),
		Title:       title,
	}
}

// send renders the template synchronously (so template errors surface), then
// delivers asynchronously so the HTTP response never waits for the email.
// The sender uses its own background context during delivery, so no request
// context is threaded through here.
func (s *ResendEmailService) send(to []string, subject, templateName, eventType string, data TemplateData) error {
	if len(to) == 0 {
		return nil
	}
	htmlBody, err := s.render(templateName, data)
	if err != nil {
		return fmt.Errorf("render email template %s: %w", templateName, err)
	}
	go s.deliver(to, subject, htmlBody, eventType)
	return nil
}

func (s *ResendEmailService) render(name string, data TemplateData) (string, error) {
	tmpl, ok := s.templates[name]
	if !ok {
		return "", fmt.Errorf("unknown email template %q", name)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// deliver performs the actual send (or console log in dev mode) and records
// the outcome in email_logs. It uses a fresh background context because the
// request context may be cancelled by the time the goroutine runs. A panic
// inside the client is recovered so a send failure never crashes the process.
func (s *ResendEmailService) deliver(to []string, subject, htmlBody, eventType string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[email-resend][%s] panic while sending to %v: %v", eventType, to, r)
		}
	}()

	sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if s.devMode {
		log.Printf("[email-dev][%s] to=%v subject=%q (template rendered, not sent)", eventType, to, subject)
		for _, recipient := range to {
			s.record(sendCtx, recipient, eventType, subject, "sent", "")
		}
		return
	}
	if s.client == nil {
		log.Printf("[email-resend][%s] client is nil, skipping send to %v", eventType, to)
		return
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      to,
		Subject: subject,
		Html:    htmlBody,
	}
	if _, err := s.client.Emails.Send(params); err != nil {
		log.Printf("[email-resend][%s] send failed to %v: %v", eventType, to, err)
		for _, recipient := range to {
			s.record(sendCtx, recipient, eventType, subject, "failed", err.Error())
		}
		return
	}
	for _, recipient := range to {
		s.record(sendCtx, recipient, eventType, subject, "sent", "")
	}
}

// record writes an email_logs entry when a DB handle is available.
func (s *ResendEmailService) record(ctx context.Context, to, eventType, subject, status, errMsg string) {
	if s.db == nil {
		return
	}
	entry := &entity.EmailLog{
		RecipientEmail: to,
		EventType:      eventType,
		Subject:        &subject,
		Status:         status,
		Provider:       "resend",
	}
	if s.devMode {
		entry.Provider = "dev"
	}
	if errMsg != "" {
		entry.ErrorMessage = &errMsg
	}
	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		// Never fail the request because logging failed.
		log.Printf("[email-resend] failed to record email_logs entry: %v", err)
	}
}

// SendTestEmail sends a diagnostic test email (used by the internal dev endpoint).
func (s *ResendEmailService) SendTestEmail(ctx context.Context, to string) error {
	subject := "[SIMTAS] Email Uji Coba"
	data := s.baseData("Email Uji Coba")
	data.Message = "Jika Anda menerima email ini, konfigurasi email SIMTAS berfungsi dengan benar."
	env := "production"
	if s.devMode {
		env = "development (dev mode)"
	}
	data.Details = []Detail{
		{Label: "Waktu Kirim", Value: time.Now().Format("02 Jan 2006 15:04:05")},
		{Label: "Lingkungan", Value: env},
	}
	data.CTA = &CTA{Label: "Buka SIMTAS", URL: s.frontendURL}
	return s.send([]string{to}, subject, "event.html", "test_email", data)
}

// ── Small helpers ────────────────────────────────────────────────────────

func formatTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("02 Jan 2006 15:04")
}

func formatScore(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *v)
}

func joinUserNames(users []entity.User) string {
	names := make([]string, 0, len(users))
	for _, u := range users {
		if u.FullName != "" {
			names = append(names, u.FullName)
		}
	}
	if len(names) == 0 {
		return "-"
	}
	return strings.Join(names, ", ")
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}

// deref safely dereferences a string pointer for display/logging.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// loadTemplates parses base.html together with every event template into a
// cached map so sends do not re-parse HTML on each call.
func loadTemplates() map[string]*template.Template {
	names := []string{
		"event.html",
		"thesis_submitted.html",
		"thesis_reviewed.html",
		"supervisor_assigned.html",
		"seminar_scheduled.html",
		"seminar_result.html",
		"title_change_requested.html",
		"title_change_approved.html",
		"title_change_rejected.html",
		"title_change_cancelled.html",

		"defense_scheduled.html",
		"defense_result.html",
		"graduation.html",
	}
	out := make(map[string]*template.Template, len(names))
	for _, name := range names {
		tmpl, err := template.ParseFS(templateFS, "templates/base.html", "templates/"+name)
		if err != nil {
			// A broken template must not take down the server; renders will fail
			// loudly per send until the template is fixed.
			log.Printf("[email-resend] failed to parse template %s: %v", name, err)
			continue
		}
		out[name] = tmpl
	}
	return out
}
