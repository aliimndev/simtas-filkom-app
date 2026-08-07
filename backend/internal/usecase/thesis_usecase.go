package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/service"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/notification"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/statemachine"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

// Role name constants (kept in sync with middleware.Role* and seed data).
const (
	ThesisRoleAdminFakultas   = "admin_fakultas"
	ThesisRoleKaprodi         = "kaprodi"
	ThesisRoleMahasiswa       = "mahasiswa"
	ThesisRoleDosenPembimbing = "dosen_pembimbing"
	ThesisRoleDosenPenguji    = "dosen_penguji"
)

var (
	ErrThesisNotFound         = errors.New("thesis tidak ditemukan")
	ErrActiveThesisExists     = errors.New("mahasiswa sudah memiliki thesis aktif")
	ErrNoActiveAcademicYear   = errors.New("tidak ada tahun akademik aktif")
	ErrTitleTooShort          = errors.New("judul minimal 10 kata")
	ErrTitleTooLong           = errors.New("judul maksimal 500 karakter")
	ErrAbstractTooShort       = errors.New("abstrak minimal 100 kata")
	ErrInvalidThesisType      = errors.New("thesis_type harus skripsi atau tugas_akhir")
	ErrDraftRequired          = errors.New("draft proposal wajib diunggah")
	ErrInvalidDecision        = errors.New("decision harus approved atau rejected")
	ErrInvalidStateTransition = errors.New("transisi status tidak valid")
	ErrInvalidSupervisorCount = errors.New("jumlah supervisor minimal 1 dan maksimal 2")
	ErrSupervisorNotEligible  = errors.New("supervisor harus dosen pembimbing yang aktif")
	ErrThesisAlreadyCancelled = errors.New("thesis sudah dibatalkan")
	ErrThesisCannotCancel     = errors.New("thesis tidak dapat dibatalkan karena sudah graduated")
	ErrForbidden              = errors.New("akses ditolak")
)

// CreateThesisRequest is the multipart payload for POST /theses (Job 05).
// Form fields: title, abstract, field_of_study, thesis_type + the proposal
// PDF as multipart field "file".
type CreateThesisRequest struct {
	Title        string
	Abstract     string
	FieldOfStudy string
	ThesisType   string

	// Draft proposal (proposal document on the thesis, created atomically).
	DraftFile   multipart.File
	DraftHeader *multipart.FileHeader
}

// ReviewThesisRequest is the payload for PUT /theses/:id/review.
type ReviewThesisRequest struct {
	Decision string `json:"decision" binding:"required"` // approved | rejected
	Notes    string `json:"notes"`
}

// AssignSupervisorRequest is the payload for PUT /theses/:id/assign-supervisor.
type AssignSupervisorRequest struct {
	SupervisorIDs []uuid.UUID `json:"supervisor_ids" binding:"required"`
}

// CancelThesisRequest is the payload for PATCH /theses/:id/cancel.
type CancelThesisRequest struct {
	Reason string `json:"reason"`
}

// ThesisDetail is the API response shape for a thesis (matches Job 05 spec).
type ThesisDetail struct {
	ID           uuid.UUID          `json:"id"`
	Title        string             `json:"title"`
	Abstract     *string            `json:"abstract,omitempty"`
	FieldOfStudy *string            `json:"field_of_study,omitempty"`
	ThesisType   string             `json:"thesis_type"`
	Status       string             `json:"status"`
	KaprodiNotes *string            `json:"kaprodi_notes,omitempty"`
	Student      ThesisStudent      `json:"student"`
	Supervisors  []ThesisSupervisor `json:"supervisors,omitempty"`
	AcademicYear ThesisAcademicYear `json:"academic_year"`
	SubmittedAt  time.Time          `json:"submitted_at"`
	ApprovedAt   *time.Time         `json:"approved_at,omitempty"`
}

type ThesisStudent struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Nim      *string   `json:"nim,omitempty"`
}

type ThesisSupervisor struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
}

