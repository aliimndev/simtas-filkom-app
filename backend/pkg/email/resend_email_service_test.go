package email

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/google/uuid"
)

// sampleData returns TemplateData exercising every optional field so template
// rendering is fully covered.
func sampleData() TemplateData {
	return TemplateData{
		AppName:     "SIMTAS FILKOM",
		FrontendURL: "http://localhost:3000",
		Year:        time.Now().Year(),
		Title:       "Judul Email",
		Greeting:    "Halo Mahasiswa Satu,",
		Message:     "Ini adalah pesan uji coba.",
		Status:      "LULUS",
		StatusGood:  true,
		Details:     []Detail{{Label: "Nama", Value: "Mahasiswa Satu"}, {Label: "NIM", Value: "231401001"}},
		Notes:       "Catatan tambahan.",
		CTA:         &CTA{Label: "Lihat Detail", URL: "http://localhost:3000/detail"},
	}
}

// TestRenderAllTemplates ensures every event template parses and renders with
// full sample data without error (done criteria: all templates render).
func TestRenderAllTemplates(t *testing.T) {
	svc := NewResendEmailService("", "noreply@test.local", "SIMTAS", "http://localhost:3000", nil, true)

	templates := []string{
		"event.html",
		"thesis_submitted.html",
		"thesis_reviewed.html",
		"supervisor_assigned.html",
		"seminar_scheduled.html",
		"seminar_result.html",
		"defense_scheduled.html",
		"defense_result.html",
		"graduation.html",
	}

	for _, name := range templates {
		t.Run(name, func(t *testing.T) {
			html, err := svc.render(name, sampleData())
			if err != nil {
				t.Fatalf("render %s: %v", name, err)
			}
			if !strings.Contains(html, "SIMTAS FILKOM") {
				t.Errorf("render %s: missing brand header", name)
			}
			if !strings.Contains(html, "Judul Email") {
				t.Errorf("render %s: missing title", name)
			}
		})
	}
}

// TestRenderEscapesHTML ensures template rendering escapes user content (XSS).
func TestRenderEscapesHTML(t *testing.T) {
	svc := NewResendEmailService("", "noreply@test.local", "SIMTAS", "http://localhost:3000", nil, true)
	data := sampleData()
	data.Message = "<script>alert('xss')</script>"

	html, err := svc.render("event.html", data)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(html, "<script>") {
		t.Errorf("render: user content was not escaped: %s", html)
	}
}

// TestDevModeDoesNotSend verifies that dev mode renders + records without
// calling the Resend API (client stays nil, no panic).
func TestDevModeDoesNotSend(t *testing.T) {
	svc := NewResendEmailService("", "noreply@test.local", "SIMTAS", "http://localhost:3000", nil, true)
	if svc.client != nil {
		t.Fatal("dev mode should not create a Resend client")
	}
	if err := svc.SendWelcomeEmail(context.Background(), "mahasiswa@test.local", "Mahasiswa Satu", "TempPass123"); err != nil {
		t.Fatalf("SendWelcomeEmail: %v", err)
	}
	if err := svc.SendTestEmail(context.Background(), "admin@test.local"); err != nil {
		t.Fatalf("SendTestEmail: %v", err)
	}
}

// TestSendSkipsEmptyRecipients ensures no work happens when the recipient
// list is empty.
func TestSendSkipsEmptyRecipients(t *testing.T) {
	svc := NewResendEmailService("", "noreply@test.local", "SIMTAS", "http://localhost:3000", nil, true)
	if err := svc.SendThesisSubmitted(context.Background(), nil, &entity.Thesis{Title: "X"}); err != nil {
		t.Fatalf("send with empty recipients: %v", err)
	}
}

