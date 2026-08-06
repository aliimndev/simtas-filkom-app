package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strconv"
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
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

// Document review decisions (Job 07).
const (
	DocDecisionApproved        = "approved"
	DocDecisionRevisionReq     = "revision_required"
	DocumentDownloadExpirySecs = 900 // 15 minutes
)

// MaxDocumentSizeBytes caps PDF uploads at 10 MB (Job 07).
const MaxDocumentSizeBytes = 10 << 20

var (
	ErrDocumentNotFound      = errors.New("dokumen tidak ditemukan")
	ErrInvalidDocumentType   = errors.New("document_type tidak valid")
	ErrInvalidChapterNumber  = errors.New("chapter_number harus 1-5 untuk draft_chapter")
	ErrDocumentThesisNotLive = errors.New("thesis harus berstatus in_progress atau lebih lanjut")
	ErrDocumentFileRequired  = errors.New("file wajib diunggah")
	ErrInvalidReviewDecision = errors.New("decision harus approved atau revision_required")
	ErrDocumentNotPending    = errors.New("dokumen tidak dalam status menunggu review")
	ErrNotDocumentReviewer   = errors.New("hanya dosen pembimbing thesis ini yang dapat mereview")
	ErrChapterNumberRequired = errors.New("chapter_number wajib diisi untuk draft_chapter")
)

// UploadDocumentRequest describes the validated multipart form fields (Job 07).
// Extra form fields (e.g. a client-supplied "notes") are ignored: the documents
// table has no uploader-notes column — reviewers record notes on review.
type UploadDocumentRequest struct {
	DocumentType  string
	ChapterNumber *int
}