type ThesisAcademicYear struct {
	Name     string `json:"name"`
	Semester string `json:"semester"`
}

// LecturerLoad describes an active dosen_pembimbing with their current load
// (used for load-balancing hints when assigning supervisors).
type LecturerLoad struct {
	ID                     uuid.UUID `json:"id"`
	FullName               string    `json:"full_name"`
	Nidn                   *string   `json:"nidn,omitempty"`
	ActiveSupervisionCount int64     `json:"active_supervision_count"`
}

// ThesisUseCase contains business logic for the thesis submission workflow.
type ThesisUseCase struct {
	thesisRepo   domainRepo.ThesisRepository
	userRepo     domainRepo.UserRepository
	acadRepo     domainRepo.AcademicYearRepository
	documentRepo domainRepo.DocumentRepository
	storage      service.StorageService
	emailSvc     email.EmailService
	auditSvc     *audit.AuditService
	notifSvc     *notification.NotificationService
}

func NewThesisUseCase(
	thesisRepo domainRepo.ThesisRepository,
	userRepo domainRepo.UserRepository,
	acadRepo domainRepo.AcademicYearRepository,
	documentRepo domainRepo.DocumentRepository,
	storage service.StorageService,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
	notifSvc *notification.NotificationService,
) *ThesisUseCase {
	return &ThesisUseCase{
		thesisRepo:   thesisRepo,
		userRepo:     userRepo,
		acadRepo:     acadRepo,
		documentRepo: documentRepo,
		storage:      storage,
		emailSvc:     emailSvc,
		auditSvc:     auditSvc,
		notifSvc:     notifSvc,
	}
}

