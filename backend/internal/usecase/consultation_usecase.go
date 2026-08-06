package usecase

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/notification"
)

// Consultation status values (kept in sync with the DB check constraint).
const (
	ConsultationStatusPending  = "pending"
	ConsultationStatusApproved = "approved"
)

var (
	ErrConsultationNotFound    = errors.New("log konsultasi tidak ditemukan")
	ErrConsultationDateFuture  = errors.New("tanggal konsultasi tidak boleh di masa depan")
	ErrInvalidDateFormat       = errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	ErrTopicsDiscussedRequired = errors.New("topics_discussed tidak boleh kosong")
	ErrThesisNotInProgress     = errors.New("thesis harus berstatus in_progress atau lebih lanjut")
	ErrConsultationAlreadyDone = errors.New("log konsultasi sudah disetujui")
	ErrNotConsultationCreator  = errors.New("hanya pembuat log yang dapat mengubah log ini")
	ErrNotSupervisorOfThesis   = errors.New("hanya dosen pembimbing thesis ini yang dapat menyetujui")
)

// CreateConsultationRequest is the payload for POST /theses/:thesis_id/consultations.
type CreateConsultationRequest struct {
	ConsultationDate string  `json:"consultation_date" binding:"required"` // YYYY-MM-DD
	TopicsDiscussed  string  `json:"topics_discussed" binding:"required"`
	Notes            *string `json:"notes"`
	FollowUp         *string `json:"follow_up"`
	AttachmentURL    *string `json:"attachment_url"`
}

// UpdateConsultationRequest is the payload for PUT /theses/:thesis_id/consultations/:id.
type UpdateConsultationRequest struct {
	ConsultationDate *string `json:"consultation_date"`
	TopicsDiscussed  *string `json:"topics_discussed"`
	Notes            *string `json:"notes"`
	FollowUp         *string `json:"follow_up"`
	AttachmentURL    *string `json:"attachment_url"`
}

