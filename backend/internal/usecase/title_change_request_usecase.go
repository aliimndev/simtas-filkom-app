package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/notification"
)

// Title change request status values (kept in sync with the DB constraint).
const (
	// pgUniqueViolation is the PostgreSQL SQLSTATE for a unique constraint
	// violation (used to translate a create race into a 409 Conflict).
	pgUniqueViolation = "23505"
)

const (
	TitleChangeStatusPending   = "PENDING"
	TitleChangeStatusApproved  = "APPROVED"
	TitleChangeStatusRejected  = "REJECTED"
	TitleChangeStatusCancelled = "CANCELLED"
)

var (
	ErrTitleChangeNotFound       = errors.New("permintaan perubahan judul tidak ditemukan")
	ErrTitleChangeForbidden      = errors.New("akses ditolak")
	ErrTitleChangeNotEligible    = errors.New("perubahan judul hanya dapat diajukan pada status approved atau in_progress")
	ErrNoSupervisorAssigned      = errors.New("thesis belum memiliki pembimbing aktif")
	ErrPendingTitleChangeExists  = errors.New("sudah ada permintaan perubahan judul yang sedang diproses")
	ErrTitleChangeNotPending     = errors.New("permintaan perubahan judul tidak dalam status pending")
	ErrTitleChangeTitleTooShort  = errors.New("judul minimal 10 kata")
	ErrTitleChangeTitleTooLong   = errors.New("judul maksimal 500 karakter")
	ErrTitleChangeReviewNotesReq = errors.New("catatan penolakan wajib diisi")
	ErrTitleChangeNotSupervisor  = errors.New("hanya dosen pembimbing thesis ini yang dapat mereview")
)

// CreateTitleChangeRequest is the payload for POST /theses/:thesis_id/title-change-requests.
type CreateTitleChangeRequest struct {
	RequestedTitle string  `json:"requested_title" binding:"required"`
	Reason         *string `json:"reason"`
}

// ReviewTitleChangeRequest is the payload for approve/reject actions.
type ReviewTitleChangeRequest struct {
	ReviewNotes *string `json:"review_notes"`
}

