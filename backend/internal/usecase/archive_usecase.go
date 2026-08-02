package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/service"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

// MaxArchiveSizeBytes caps final-thesis PDFs at 25 MB (Job 10).
const MaxArchiveSizeBytes = 25 << 20

// ArchiveDownloadExpirySecs is the presigned URL lifetime for archives.
const ArchiveDownloadExpirySecs = 1800 // 30 minutes

// MinKeywords is the minimum number of archive keywords (Job 10).
const MinKeywords = 3

// MinAbstractWords is the minimum word count of the Indonesian abstract.
const MinAbstractWords = 50

var (
	ErrArchiveNotFound       = errors.New("arsip tidak ditemukan")
	ErrArchiveExists         = errors.New("thesis ini sudah memiliki arsip")
	ErrArchiveThesisNotGrad  = errors.New("hanya thesis berstatus graduated yang dapat diarsipkan")
	ErrArchiveAbstractShort  = errors.New("abstract_id minimal 50 kata")
	ErrArchiveKeywordsShort  = errors.New("minimal 3 kata kunci")
	ErrArchiveInvalidYear    = errors.New("graduation_year tidak valid")
	ErrArchiveFileRequired   = errors.New("file PDF wajib diunggah")
	ErrArchiveDownloadDenied = errors.New("akses download arsip ditolak")
)

// CreateArchiveRequest carries the validated multipart form fields (Job 10).
type CreateArchiveRequest struct {
	AbstractID     string
	AbstractEN     string
	Keywords       []string
	GraduationYear int
}

// ArchiveDetail is the API response shape for an archive (Job 10).
// FileURL is intentionally excluded — downloads go through the dedicated
// presigned-URL endpoint.
type ArchiveDetail struct {
	ID             uuid.UUID           `json:"id"`
	ThesisID       uuid.UUID           `json:"thesis_id"`
	Title          string              `json:"title"`
	AbstractID     string              `json:"abstract_id"`
	AbstractEN     *string             `json:"abstract_en,omitempty"`
	Keywords       []string            `json:"keywords"`
	GraduationYear int                 `json:"graduation_year"`
	FieldOfStudy   string              `json:"field_of_study"`
	StudyProgram   string              `json:"study_program"`
	Student        *ThesisStudentBrief `json:"student,omitempty"`
	Supervisors    []*UserBrief        `json:"supervisors,omitempty"`
	FileName       string              `json:"file_name"`
	ArchivedAt     time.Time           `json:"archived_at"`
}

// ArchiveDownloadResult wraps the presigned download URL (Job 10).
type ArchiveDownloadResult struct {
	DownloadURL string `json:"download_url"`
	ExpiresIn   int    `json:"expires_in"`
}

// ArchiveUseCase contains business logic for the digital archive module (Job 10).
type ArchiveUseCase struct {
	archiveRepo repository.ArchiveRepository
	thesisRepo  repository.ThesisRepository
	storage     service.StorageService
	access      *ThesisAccess
	emailSvc    email.EmailService
	auditSvc    *audit.AuditService
}

func NewArchiveUseCase(
	archiveRepo repository.ArchiveRepository,
	thesisRepo repository.ThesisRepository,
	storage service.StorageService,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
) *ArchiveUseCase {
	return &ArchiveUseCase{
		archiveRepo: archiveRepo,
		thesisRepo:  thesisRepo,
		storage:     storage,
		access:      NewThesisAccess(thesisRepo),
		emailSvc:    emailSvc,
		auditSvc:    auditSvc,
	}
}

// Create validates the file + metadata, stores the PDF, and records the archive
// (Mahasiswa pemilik + Admin only; the role comes from the auth middleware).
func (uc *ArchiveUseCase) Create(
	ctx context.Context,
	thesisID uuid.UUID,
	file multipart.File,
	header *multipart.FileHeader,
	req CreateArchiveRequest,
	actor Actor,
	role string,
) (*ArchiveDetail, error) {
	req.AbstractID = strings.TrimSpace(req.AbstractID)
	if wordCount(req.AbstractID) < MinAbstractWords {
		return nil, ErrArchiveAbstractShort
	}
	if len(req.Keywords) < MinKeywords {
		return nil, ErrArchiveKeywordsShort
	}
	if req.GraduationYear < 2000 || req.GraduationYear > time.Now().Year() {
		return nil, ErrArchiveInvalidYear
	}
	if file == nil || header == nil {
		return nil, ErrArchiveFileRequired
	}
	if err := utils.ValidatePDF(file, header, MaxArchiveSizeBytes); err != nil {
		return nil, err
	}

	t, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}
	if t.Status != "graduated" {
		return nil, ErrArchiveThesisNotGrad
	}
	if !isThesisOwner(t, actor.UserID) && role != ThesisRoleAdminFakultas {
		return nil, ErrForbidden
	}

	// Unique archive per thesis.
	existing, err := uc.archiveRepo.FindByThesisID(ctx, thesisID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrArchiveExists
	}

	// Storage path convention: archives/{graduation_year}/{thesis_id}/{filename}
	baseName := sanitizeFileName(header.Filename)
	storagePath := "archives/" + strconv.Itoa(req.GraduationYear) + "/" + thesisID.String() + "/" + baseName
	fileURL, err := uc.storage.Upload(ctx, storagePath, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	archive := &entity.ThesisArchive{
		ThesisID:       thesisID,
		FileURL:        fileURL,
		FileName:       baseName,
		AbstractID:     req.AbstractID,
		Keywords:       req.Keywords,
		GraduationYear: req.GraduationYear,
		ArchivedBy:     actor.UserID,
	}
	if req.AbstractEN != "" {
		archive.AbstractEN = &req.AbstractEN
	}
	if err := uc.archiveRepo.Create(ctx, archive); err != nil {
		return nil, err
	}

	// Notify the student (async, non-fatal). Capture a local copy — `archive` is
	// reassigned below and goroutines capture by reference.
	created := archive
	go func() { _ = uc.emailSvc.SendArchiveCreated(context.Background(), t.Student.Email, created) }()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionArchiveCreated,
		EntityType: "archive",
		EntityID:   &archive.ID,
		NewValue: map[string]interface{}{
			"thesis_id":       thesisID,
			"graduation_year": req.GraduationYear,
			"file_name":       baseName,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})

	archive, err = uc.archiveRepo.FindByID(ctx, archive.ID)
	if err != nil {
		return nil, err
	}
	return toArchiveDetail(archive), nil
}