// ConsultationDetail is the API response shape for a single consultation log.
type ConsultationDetail struct {
	ID               uuid.UUID  `json:"id"`
	ThesisID         uuid.UUID  `json:"thesis_id"`
	CreatedBy        uuid.UUID  `json:"created_by"`
	Creator          *UserBrief `json:"creator,omitempty"`
	ConsultationDate string     `json:"consultation_date"` // YYYY-MM-DD
	TopicsDiscussed  string     `json:"topics_discussed"`
	Notes            *string    `json:"notes,omitempty"`
	FollowUp         *string    `json:"follow_up,omitempty"`
	AttachmentURL    *string    `json:"attachment_url,omitempty"`
	Status           string     `json:"status"`
	ApprovedBy       *uuid.UUID `json:"approved_by,omitempty"`
	Approver         *UserBrief `json:"approver,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UserBrief is a minimal user representation used in nested responses.
type UserBrief struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	// NimNidn is populated only where the underlying query preloads the user
	// (e.g. title change requests); omitted elsewhere.
	NimNidn *string `json:"nim_nidn,omitempty"`
}

// ConsultationSummary aggregates consultation statistics for a thesis.
type ConsultationSummary struct {
	TotalConsultations     int     `json:"total_consultations"`
	ApprovedCount          int     `json:"approved_count"`
	PendingCount           int     `json:"pending_count"`
	LastConsultationDate   *string `json:"last_consultation_date"`
	AverageIntervalDays    *int    `json:"average_interval_days"`
	ConsultationsThisMonth int     `json:"consultations_this_month"`
}

// ConsultationListResult bundles the paginated logs with their summary.
type ConsultationListResult struct {
	Consultations []*ConsultationDetail `json:"consultations"`
	Summary       ConsultationSummary   `json:"summary"`
}

// ConsultationUseCase contains business logic for the supervision module.
type ConsultationUseCase struct {
	consultationRepo domainRepo.ConsultationRepository
	thesisRepo       domainRepo.ThesisRepository
	access           *ThesisAccess
	emailSvc         email.EmailService
	auditSvc         *audit.AuditService
	notifSvc         *notification.NotificationService
}

func NewConsultationUseCase(
	consultationRepo domainRepo.ConsultationRepository,
	thesisRepo domainRepo.ThesisRepository,
	emailSvc email.EmailService,
	auditSvc *audit.AuditService,
	notifSvc *notification.NotificationService,
) *ConsultationUseCase {
	return &ConsultationUseCase{
		consultationRepo: consultationRepo,
		thesisRepo:       thesisRepo,
		access:           NewThesisAccess(thesisRepo),
		emailSvc:         emailSvc,
		auditSvc:         auditSvc,
		notifSvc:         notifSvc,
	}
}

// Create records a new consultation log (Mahasiswa pemilik or Dosen Pembimbing).
// The access check is identity-based (owner or supervisor), so no role is needed.
func (uc *ConsultationUseCase) Create(ctx context.Context, thesisID uuid.UUID, req CreateConsultationRequest, actor Actor) (*ConsultationDetail, error) {
	req.TopicsDiscussed = strings.TrimSpace(req.TopicsDiscussed)
	if req.TopicsDiscussed == "" {
		return nil, ErrTopicsDiscussedRequired
	}
	date, err := parseConsultationDate(req.ConsultationDate)
	if err != nil {
		return nil, err
	}

	thesis, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThesisNotFound
		}
		return nil, err
	}

	// Thesis must be in_progress or later (not submitted/approved/rejected).
	if !isConsultationEligible(thesis.Status) {
		return nil, ErrThesisNotInProgress
	}

	// Caller must be the owner or one of the supervisors.
	owner := isThesisOwner(thesis, actor.UserID)
	sup := isSupervisor(thesis, actor.UserID)
	if !owner && !sup {
		return nil, ErrForbidden
	}

	log := &entity.ConsultationLog{
		ThesisID:         thesisID,
		CreatedBy:        actor.UserID,
		ConsultationDate: date,
		TopicsDiscussed:  req.TopicsDiscussed,
		Notes:            req.Notes,
		FollowUp:         req.FollowUp,
		AttachmentURL:    req.AttachmentURL,
		Status:           ConsultationStatusPending,
	}
	if err := uc.consultationRepo.Create(ctx, log); err != nil {
		return nil, err
	}

	// Email notification (stub): student → supervisors, supervisor → student.
	if owner {
		emails := supervisorEmails(thesis)
		go func() {
			_ = uc.emailSvc.SendConsultationCreated(context.Background(), emails, log)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: userIDs(thesis.Supervisors),
				Title:   "Log Konsultasi Baru",
				Message: "Mahasiswa " + thesis.Student.FullName + " mencatat log konsultasi baru.",
				Type:    "consultation",
				Link:    notification.Path("/theses/%s/consultations", thesisID),
			})
		}()
	} else {
		go func() {
			_ = uc.emailSvc.SendConsultationCreated(context.Background(), []string{thesis.Student.Email}, log)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: []uuid.UUID{thesis.Student.ID},
				Title:   "Log Konsultasi Baru",
				Message: "Dosen pembimbing mencatat log konsultasi untuk skripsi Anda.",
				Type:    "consultation",
				Link:    notification.Path("/theses/%s/consultations", thesisID),
			})
		}()
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionConsultationCreated,
		EntityType: "consultation",
		EntityID:   &log.ID,
		NewValue:   consultationLogValue(log),
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})

	return toConsultationDetail(log), nil
}

// List returns paginated consultation logs plus a full summary (owner, supervisor, admin, kaprodi).
func (uc *ConsultationUseCase) List(ctx context.Context, thesisID uuid.UUID, filter domainRepo.ConsultationFilter, userID uuid.UUID, role string) (*ConsultationListResult, int64, error) {
	if err := uc.canView(ctx, thesisID, userID, role); err != nil {
		return nil, 0, err
	}

	logs, total, err := uc.consultationRepo.FindByThesisID(ctx, thesisID, filter)
	if err != nil {
		return nil, 0, err
	}
	summary, err := uc.summary(ctx, thesisID)
	if err != nil {
		return nil, 0, err
	}

	details := make([]*ConsultationDetail, 0, len(logs))
	for _, l := range logs {
		details = append(details, toConsultationDetail(l))
	}
	return &ConsultationListResult{Consultations: details, Summary: summary}, total, nil
}

// GetByID returns a single consultation log (same access as List).
func (uc *ConsultationUseCase) GetByID(ctx context.Context, thesisID, id, userID uuid.UUID, role string) (*ConsultationDetail, error) {
	if err := uc.canView(ctx, thesisID, userID, role); err != nil {
		return nil, err
	}
	log, err := uc.consultationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConsultationNotFound
		}
		return nil, err
	}
	if log.ThesisID != thesisID {
		return nil, ErrConsultationNotFound
	}
	return toConsultationDetail(log), nil
}

// Update edits a pending log; only the creator may update while it is pending.
func (uc *ConsultationUseCase) Update(ctx context.Context, thesisID, id uuid.UUID, req UpdateConsultationRequest, actor Actor) (*ConsultationDetail, error) {
	log, err := uc.consultationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConsultationNotFound
		}
		return nil, err
	}
	if log.ThesisID != thesisID {
		return nil, ErrConsultationNotFound
	}
	if log.CreatedBy != actor.UserID {
		return nil, ErrNotConsultationCreator
	}
	if log.Status != ConsultationStatusPending {
		return nil, ErrConsultationAlreadyDone
	}

	if req.ConsultationDate != nil {
		d, err := parseConsultationDate(*req.ConsultationDate)
		if err != nil {
			return nil, err
		}
		log.ConsultationDate = d
	}
	if req.TopicsDiscussed != nil {
		v := strings.TrimSpace(*req.TopicsDiscussed)
		if v == "" {
			return nil, ErrTopicsDiscussedRequired
		}
		log.TopicsDiscussed = v
	}
	if req.Notes != nil {
		log.Notes = req.Notes
	}
	if req.FollowUp != nil {
		log.FollowUp = req.FollowUp
	}
	if req.AttachmentURL != nil {
		log.AttachmentURL = req.AttachmentURL
	}

	if err := uc.consultationRepo.Update(ctx, log); err != nil {
		return nil, err
	}
	log.UpdatedAt = time.Now()

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor.UserID,
		Action:     audit.ActionConsultationUpdated,
		EntityType: "consultation",
		EntityID:   &log.ID,
		NewValue:   consultationLogValue(log),
		IPAddress:  actor.IPAddress,
		UserAgent:  actor.UserAgent,
	})
	return toConsultationDetail(log), nil
} // Approve approves a pending log; only a supervisor of the thesis may do so.
func (uc *ConsultationUseCase) Approve(ctx context.Context, thesisID, id, actor uuid.UUID, actorInfo Actor) (*ConsultationDetail, error) {
	log, err := uc.consultationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConsultationNotFound
		}
		return nil, err
	}
	if log.ThesisID != thesisID {
		return nil, ErrConsultationNotFound
	}
	if log.Status != ConsultationStatusPending {
		return nil, ErrConsultationAlreadyDone
	}

	ok, err := uc.access.IsSupervisor(ctx, actor, thesisID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotSupervisorOfThesis
	}

	if err := uc.consultationRepo.Approve(ctx, id, actor); err != nil {
		return nil, err
	}
	log.Status = ConsultationStatusApproved
	log.ApprovedBy = &actor
	now := time.Now()
	log.ApprovedAt = &now

	count, err := uc.consultationRepo.CountApprovedByThesisID(ctx, thesisID)
	if err != nil {
		return nil, err
	}

	thesis, err := uc.thesisRepo.FindByID(ctx, thesisID)
	if err == nil {
		studentEmail := thesis.Student.Email
		go func() {
			_ = uc.emailSvc.SendConsultationApproved(context.Background(), studentEmail, log)
			uc.notifSvc.Notify(notification.Params{
				UserIDs: []uuid.UUID{thesis.Student.ID},
				Title:   "Log Konsultasi Disetujui",
				Message: "Log konsultasi Anda telah disetujui oleh dosen pembimbing.",
				Type:    "consultation",
				Link:    notification.Path("/theses/%s/consultations", thesisID),
			})
		}()
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor,
		Action:     audit.ActionConsultationApproved,
		EntityType: "consultation",
		EntityID:   &log.ID,
		NewValue: map[string]interface{}{
			"status":         ConsultationStatusApproved,
			"approved_by":    actor,
			"approved_count": count,
		},
		IPAddress: actorInfo.IPAddress,
		UserAgent: actorInfo.UserAgent,
	})
	return toConsultationDetail(log), nil
}

// Delete removes a pending log; only the creator may delete while pending.
func (uc *ConsultationUseCase) Delete(ctx context.Context, thesisID, id, actor uuid.UUID, actorInfo Actor) error {
	log, err := uc.consultationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrConsultationNotFound
		}
		return err
	}
	if log.ThesisID != thesisID {
		return ErrConsultationNotFound
	}
	if log.CreatedBy != actor {
		return ErrNotConsultationCreator
	}
	if log.Status != ConsultationStatusPending {
		return ErrConsultationAlreadyDone
	}
	if err := uc.consultationRepo.Delete(ctx, id); err != nil {
		return err
	}

	uc.auditSvc.Log(ctx, audit.AuditParams{
		UserID:     &actor,
		Action:     audit.ActionConsultationDeleted,
		EntityType: "consultation",
		EntityID:   &id,
		IPAddress:  actorInfo.IPAddress,
		UserAgent:  actorInfo.UserAgent,
	})
	return nil
}

// Summary returns consultation statistics without pagination (same access as List).
func (uc *ConsultationUseCase) Summary(ctx context.Context, thesisID, userID uuid.UUID, role string) (*ConsultationSummary, error) {
	if err := uc.canView(ctx, thesisID, userID, role); err != nil {
		return nil, err
	}
	summary, err := uc.summary(ctx, thesisID)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// canView enforces the read access rule for consultation logs.
// Per job spec: Mahasiswa pemilik + Dosen Pembimbing + Admin + Kaprodi.
func (uc *ConsultationUseCase) canView(ctx context.Context, thesisID, userID uuid.UUID, role string) error {
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
	default:
		return ErrForbidden
	}
}

// summary aggregates statistics over ALL logs of a thesis (unpaginated).
func (uc *ConsultationUseCase) summary(ctx context.Context, thesisID uuid.UUID) (ConsultationSummary, error) {
	all, _, err := uc.consultationRepo.FindByThesisID(ctx, thesisID, domainRepo.ConsultationFilter{})
	if err != nil {
		return ConsultationSummary{}, err
	}

	s := ConsultationSummary{TotalConsultations: len(all)}
	now := time.Now()
	intervalDays := 0
	dates := make([]time.Time, 0, len(all))
	var lastDate time.Time

	for _, l := range all {
		if l.Status == ConsultationStatusApproved {
			s.ApprovedCount++
		} else {
			s.PendingCount++
		}
		if l.ConsultationDate.Year() == now.Year() && l.ConsultationDate.Month() == now.Month() {
			s.ConsultationsThisMonth++
		}
		if l.ConsultationDate.After(lastDate) {
			lastDate = l.ConsultationDate
		}
		dates = append(dates, l.ConsultationDate)
	}
	if !lastDate.IsZero() {
		v := lastDate.Format("2006-01-02")
		s.LastConsultationDate = &v
	}

	// Average gap between consecutive consultations (in days).
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	if len(dates) >= 2 {
		for i := 1; i < len(dates); i++ {
			intervalDays += int(dates[i].Sub(dates[i-1]).Hours() / 24)
		}
		avg := intervalDays / (len(dates) - 1)
		s.AverageIntervalDays = &avg
	}
	return s, nil
}

// isConsultationEligible reports whether the thesis status allows consultations.
func isConsultationEligible(status string) bool {
	switch status {
	case "in_progress", "seminar_ready", "seminar_done", "defense_ready", "defense_done", "graduated":
		return true
	default:
		return false
	}
}

// parseConsultationDate parses YYYY-MM-DD and rejects future dates.
func parseConsultationDate(s string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}
	today := time.Now().Format("2006-01-02")
	if s > today {
		return time.Time{}, ErrConsultationDateFuture
	}
	return date, nil
}

// supervisorEmails collects supervisor email addresses from the thesis.
func supervisorEmails(thesis *entity.Thesis) []string {
	emails := make([]string, 0, len(thesis.Supervisors))
	for _, s := range thesis.Supervisors {
		if s.Email != "" {
			emails = append(emails, s.Email)
		}
	}
	return emails
}

// consultationLogValue builds an audit-friendly snapshot of a log.
func consultationLogValue(l *entity.ConsultationLog) map[string]interface{} {
	return map[string]interface{}{
		"thesis_id":         l.ThesisID,
		"consultation_date": l.ConsultationDate.Format("2006-01-02"),
		"topics_discussed":  l.TopicsDiscussed,
		"status":            l.Status,
	}
}

// toConsultationDetail maps a persisted log into the API response shape.
func toConsultationDetail(l *entity.ConsultationLog) *ConsultationDetail {
	d := &ConsultationDetail{
		ID:               l.ID,
		ThesisID:         l.ThesisID,
		CreatedBy:        l.CreatedBy,
		ConsultationDate: l.ConsultationDate.Format("2006-01-02"),
		TopicsDiscussed:  l.TopicsDiscussed,
		Notes:            l.Notes,
		FollowUp:         l.FollowUp,
		AttachmentURL:    l.AttachmentURL,
		Status:           l.Status,
		ApprovedBy:       l.ApprovedBy,
		ApprovedAt:       l.ApprovedAt,
		CreatedAt:        l.CreatedAt,
		UpdatedAt:        l.UpdatedAt,
	}
	if l.Creator.ID != uuid.Nil {
		d.Creator = &UserBrief{ID: l.Creator.ID, FullName: l.Creator.FullName}
	}
	if l.Approver != nil && l.Approver.ID != uuid.Nil {
		d.Approver = &UserBrief{ID: l.Approver.ID, FullName: l.Approver.FullName}
	}
	return d
}