// Submit creates a new thesis submission for the student.
func (uc *ThesisUseCase) Submit(ctx context.Context, req CreateThesisRequest, studentID uuid.UUID, actor Actor) (*ThesisDetail, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Abstract = strings.TrimSpace(req.Abstract)

	if len(strings.Fields(req.Title)) < 10 {
		return nil, ErrTitleTooShort
	}
	if len(req.Title) > 500 {
		return nil, ErrTitleTooLong
	}
	if len(strings.Fields(req.Abstract)) < 100 {
		return nil, ErrAbstractTooShort
	}
	if req.ThesisType != "skripsi" && req.ThesisType != "tugas_akhir" {
		return nil, ErrInvalidThesisType
	}
	if req.DraftFile == nil || req.DraftHeader == nil {
		return nil, ErrDraftRequired
	}
	// PDF-only, ≤ 10 MB (same rules as the document upload module).
	if err := utils.ValidatePDF(req.DraftFile, req.DraftHeader, MaxDocumentSizeBytes); err != nil {
		return nil, err
	}

	// A student may only have one active thesis (not cancelled/graduated).
	active, err := uc.thesisRepo.FindActiveByStudentID(ctx, studentID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if active != nil {
		return nil, ErrActiveThesisExists
	}

	// Attach to the currently active academic year.
	year, err := uc.acadRepo.FindActive(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoActiveAcademicYear
		}
		return nil, err
	}

	var abstract, fieldOfStudy *string
	abstract = &req.Abstract
	if req.FieldOfStudy != "" {
		fieldOfStudy = &req.FieldOfStudy
	}

	thesis := &entity.Thesis{
		ID:             uuid.New(),
		StudentID:      studentID,
		AcademicYearID: year.ID,
		Title:          req.Title,
		Abstract:       abstract,
		FieldOfStudy:   fieldOfStudy,
		ThesisType:     req.ThesisType,
		Status:         "submitted",
		SubmittedAt:    time.Now(),
	}

	// Upload the draft first (no DB writes yet) so a storage failure leaves
	// nothing behind; then create the thesis row and its proposal document.
	baseName := sanitizeFileName(req.DraftHeader.Filename)
	storagePath := fmt.Sprintf("theses/%s/%s/v1_%s", thesis.ID, entity.DocTypeProposal, baseName)
	fileURL, err := uc.storage.Upload(ctx, storagePath, req.DraftFile, req.DraftHeader.Size, req.DraftHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	if err := uc.thesisRepo.Create(ctx, thesis); err != nil {
		_ = uc.storage.Delete(ctx, fileURL)
		return nil, err
	}

	doc := &entity.Document{
		ThesisID:     thesis.ID,
		UploadedBy:   studentID,
		DocumentType: entity.DocTypeProposal,
		Version:      1,
		FileName:     baseName,
		FileURL:      fileURL,
		FileSize:     &req.DraftHeader.Size,
		Status:       entity.DocStatusPendingReview,
	}
	if err := uc.documentRepo.Create(ctx, doc); err != nil {
		_ = uc.storage.Delete(ctx, fileURL)
		return nil, err
	}

	// Re-fetch so associations (student, academic year) are populated in the response.
	thesis, err = uc.thesisRepo.FindByID(ctx, thesis.ID)
	if err != nil {
		return nil, err
	}

	// Notify all kaprodi (async, non-fatal).
	go func() {
		kaprodi, err := uc.userRepo.FindByRole(context.Background(), ThesisRoleKaprodi)
		if err != nil {
			return
		}
		emails := make([]string, 0, len(kaprodi))
		ids := make([]uuid.UUID, 0, len(kaprodi))
		for _, k := range kaprodi {
			emails = append(emails, k.Email)
			ids = append(ids, k.ID)
		}
		if len(emails) > 0 {
			_ = uc.emailSvc.SendThesisSubmitted(context.Background(), emails, thesis)
		}
		if len(ids) > 0 {
			uc.notifSvc.Notify(notification.Params{
				UserIDs: ids,
				Title:   "Pengajuan Judul Skripsi Baru",
				Message: thesis.Student.FullName + " mengajukan judul skripsi baru.",
				Type:    "thesis",
				Link:    notification.Path("/theses/%s", thesis.ID),
			})
		}
	}()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionThesisSubmitted,
		EntityType: "thesis",
		EntityID:   &thesis.ID,
		NewValue:   map[string]interface{}{"title": thesis.Title, "status": thesis.Status},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	return toThesisDetail(thesis), nil
}

// List returns theses scoped by the caller's role and matching the filters.
func (uc *ThesisUseCase) List(ctx context.Context, filter domainRepo.ThesisFilter, userID uuid.UUID, role string) ([]*ThesisDetail, int64, error) {
	switch role {
	case ThesisRoleMahasiswa:
		filter.StudentID = userID
	case ThesisRoleDosenPembimbing:
		filter.SupervisorID = userID
	case ThesisRoleDosenPenguji:
		filter.ExaminerID = userID
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		// all theses
	}

	theses, total, err := uc.thesisRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	details := make([]*ThesisDetail, 0, len(theses))
	for _, t := range theses {
		details = append(details, toThesisDetail(t))
	}
	return details, total, nil
}

// GetByID returns one thesis, enforcing role-based access.
func (uc *ThesisUseCase) GetByID(ctx context.Context, id, userID uuid.UUID, role string) (*ThesisDetail, error) {
	thesis, err := uc.thesisRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}

	ok, err := uc.canAccessThesis(ctx, thesis, userID, role)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return toThesisDetail(thesis), nil
}