// Search lists archives with optional full-text search and filters (all roles).
func (uc *ArchiveUseCase) Search(ctx context.Context, filter repository.ArchiveFilter) ([]*ArchiveDetail, int64, error) {
	archives, total, err := uc.archiveRepo.Search(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	details := make([]*ArchiveDetail, 0, len(archives))
	for _, a := range archives {
		details = append(details, toArchiveDetail(a))
	}
	return details, total, nil
}

// GetByID returns a single archive (all roles).
func (uc *ArchiveUseCase) GetByID(ctx context.Context, id uuid.UUID) (*ArchiveDetail, error) {
	archive, err := uc.archiveRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArchiveNotFound
		}
		return nil, err
	}
	return toArchiveDetail(archive), nil
}

// GetByThesisID is the shortcut for GET /theses/:thesis_id/archive.
func (uc *ArchiveUseCase) GetByThesisID(ctx context.Context, thesisID uuid.UUID) (*ArchiveDetail, error) {
	archive, err := uc.archiveRepo.FindByThesisID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArchiveNotFound
		}
		return nil, err
	}
	return toArchiveDetail(archive), nil
}

// Download generates a presigned URL and logs the download (role-based access).
func (uc *ArchiveUseCase) Download(ctx context.Context, id, userID uuid.UUID, role string, actor Actor) (*ArchiveDownloadResult, error) {
	archive, err := uc.archiveRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArchiveNotFound
		}
		return nil, err
	}

	// Access: Mahasiswa → only own thesis; all other roles → any archive.
	if role == ThesisRoleMahasiswa {
		ok, err := uc.access.IsThesisOwner(ctx, userID, archive.ThesisID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrArchiveDownloadDenied
		}
	}

	url, err := uc.storage.GeneratePresignedURL(ctx, archive.FileURL, ArchiveDownloadExpirySecs)
	if err != nil {
		return nil, err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionArchiveDownloaded,
		EntityType: "archive",
		EntityID:   &archive.ID,
		NewValue:   map[string]interface{}{"thesis_id": archive.ThesisID},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return &ArchiveDownloadResult{DownloadURL: url, ExpiresIn: ArchiveDownloadExpirySecs}, nil
}

// Stats returns aggregate archive counts (Admin + Kaprodi only, enforced by route).
func (uc *ArchiveUseCase) Stats(ctx context.Context) (*repository.ArchiveStats, error) {
	return uc.archiveRepo.Stats(ctx)
}

// toArchiveDetail maps a persisted archive into the API response shape.
func toArchiveDetail(a *entity.ThesisArchive) *ArchiveDetail {
	d := &ArchiveDetail{
		ID:             a.ID,
		ThesisID:       a.ThesisID,
		Title:          a.Thesis.Title,
		AbstractID:     a.AbstractID,
		AbstractEN:     a.AbstractEN,
		Keywords:       a.Keywords,
		GraduationYear: a.GraduationYear,
		FileName:       a.FileName,
		ArchivedAt:     a.ArchivedAt,
	}
	if a.Thesis.Student.ID != uuid.Nil {
		d.Student = &ThesisStudentBrief{
			FullName: a.Thesis.Student.FullName,
			Nim:      a.Thesis.Student.NimNidn,
		}
		d.StudyProgram = derefStr(a.Thesis.Student.StudyProgram)
	}
	if a.Thesis.FieldOfStudy != nil {
		d.FieldOfStudy = *a.Thesis.FieldOfStudy
	}
	for _, s := range a.Thesis.Supervisors {
		if s.ID != uuid.Nil {
			d.Supervisors = append(d.Supervisors, &UserBrief{ID: s.ID, FullName: s.FullName})
		}
	}
	return d
}

// wordCount counts whitespace-separated words.
func wordCount(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	return len(strings.Fields(s))
}