// TestSendSupervisorAssigned verifies both the student and supervisor emails
// are rendered (dev mode, nil DB).
func TestSendSupervisorAssigned(t *testing.T) {
	svc := NewResendEmailService("", "noreply@test.local", "SIMTAS", "http://localhost:3000", nil, true)
	thesis := &entity.Thesis{
		ID:    uuid.New(),
		Title: "Sistem Rekomendasi Skripsi",
		Student: entity.User{
			FullName: "Mahasiswa Satu",
			Email:    "mahasiswa@test.local",
		},
		Supervisors: []entity.User{
			{FullName: "Dosen Satu", Email: "dosen1@test.local"},
			{FullName: "Dosen Dua", Email: "dosen2@test.local"},
		},
	}
	if err := svc.SendSupervisorAssigned(context.Background(), "mahasiswa@test.local", []string{"dosen1@test.local", "dosen2@test.local"}, thesis); err != nil {
		t.Fatalf("SendSupervisorAssigned: %v", err)
	}
}

// TestFullInterfaceSurface ensures every interface method is reachable in dev
// mode without a DB and without panicking (nil client).
func TestFullInterfaceSurface(t *testing.T) {
	svc := NewResendEmailService("", "noreply@test.local", "SIMTAS", "http://localhost:3000", nil, true)
	ctx := context.Background()
	thesis := &entity.Thesis{
		ID:      uuid.New(),
		Title:   "Judul Skripsi",
		Student: entity.User{FullName: "Mahasiswa Satu", Email: "mhs@test.local"},
	}
	consultation := &entity.ConsultationLog{ID: uuid.New(), ThesisID: thesis.ID, ConsultationDate: time.Now(), TopicsDiscussed: "Bab 1"}
	doc := &entity.Document{ID: uuid.New(), ThesisID: thesis.ID, DocumentType: "proposal", Version: 1, FileName: "bab1.pdf"}
	seminar := &entity.Seminar{ID: uuid.New(), ThesisID: thesis.ID, Thesis: *thesis, Status: "passed"}
	defense := &entity.ThesisDefense{ID: uuid.New(), ThesisID: thesis.ID, Thesis: *thesis, Status: "passed"}
	archive := &entity.ThesisArchive{ID: uuid.New(), ThesisID: thesis.ID, Thesis: *thesis, GraduationYear: 2026}

	calls := []func() error{
		func() error { return svc.SendWelcomeEmail(ctx, "mhs@test.local", "Mahasiswa Satu", "pw123") },
		func() error { return svc.SendPasswordReset(ctx, "mhs@test.local", "Mahasiswa Satu", "pw123") },
		func() error { return svc.SendThesisSubmitted(ctx, []string{"kaprodi@test.local"}, thesis) },
		func() error { return svc.SendThesisApproved(ctx, "mhs@test.local", thesis) },
		func() error { return svc.SendThesisRejected(ctx, "mhs@test.local", thesis, "revisi abstrak") },
		func() error {
			return svc.SendSupervisorAssigned(ctx, "mhs@test.local", []string{"dosen@test.local"}, thesis)
		},
		func() error { return svc.SendConsultationCreated(ctx, []string{"dosen@test.local"}, consultation) },
		func() error { return svc.SendConsultationApproved(ctx, "mhs@test.local", consultation) },
		func() error { return svc.SendDocumentUploaded(ctx, []string{"dosen@test.local"}, doc) },
		func() error { return svc.SendDocumentReviewed(ctx, "mhs@test.local", doc, "approved") },
		func() error { return svc.SendSeminarSubmitted(ctx, []string{"kaprodi@test.local"}, seminar) },
		func() error { return svc.SendSeminarScheduled(ctx, []string{"mhs@test.local"}, seminar) },
		func() error { return svc.SendSeminarFinalized(ctx, "mhs@test.local", seminar) },
		func() error { return svc.SendDefenseSubmitted(ctx, []string{"kaprodi@test.local"}, defense) },
		func() error { return svc.SendDefenseScheduled(ctx, []string{"mhs@test.local"}, defense) },
		func() error { return svc.SendDefenseFinalized(ctx, "mhs@test.local", defense) },
		func() error { return svc.SendGraduated(ctx, "mhs@test.local", thesis) },
		func() error { return svc.SendArchiveCreated(ctx, "mhs@test.local", archive) },
		func() error { return svc.SendTestEmail(ctx, "admin@test.local") },
	}

	for i, call := range calls {
		if err := call(); err != nil {
			t.Errorf("interface method #%d: %v", i, err)
		}
	}
}