// Review approves or rejects a submitted thesis (Kaprodi only).
func (uc *ThesisUseCase) Review(ctx context.Context, id uuid.UUID, req ReviewThesisRequest, actor Actor) (*ThesisDetail, error) {
	if req.Decision != "approved" && req.Decision != "rejected" {
		return nil, ErrInvalidDecision
	}

	thesis, err := uc.thesisRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}
	if !statemachine.CanTransition(thesis.Status, req.Decision) {
		return nil, ErrInvalidStateTransition
	}

	oldStatus := thesis.Status
	thesis.Status = req.Decision
	if req.Notes != "" {
		thesis.KaprodiNotes = &req.Notes
	}
	if req.Decision == "approved" {
		now := time.Now()
		thesis.ApprovedAt = &now
	}

	if err := uc.thesisRepo.Update(ctx, thesis); err != nil {
		return nil, err
	}

	// Notify the student (async, non-fatal).
	studentEmail := thesis.Student.Email
	go func() {
		if req.Decision == "approved" {
			_ = uc.emailSvc.SendThesisApproved(context.Background(), studentEmail, thesis)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: []uuid.UUID{thesis.Student.ID},
				Title:   "Judul Skripsi Disetujui",
				Message: "Selamat, judul skripsi Anda telah disetujui oleh Kaprodi.",
				Type:    "thesis",
				Link:    notification.Path("/theses/%s", thesis.ID),
			})
		} else {
			_ = uc.emailSvc.SendThesisRejected(context.Background(), studentEmail, thesis, req.Notes)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: []uuid.UUID{thesis.Student.ID},
				Title:   "Judul Skripsi Perlu Revisi",
				Message: "Judul skripsi Anda belum dapat disetujui. Periksa catatan Kaprodi.",
				Type:    "thesis",
				Link:    notification.Path("/theses/%s", thesis.ID),
			})
		}
	}()

	action := audit.ActionThesisApproved
	if req.Decision == "rejected" {
		action = audit.ActionThesisRejected
	}
	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     action,
		EntityType: "thesis",
		EntityID:   &thesis.ID,
		OldValue:   map[string]interface{}{"status": oldStatus},
		NewValue:   map[string]interface{}{"status": thesis.Status, "notes": req.Notes},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	return toThesisDetail(thesis), nil
}

// AssignSupervisor assigns 1–2 dosen_pembimbing and moves the thesis to in_progress.
func (uc *ThesisUseCase) AssignSupervisor(ctx context.Context, id uuid.UUID, req AssignSupervisorRequest, actor Actor) (*ThesisDetail, error) {
	// Dedupe while preserving order so duplicates don't trip the DB unique constraint.
	seen := make(map[uuid.UUID]bool, len(req.SupervisorIDs))
	unique := make([]uuid.UUID, 0, len(req.SupervisorIDs))
	for _, sid := range req.SupervisorIDs {
		if !seen[sid] {
			seen[sid] = true
			unique = append(unique, sid)
		}
	}
	req.SupervisorIDs = unique

	if len(req.SupervisorIDs) < 1 || len(req.SupervisorIDs) > 2 {
		return nil, ErrInvalidSupervisorCount
	}

	thesis, err := uc.thesisRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}
	if !statemachine.CanTransition(thesis.Status, "in_progress") {
		return nil, ErrInvalidStateTransition
	}

	// Every requested supervisor must be an active dosen_pembimbing.
	lecturers, err := uc.userRepo.FindByRole(ctx, ThesisRoleDosenPembimbing)
	if err != nil {
		return nil, err
	}
	validIDs := make(map[uuid.UUID]*entity.User, len(lecturers))
	for _, l := range lecturers {
		validIDs[l.ID] = l
	}

	supervisorEmails := make([]string, 0, len(req.SupervisorIDs))
	for _, sid := range req.SupervisorIDs {
		if _, ok := validIDs[sid]; !ok {
			return nil, ErrSupervisorNotEligible
		}
		supervisorEmails = append(supervisorEmails, validIDs[sid].Email)
	}

	// Assign all supervisors and move the thesis to in_progress atomically so a
	// failure cannot leave the thesis with a partial supervisor set.
	if err := uc.thesisRepo.AssignSupervisors(ctx, id, req.SupervisorIDs, actor.UserID); err != nil {
		return nil, err
	}
	// Re-fetch so the supervisors association reflects the new assignment.
	thesis, err = uc.thesisRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Notify student + newly assigned supervisors (async, non-fatal).
	studentEmail := thesis.Student.Email
	go func() {
		_ = uc.emailSvc.SendSupervisorAssigned(context.Background(), studentEmail, supervisorEmails, thesis)
		uc.notifSvc.Notify(notification.Params{
			UserIDs: append(userIDs(thesis.Supervisors), thesis.Student.ID),
			Title:   "Dosen Pembimbing Ditetapkan",
			Message: "Dosen pembimbing telah ditetapkan untuk skripsi \"" + thesis.Title + "\".",
			Type:    "thesis",
			Link:    notification.Path("/theses/%s", thesis.ID),
		})
	}()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionSupervisorAssigned,
		EntityType: "thesis",
		EntityID:   &thesis.ID,
		NewValue: map[string]interface{}{
			"status":         "in_progress",
			"supervisor_ids": req.SupervisorIDs,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})

	return toThesisDetail(thesis), nil
}