// TitleChangeRequestDetail is the API response shape for a title change request.
type TitleChangeRequestDetail struct {
	ID             uuid.UUID  `json:"id"`
	ThesisID       uuid.UUID  `json:"thesis_id"`
	PreviousTitle  string     `json:"previous_title"`
	RequestedTitle string     `json:"requested_title"`
	Reason         *string    `json:"reason,omitempty"`
	Status         string     `json:"status"`
	RequestedBy    *UserBrief `json:"requested_by,omitempty"`
	ReviewedBy     *UserBrief `json:"reviewed_by,omitempty"`
	ReviewNotes    *string    `json:"review_notes,omitempty"`
	CancelledBy    *UserBrief `json:"cancelled_by,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TitleChangeRequestUseCase contains the business logic for title change requests.
type TitleChangeRequestUseCase struct {
	titleChangeRepo domainRepo.TitleChangeRequestRepository
	thesisRepo      domainRepo.ThesisRepository
	access          *ThesisAccess
	emailSvc        email.EmailService
	auditSvc        *audit.AuditService
	notifSvc        *notification.NotificationService
}

func NewTitleChangeRequestUseCase(
	titleChangeRepo domainRepo.TitleChangeRequestRepository,
	thesisRepo domainRepo.ThesisRepository,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
	notifSvc *notification.NotificationService,
) *TitleChangeRequestUseCase {
	return &TitleChangeRequestUseCase{
		titleChangeRepo: titleChangeRepo,
		thesisRepo:      thesisRepo,
		access:          NewThesisAccess(thesisRepo),
		emailSvc:        emailSvc,
		auditSvc:        auditSvc,
		notifSvc:        notifSvc,
	}
}

// Submit creates a PENDING title change request (Mahasiswa pemilik only).
func (uc *TitleChangeRequestUseCase) Submit(ctx context.Context, thesisID uuid.UUID, req CreateTitleChangeRequest, actor Actor) (*TitleChangeRequestDetail, error) {
	req.RequestedTitle = strings.TrimSpace(req.RequestedTitle)
	if err := validateTitleChange(req.RequestedTitle); err != nil {
		return nil, err
	}

	thesis, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}

	// Only the thesis owner may submit a title change request.
	if !isThesisOwner(thesis, actor.UserID) {
		return nil, ErrTitleChangeForbidden
	}

	// Title changes are only allowed while the thesis is approved or in progress.
	if thesis.Status != "approved" && thesis.Status != "in_progress" {
		return nil, ErrTitleChangeNotEligible
	}
	if len(thesis.Supervisors) == 0 {
		return nil, ErrNoSupervisorAssigned
	}

	// Only one PENDING request per thesis.
	pendingFound, err := uc.titleChangeRepo.FindPendingByThesisID(ctx, thesisID)
	if err == nil && pendingFound != nil {
		return nil, ErrPendingTitleChangeExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tcr := &entity.TitleChangeRequest{
		ThesisID:       thesisID,
		RequestedByID:  actor.UserID,
		PreviousTitle:  thesis.Title,
		RequestedTitle: req.RequestedTitle,
		Reason:         req.Reason,
		Status:         TitleChangeStatusPending,
	}
	if err := uc.titleChangeRepo.Create(ctx, tcr); err != nil {
		// Two students could race the pending check; the partial unique index on
		// (thesis_id) WHERE status = 'PENDING' is the final arbiter (ADR-001 §7.5).
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, ErrPendingTitleChangeExists
		}
		return nil, err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionTitleChangeRequested,
		EntityType: "title_change_request",
		EntityID:   &tcr.ID,
		NewValue:   map[string]interface{}{"previous_title": tcr.PreviousTitle, "requested_title": tcr.RequestedTitle},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	// Notify the student (confirmation) and the assigned supervisors (review prompt),
	// async and non-fatal like every other email in this codebase.
	go func() {
		_ = uc.emailSvc.SendTitleChangeRequested(context.Background(), subscriberEmails(thesis, true), thesis, tcr)
		uc.notifSvc.Notify(notification.Params{
			UserIDs: userIDs(thesis.Supervisors),
			Title:   "Pengajuan Perubahan Judul",
			Message: thesis.Student.FullName + " mengajukan perubahan judul skripsi.",
			Type:    "title_change",
			Link:    notification.Path("/theses/%s", thesis.ID),
		})
	}()

	return toTitleChangeRequestDetail(tcr), nil
}

// List returns the title change requests for a thesis.
// Read access: thesis owner, assigned supervisors, Kaprodi, and Admin.
func (uc *TitleChangeRequestUseCase) List(ctx context.Context, thesisID uuid.UUID, userID uuid.UUID, role string) ([]*TitleChangeRequestDetail, error) {
	if _, err := uc.thesisRepo.FindByID(ctx, thesisID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}
	if !uc.canRead(ctx, thesisID, userID, role) {
		return nil, ErrTitleChangeForbidden
	}

	reqs, err := uc.titleChangeRepo.FindByThesisID(ctx, thesisID)
	if err != nil {
		return nil, err
	}
	out := make([]*TitleChangeRequestDetail, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, toTitleChangeRequestDetail(r))
	}
	return out, nil
}

// Cancel retracts a PENDING title change request (Mahasiswa pemilik only).
func (uc *TitleChangeRequestUseCase) Cancel(ctx context.Context, requestID uuid.UUID, actor Actor) (*TitleChangeRequestDetail, error) {
	tcr, err := uc.titleChangeRepo.FindByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTitleChangeNotFound
		}
		return nil, err
	}
	if tcr.RequestedByID != actor.UserID {
		return nil, ErrTitleChangeForbidden
	}
	if tcr.Status != TitleChangeStatusPending {
		return nil, ErrTitleChangeNotPending
	}

	now := time.Now()
	if err := uc.titleChangeRepo.UpdateStatus(ctx, tcr.ID, TitleChangeStatusCancelled, nil, nil, nil, &actor.UserID, &now); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTitleChangeNotPending
		}
		return nil, err
	}

	tcr.Status = TitleChangeStatusCancelled
	tcr.CancelledByID = &actor.UserID
	tcr.CancelledAt = &now

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionTitleChangeCancelled,
		EntityType: "title_change_request",
		EntityID:   &tcr.ID,
		NewValue:   map[string]interface{}{"status": TitleChangeStatusCancelled},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	// Notify the assigned supervisors for transparency (async, non-fatal).
	// The thesis (with supervisors) was preloaded by FindByID, so no refetch is
	// needed and a post-commit error cannot be misreported as a failure.
	go func() {
		_ = uc.emailSvc.SendTitleChangeCancelled(context.Background(), supervisorEmails(&tcr.Thesis), &tcr.Thesis, tcr)
		uc.notifSvc.Notify(notification.Params{
			UserIDs: userIDs(tcr.Thesis.Supervisors),
			Title:   "Perubahan Judul Dibatalkan",
			Message: tcr.Thesis.Student.FullName + " membatalkan permintaan perubahan judul skripsi.",
			Type:    "title_change",
			Link:    notification.Path("/theses/%s", tcr.ThesisID),
		})
	}()

	return toTitleChangeRequestDetail(tcr), nil
}

// Approve approves a PENDING title change request (assigned Dosen Pembimbing
// only). The request status and the thesis title are updated atomically in a
// single database transaction (ADR-001 §5.3).
func (uc *TitleChangeRequestUseCase) Approve(ctx context.Context, requestID uuid.UUID, req ReviewTitleChangeRequest, actor Actor) (*TitleChangeRequestDetail, error) {
	tcr, err := uc.titleChangeRepo.FindByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTitleChangeNotFound
		}
		return nil, err
	}
	if tcr.Status != TitleChangeStatusPending {
		return nil, ErrTitleChangeNotPending
	}
	if !isSupervisor(&tcr.Thesis, actor.UserID) {
		return nil, ErrTitleChangeNotSupervisor
	}

	now := time.Now()
	if err := uc.titleChangeRepo.Approve(ctx, tcr.ID, actor.UserID, now, req.ReviewNotes, tcr.RequestedTitle); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTitleChangeNotPending
		}
		return nil, err
	}

	tcr.Status = TitleChangeStatusApproved
	tcr.ReviewedByID = &actor.UserID
	tcr.ReviewedAt = &now
	tcr.ReviewNotes = req.ReviewNotes

	// Re-fetch so the preloaded Reviewer association reflects the transition
	// (the in-memory copy only carries the raw ID).
	tcr, err = uc.titleChangeRepo.FindByID(ctx, tcr.ID)
	if err != nil {
		return nil, err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionTitleChangeApproved,
		EntityType: "title_change_request",
		EntityID:   &tcr.ID,
		OldValue:   map[string]interface{}{"status": TitleChangeStatusPending, "previous_title": tcr.PreviousTitle},
		NewValue:   map[string]interface{}{"status": TitleChangeStatusApproved, "requested_title": tcr.RequestedTitle},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionThesisTitleUpdated,
		EntityType: "thesis",
		EntityID:   &tcr.ThesisID,
		OldValue:   map[string]interface{}{"title": tcr.PreviousTitle},
		NewValue:   map[string]interface{}{"title": tcr.RequestedTitle},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	// Notify the student that the title change was approved (async, non-fatal).
	go func() {
		_ = uc.emailSvc.SendTitleChangeApproved(context.Background(), []string{tcr.Thesis.Student.Email}, &tcr.Thesis, tcr)
		uc.notifSvc.Notify(notification.Params{
			UserIDs: []uuid.UUID{tcr.Thesis.Student.ID},
			Title:   "Perubahan Judul Disetujui",
			Message: "Permintaan perubahan judul skripsi Anda telah disetujui oleh Dosen Pembimbing.",
			Type:    "title_change",
			Link:    notification.Path("/theses/%s", tcr.ThesisID),
		})
	}()

	return toTitleChangeRequestDetail(tcr), nil
}

// Reject rejects a PENDING title change request (assigned Dosen Pembimbing
// only). A rejection note is required (ADR-001 §5.4).
func (uc *TitleChangeRequestUseCase) Reject(ctx context.Context, requestID uuid.UUID, req ReviewTitleChangeRequest, actor Actor) (*TitleChangeRequestDetail, error) {
	tcr, err := uc.titleChangeRepo.FindByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTitleChangeNotFound
		}
		return nil, err
	}
	if tcr.Status != TitleChangeStatusPending {
		return nil, ErrTitleChangeNotPending
	}
	if req.ReviewNotes == nil || strings.TrimSpace(*req.ReviewNotes) == "" {
		return nil, ErrTitleChangeReviewNotesReq
	}
	if !isSupervisor(&tcr.Thesis, actor.UserID) {
		return nil, ErrTitleChangeNotSupervisor
	}

	now := time.Now()
	if err := uc.titleChangeRepo.UpdateStatus(ctx, tcr.ID, TitleChangeStatusRejected, &actor.UserID, &now, req.ReviewNotes, nil, nil); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTitleChangeNotPending
		}
		return nil, err
	}

	tcr.Status = TitleChangeStatusRejected
	tcr.ReviewedByID = &actor.UserID
	tcr.ReviewedAt = &now
	tcr.ReviewNotes = req.ReviewNotes

	// Re-fetch so the preloaded Reviewer association reflects the transition.
	tcr, err = uc.titleChangeRepo.FindByID(ctx, tcr.ID)
	if err != nil {
		return nil, err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionTitleChangeRejected,
		EntityType: "title_change_request",
		EntityID:   &tcr.ID,
		OldValue:   map[string]interface{}{"status": TitleChangeStatusPending},
		NewValue:   map[string]interface{}{"status": TitleChangeStatusRejected, "review_notes": *req.ReviewNotes},
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	// Notify the student that the title change was rejected (async, non-fatal).
	go func() {
		_ = uc.emailSvc.SendTitleChangeRejected(context.Background(), []string{tcr.Thesis.Student.Email}, &tcr.Thesis, tcr)
		uc.notifSvc.Notify(notification.Params{
			UserIDs: []uuid.UUID{tcr.Thesis.Student.ID},
			Title:   "Perubahan Judul Ditolak",
			Message: "Permintaan perubahan judul skripsi Anda ditolak oleh Dosen Pembimbing.",
			Type:    "title_change",
			Link:    notification.Path("/theses/%s", tcr.ThesisID),
		})
	}()

	return toTitleChangeRequestDetail(tcr), nil
}

// ListPendingForSupervisor returns the PENDING review queue for a dosen
// pembimbing: all pending requests across theses they supervise.
func (uc *TitleChangeRequestUseCase) ListPendingForSupervisor(ctx context.Context, supervisorID uuid.UUID) ([]*TitleChangeRequestDetail, error) {
	reqs, err := uc.titleChangeRepo.FindPendingBySupervisorID(ctx, supervisorID)
	if err != nil {
		return nil, err
	}
	out := make([]*TitleChangeRequestDetail, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, toTitleChangeRequestDetail(r))
	}
	return out, nil
}

// canRead reports whether userID may view title change requests for the thesis.
func (uc *TitleChangeRequestUseCase) canRead(ctx context.Context, thesisID uuid.UUID, userID uuid.UUID, role string) bool {
	switch role {
	case ThesisRoleAdminFakultas, ThesisRoleKaprodi:
		return true
	default:
		owner, err := uc.access.IsThesisOwner(ctx, userID, thesisID)
		if err != nil {
			return false
		}
		if owner {
			return true
		}
		sup, err := uc.access.IsSupervisor(ctx, userID, thesisID)
		if err != nil {
			return false
		}
		return sup
	}
}

// validateTitleChange enforces the same title rules as thesis submission.
func validateTitleChange(title string) error {
	words := strings.Fields(title)
	if len(words) < 10 {
		return ErrTitleChangeTitleTooShort
	}
	if len(title) > 500 {
		return ErrTitleChangeTitleTooLong
	}
	return nil
}

// toTitleChangeRequestDetail maps an entity to the API shape.
// A preloaded user association is trusted; raw IDs alone do not synthesize a brief.
func toTitleChangeRequestDetail(r *entity.TitleChangeRequest) *TitleChangeRequestDetail {
	d := &TitleChangeRequestDetail{
		ID:             r.ID,
		ThesisID:       r.ThesisID,
		PreviousTitle:  r.PreviousTitle,
		RequestedTitle: r.RequestedTitle,
		Reason:         r.Reason,
		Status:         r.Status,
		ReviewNotes:    r.ReviewNotes,
		CancelledAt:    r.CancelledAt,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.RequestedBy != nil {
		d.RequestedBy = userToBrief(r.RequestedBy)
	}
	if r.ReviewedBy != nil {
		d.ReviewedBy = userToBrief(r.ReviewedBy)
	}
	if r.CancelledBy != nil {
		d.CancelledBy = userToBrief(r.CancelledBy)
	}
	return d
}

func userToBrief(u *entity.User) *UserBrief {
	if u == nil {
		return nil
	}
	return &UserBrief{ID: u.ID, FullName: u.FullName, NimNidn: u.NimNidn}
}

// subscriberEmails builds the recipient list for a "requested" notification:
// the student first, then all assigned supervisors.
func subscriberEmails(thesis *entity.Thesis, includeStudent bool) []string {
	out := make([]string, 0)
	if includeStudent && thesis.Student.Email != "" {
		out = append(out, thesis.Student.Email)
	}
	return append(out, supervisorEmails(thesis)...)
}