// DocumentDetail is the API response shape for a document (Job 07).
type DocumentDetail struct {
	ID            uuid.UUID  `json:"id"`
	DocumentType  string     `json:"document_type"`
	ChapterNumber *int       `json:"chapter_number,omitempty"`
	Version       int        `json:"version"`
	FileName      string     `json:"file_name"`
	FileSize      *int64     `json:"file_size,omitempty"`
	Status        string     `json:"status"`
	Reviewer      *UserBrief `json:"reviewer,omitempty"`
	ReviewerNotes *string    `json:"reviewer_notes,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// DownloadResult wraps a presigned download URL (Job 07).
type DownloadResult struct {
	DownloadURL string `json:"download_url"`
	ExpiresIn   int    `json:"expires_in"`
}

// DocumentUseCase contains business logic for the document management module (Job 07).
type DocumentUseCase struct {
	documentRepo domainRepo.DocumentRepository
	thesisRepo   domainRepo.ThesisRepository
	storage      service.StorageService
	access       *ThesisAccess
	emailSvc     email.EmailService
	auditSvc     *audit.AuditService
	notifSvc     *notification.NotificationService
}

func NewDocumentUseCase(
	documentRepo domainRepo.DocumentRepository,
	thesisRepo domainRepo.ThesisRepository,
	storage service.StorageService,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
	notifSvc *notification.NotificationService,
) *DocumentUseCase {
	return &DocumentUseCase{
		documentRepo: documentRepo,
		thesisRepo:   thesisRepo,
		storage:      storage,
		access:       NewThesisAccess(thesisRepo),
		emailSvc:     emailSvc,
		auditSvc:     auditSvc,
		notifSvc:     notifSvc,
	}
}

// Upload validates the file, stores it, and creates a pending_review document.
// Only the owning mahasiswa may upload.
func (uc *DocumentUseCase) Upload(
	ctx context.Context,
	thesisID uuid.UUID,
	file multipart.File,
	header *multipart.FileHeader,
	req UploadDocumentRequest,
	actor Actor,
) (*DocumentDetail, error) {
	req.DocumentType = strings.TrimSpace(req.DocumentType)
	if !entity.ValidDocumentType(req.DocumentType) {
		return nil, ErrInvalidDocumentType
	}
	if req.DocumentType == entity.DocTypeDraftChapter && req.ChapterNumber == nil {
		return nil, ErrChapterNumberRequired
	}
	if req.ChapterNumber != nil && (*req.ChapterNumber < 1 || *req.ChapterNumber > 5) {
		return nil, ErrInvalidChapterNumber
	}

	// Validate the file (PDF only, ≤ 10 MB).
	if file == nil || header == nil {
		return nil, ErrDocumentFileRequired
	}
	if err := utils.ValidatePDF(file, header, MaxDocumentSizeBytes); err != nil {
		return nil, err
	}

	thesis, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}
	if !isDocumentEligible(thesis.Status) {
		return nil, ErrDocumentThesisNotLive
	}
	// Only the owning student uploads documents.
	if !isThesisOwner(thesis, actor.UserID) {
		return nil, ErrForbidden
	}

	// Auto-increment version per document type (and chapter for drafts).
	version := 1
	latest, err := uc.documentRepo.FindLatestByType(ctx, thesisID, req.DocumentType, req.ChapterNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if latest != nil {
		version = latest.Version + 1
	}

	// Storage path convention: theses/{thesis_id}/{document_type}/v{version}_{filename}
	baseName := sanitizeFileName(header.Filename)
	storagePath := "theses/" + thesisID.String() + "/" + req.DocumentType +
		"/v" + strconv.Itoa(version) + "_" + baseName
	fileURL, err := uc.storage.Upload(ctx, storagePath, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	doc := &entity.Document{
		ThesisID:      thesisID,
		UploadedBy:    actor.UserID,
		DocumentType:  req.DocumentType,
		ChapterNumber: req.ChapterNumber,
		Version:       version,
		FileName:      baseName,
		FileURL:       fileURL,
		FileSize:      &header.Size,
		Status:        entity.DocStatusPendingReview,
	}
	if err := uc.documentRepo.Create(ctx, doc); err != nil {
		return nil, err
	}

	// Notify the supervisors (async, non-fatal).
	emails := supervisorEmails(thesis)
	if len(emails) > 0 {
		go func() {
			_ = uc.emailSvc.SendDocumentUploaded(context.Background(), emails, doc)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: userIDs(thesis.Supervisors),
				Title:   "Dokumen Baru Diunggah",
				Message: thesis.Student.FullName + " mengunggah " + documentTypeLabel(doc.DocumentType) + " versi " + strconv.Itoa(doc.Version) + ".",
				Type:    "document",
				Link:    notification.Path("/theses/%s/documents", thesisID),
			})
		}()
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionDocumentUploaded,
		EntityType: "document",
		EntityID:   &doc.ID,
		NewValue: map[string]interface{}{
			"document_type": doc.DocumentType,
			"version":       doc.Version,
			"file_name":     doc.FileName,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})

	return toDocumentDetail(doc), nil
}

// List returns the active (latest version per type) documents of a thesis.
// Access: owner + supervisor + examiner + admin + kaprodi.
func (uc *DocumentUseCase) List(ctx context.Context, thesisID uuid.UUID, filter domainRepo.DocumentFilter, userID uuid.UUID, role string) ([]*DocumentDetail, int64, error) {
	if err := uc.canView(ctx, thesisID, userID, role); err != nil {
		return nil, 0, err
	}
	docs, total, err := uc.documentRepo.FindByThesisID(ctx, thesisID, filter)
	if err != nil {
		return nil, 0, err
	}
	details := make([]*DocumentDetail, 0, len(docs))
	for _, d := range docs {
		details = append(details, toDocumentDetail(d))
	}
	return details, total, nil
}

// GetByID returns a single document (same access as List).
func (uc *DocumentUseCase) GetByID(ctx context.Context, thesisID, id, userID uuid.UUID, role string) (*DocumentDetail, error) {
	if err := uc.canView(ctx, thesisID, userID, role); err != nil {
		return nil, err
	}
	doc, err := uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if doc.ThesisID != thesisID {
		return nil, ErrDocumentNotFound
	}
	return toDocumentDetail(doc), nil
}

// Download generates a presigned URL and logs the download (same access as List).
func (uc *DocumentUseCase) Download(ctx context.Context, thesisID, id, userID uuid.UUID, role string, actor Actor) (*DownloadResult, error) {
	if err := uc.canView(ctx, thesisID, userID, role); err != nil {
		return nil, err
	}
	doc, err := uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if doc.ThesisID != thesisID {
		return nil, ErrDocumentNotFound
	}

	url, err := uc.storage.GeneratePresignedURL(ctx, doc.FileURL, DocumentDownloadExpirySecs)
	if err != nil {
		return nil, err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionDocumentDownloaded,
		EntityType: "document",
		EntityID:   &doc.ID,
		NewValue:   map[string]interface{}{"document_type": doc.DocumentType, "version": doc.Version},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return &DownloadResult{DownloadURL: url, ExpiresIn: DocumentDownloadExpirySecs}, nil
}

// History returns all versions of a document type, newest first (same access as List).
func (uc *DocumentUseCase) History(ctx context.Context, thesisID uuid.UUID, docType string, chapterNum *int, userID uuid.UUID, role string) ([]*DocumentDetail, error) {
	if err := uc.canView(ctx, thesisID, userID, role); err != nil {
		return nil, err
	}
	docs, err := uc.documentRepo.GetVersionHistory(ctx, thesisID, docType, chapterNum)
	if err != nil {
		return nil, err
	}
	details := make([]*DocumentDetail, 0, len(docs))
	for _, d := range docs {
		details = append(details, toDocumentDetail(d))
	}
	return details, nil
}

// Review approves or requests revision of a pending document (supervisor only).
func (uc *DocumentUseCase) Review(ctx context.Context, id uuid.UUID, decision, notes string, actor Actor) (*DocumentDetail, error) {
	if decision != DocDecisionApproved && decision != DocDecisionRevisionReq {
		return nil, ErrInvalidReviewDecision
	}

	doc, err := uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if doc.Status != entity.DocStatusPendingReview {
		return nil, ErrDocumentNotPending
	}

	// Only a supervisor of the thesis may review.
	ok, err := uc.access.IsSupervisor(ctx, actor.UserID, doc.ThesisID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotDocumentReviewer
	}

	if err := uc.documentRepo.UpdateStatus(ctx, id, decision, actor.UserID, notes); err != nil {
		// gorm.ErrRecordNotFound means the status guard failed (already reviewed).
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotPending
		}
		return nil, err
	}
	// Re-fetch so the reviewer association is populated in the response.
	doc, err = uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Notify the student (async, non-fatal).
	thesis, err := uc.thesisRepo.FindByID(ctx, doc.ThesisID)
	if err == nil {
		studentEmail := thesis.Student.Email
		go func() {
			_ = uc.emailSvc.SendDocumentReviewed(context.Background(), studentEmail, doc, decision)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: []uuid.UUID{thesis.Student.ID},
				Title:   "Dokumen Diperbarui",
				Message: documentTypeLabel(doc.DocumentType) + " versi " + strconv.Itoa(doc.Version) + " Anda telah direview oleh dosen pembimbing.",
				Type:    "document",
				Link:    notification.Path("/theses/%s/documents", doc.ThesisID),
			})
		}()
	}

	action := audit.ActionDocumentApproved
	if decision == DocDecisionRevisionReq {
		action = audit.ActionDocumentRevision
	}
	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     action,
		EntityType: "document",
		EntityID:   &doc.ID,
		OldValue:   map[string]interface{}{"status": entity.DocStatusPendingReview},
		NewValue:   map[string]interface{}{"status": decision, "notes": notes},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return toDocumentDetail(doc), nil
}

// CanSubmitSeminar reports whether the seminar_doc is approved (Job 07 gate).
func (uc *DocumentUseCase) CanSubmitSeminar(ctx context.Context, thesisID uuid.UUID) (bool, error) {
	return uc.documentRepo.IsDocumentApproved(ctx, thesisID, entity.DocTypeSeminarDoc)
}

// CanSubmitDefense reports whether the defense_doc is approved (Job 07 gate).
func (uc *DocumentUseCase) CanSubmitDefense(ctx context.Context, thesisID uuid.UUID) (bool, error) {
	return uc.documentRepo.IsDocumentApproved(ctx, thesisID, entity.DocTypeDefenseDoc)
}

// IsFinalThesisApproved reports whether the final_thesis document is approved
// (used by the graduation gate in Job 09).
func (uc *DocumentUseCase) IsFinalThesisApproved(ctx context.Context, thesisID uuid.UUID) (bool, error) {
	return uc.documentRepo.IsDocumentApproved(ctx, thesisID, entity.DocTypeFinalThesis)
}

// canView enforces the read access rule for documents.
// Per job spec: Mahasiswa pemilik + Dosen Pembimbing + Dosen Penguji + Admin + Kaprodi.
func (uc *DocumentUseCase) canView(ctx context.Context, thesisID, userID uuid.UUID, role string) error {
	switch role {
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		return nil
	case ThesisRoleMahasiswa:
		ok, err := uc.access.IsThesisOwner(ctx, userID, thesisID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrForbidden
		}
		return nil
	case ThesisRoleDosenPembimbing:
		ok, err := uc.access.IsSupervisor(ctx, userID, thesisID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrForbidden
		}
		return nil
	case ThesisRoleDosenPenguji:
		ok, err := uc.access.IsExaminer(ctx, userID, thesisID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrForbidden
		}
		return nil
	default:
		return ErrForbidden
	}
}

// isDocumentEligible reports whether the thesis status allows document uploads.
func isDocumentEligible(status string) bool {
	switch status {
	case "in_progress", "seminar_ready", "seminar_done", "defense_ready", "defense_done", "graduated":
		return true
	default:
		return false
	}
}

// toDocumentDetail maps a persisted document into the API response shape.
func toDocumentDetail(d *entity.Document) *DocumentDetail {
	detail := &DocumentDetail{
		ID:            d.ID,
		DocumentType:  d.DocumentType,
		ChapterNumber: d.ChapterNumber,
		Version:       d.Version,
		FileName:      d.FileName,
		FileSize:      d.FileSize,
		Status:        d.Status,
		ReviewerNotes: d.ReviewerNotes,
		ReviewedAt:    d.ReviewedAt,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
	if d.Reviewer != nil && d.Reviewer.ID != uuid.Nil {
		detail.Reviewer = &UserBrief{ID: d.Reviewer.ID, FullName: d.Reviewer.FullName}
	}
	return detail
}

// sanitizeFileName strips path components (both / and \ separators) so a
// client-supplied filename can never escape the storage root.
func sanitizeFileName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	return filepath.Base(name)
}