// Cancel cancels a thesis (Admin + Kaprodi only).
func (uc *ThesisUseCase) Cancel(ctx context.Context, id uuid.UUID, req CancelThesisRequest, actor Actor) error {
	thesis, err := uc.thesisRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrThesisNotFound
		}
		return err
	}
	if thesis.Status == "cancelled" {
		return ErrThesisAlreadyCancelled
	}
	if thesis.Status == "graduated" {
		return ErrThesisCannotCancel
	}

	if err := uc.thesisRepo.UpdateStatus(ctx, id, "cancelled", req.Reason); err != nil {
		return err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionThesisCancelled,
		EntityType: "thesis",
		EntityID:   &thesis.ID,
		OldValue:   map[string]interface{}{"status": thesis.Status},
		NewValue:   map[string]interface{}{"status": "cancelled", "reason": req.Reason},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return nil
}

// ListLecturers returns active dosen_pembimbing sorted by ascending load
// (lowest supervision count first) for load-balancing hints.
func (uc *ThesisUseCase) ListLecturers(ctx context.Context) ([]LecturerLoad, error) {
	lecturers, err := uc.userRepo.FindByRole(ctx, ThesisRoleDosenPembimbing)
	if err != nil {
		return nil, err
	}

	result := make([]LecturerLoad, 0, len(lecturers))
	for _, l := range lecturers {
		count, err := uc.thesisRepo.CountActiveSupervisions(ctx, l.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, LecturerLoad{
			ID:                     l.ID,
			FullName:               l.FullName,
			Nidn:                   l.NimNidn,
			ActiveSupervisionCount: count,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].ActiveSupervisionCount < result[j].ActiveSupervisionCount
	})
	return result, nil
}

// canAccessThesis enforces per-role access to a single thesis.
func (uc *ThesisUseCase) canAccessThesis(ctx context.Context, thesis *entity.Thesis, userID uuid.UUID, role string) (bool, error) {
	switch role {
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		return true, nil
	case ThesisRoleMahasiswa:
		return thesis.StudentID == userID, nil
	case ThesisRoleDosenPembimbing:
		for _, s := range thesis.Supervisors {
			if s.ID == userID {
				return true, nil
			}
		}
		return false, nil
	case ThesisRoleDosenPenguji:
		return uc.thesisRepo.IsExaminer(ctx, thesis.ID, userID)
	default:
		return false, nil
	}
}

// toThesisDetail maps a persisted thesis into the API response shape.
func toThesisDetail(t *entity.Thesis) *ThesisDetail {
	detail := &ThesisDetail{
		ID:           t.ID,
		Title:        t.Title,
		Abstract:     t.Abstract,
		FieldOfStudy: t.FieldOfStudy,
		ThesisType:   t.ThesisType,
		Status:       t.Status,
		KaprodiNotes: t.KaprodiNotes,
		Student: ThesisStudent{
			ID:       t.Student.ID,
			FullName: t.Student.FullName,
			Nim:      t.Student.NimNidn,
		},
		AcademicYear: ThesisAcademicYear{
			Name:     t.AcademicYear.Name,
			Semester: t.AcademicYear.Semester,
		},
		SubmittedAt: t.SubmittedAt,
		ApprovedAt:  t.ApprovedAt,
	}
	for _, s := range t.Supervisors {
		detail.Supervisors = append(detail.Supervisors, ThesisSupervisor{ID: s.ID, FullName: s.FullName})
	}
	return detail
}
