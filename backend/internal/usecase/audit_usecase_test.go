package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

// fakeAuditRepo is an in-memory AuditRepository for usecase tests.
type fakeAuditRepo struct {
	logs []*entity.AuditLog
}

var _ domainRepo.AuditRepository = (*fakeAuditRepo)(nil)

func (f *fakeAuditRepo) Create(_ context.Context, log *entity.AuditLog) error {
	log.ID = uuid.New()
	f.logs = append(f.logs, log)
	return nil
}

func (f *fakeAuditRepo) FindAll(_ context.Context, filter domainRepo.AuditFilter) ([]*entity.AuditLog, int64, error) {
	var matched []*entity.AuditLog
	for _, l := range f.logs {
		if filter.UserID != nil && (l.UserID == nil || *l.UserID != *filter.UserID) {
			continue
		}
		if filter.Action != "" && l.Action != filter.Action {
			continue
		}
		if filter.EntityType != "" && (l.EntityType == nil || *l.EntityType != filter.EntityType) {
			continue
		}
		if filter.EntityID != nil && (l.EntityID == nil || *l.EntityID != *filter.EntityID) {
			continue
		}
		if filter.DateFrom != nil && l.CreatedAt.Before(*filter.DateFrom) {
			continue
		}
		if filter.DateTo != nil && l.CreatedAt.After(*filter.DateTo) {
			continue
		}
		clone := *l
		matched = append(matched, &clone)
	}

	// Mirror the real repository: total counts every match; only the page
	// slice is returned.
	total := len(matched)
	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}
	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return matched[start:end], int64(total), nil
}

func (f *fakeAuditRepo) FindByEntity(_ context.Context, entityType string, entityID uuid.UUID) ([]*entity.AuditLog, error) {
	var out []*entity.AuditLog
	for _, l := range f.logs {
		if l.EntityType != nil && *l.EntityType == entityType && l.EntityID != nil && *l.EntityID == entityID {
			clone := *l
			out = append(out, &clone)
		}
	}
	// Mirrors the real repository: no history returns an empty slice, not an error.
	return out, nil
}

func seedAuditLog(uid uuid.UUID, action, entityType string, eid uuid.UUID) *entity.AuditLog {
	ts := time.Now()
	return &entity.AuditLog{
		UserID:     &uid,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &eid,
		CreatedAt:  ts,
	}
}

func TestAuditList(t *testing.T) {
	repo := &fakeAuditRepo{}
	uc := NewAuditUseCase(repo)

	userID := uuid.New()
	eid := uuid.New()
	repo.logs = append(repo.logs,
		seedAuditLog(userID, "THESIS_SUBMITTED", "thesis", eid),
		seedAuditLog(userID, "THESIS_APPROVED", "thesis", eid),
		seedAuditLog(uuid.New(), "USER_LOGIN", "user", userID),
	)

	// All logs
	details, total, err := uc.List(context.Background(), domainRepo.AuditFilter{Page: 1, PerPage: 50})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 || len(details) != 3 {
		t.Errorf("all logs: total=%d len=%d, want 3/3", total, len(details))
	}

	// Filter by action
	details, total, err = uc.List(context.Background(), domainRepo.AuditFilter{Action: "USER_LOGIN", Page: 1, PerPage: 50})
	if err != nil {
		t.Fatalf("List action filter: %v", err)
	}
	if total != 1 || len(details) != 1 || details[0].Action != "USER_LOGIN" {
		t.Errorf("action filter: total=%d len=%d, want 1/1 USER_LOGIN", total, len(details))
	}

	// Filter by user
	details, total, err = uc.List(context.Background(), domainRepo.AuditFilter{UserID: &userID, Page: 1, PerPage: 50})
	if err != nil {
		t.Fatalf("List user filter: %v", err)
	}
	if total != 2 || len(details) != 2 {
		t.Errorf("user filter: total=%d len=%d, want 2/2", total, len(details))
	}

	// Pagination: page 2 with per_page=2 returns the remaining log
	details, total, err = uc.List(context.Background(), domainRepo.AuditFilter{Page: 2, PerPage: 2})
	if err != nil {
		t.Fatalf("List pagination: %v", err)
	}
	if total != 3 || len(details) != 1 {
		t.Errorf("pagination: total=%d len=%d, want 3 total / 1 on page 2", total, len(details))
	}
}

func TestAuditByEntity(t *testing.T) {
	repo := &fakeAuditRepo{}
	uc := NewAuditUseCase(repo)

	userID := uuid.New()
	eid := uuid.New()
	repo.logs = append(repo.logs,
		seedAuditLog(userID, "THESIS_SUBMITTED", "thesis", eid),
		seedAuditLog(userID, "THESIS_APPROVED", "thesis", eid),
		seedAuditLog(userID, "USER_LOGIN", "user", userID), // different entity
	)

	details, err := uc.ByEntity(context.Background(), "thesis", eid)
	if err != nil {
		t.Fatalf("ByEntity: %v", err)
	}
	if len(details) != 2 {
		t.Errorf("ByEntity len=%d, want 2", len(details))
	}
	if details[0].Action != "THESIS_SUBMITTED" {
		t.Errorf("first action = %q, want THESIS_SUBMITTED", details[0].Action)
	}
}

func TestAuditByEntityEmpty(t *testing.T) {
	repo := &fakeAuditRepo{}
	uc := NewAuditUseCase(repo)
	details, err := uc.ByEntity(context.Background(), "thesis", uuid.New())
	if err != nil {
		t.Fatalf("ByEntity on empty history should not error: %v", err)
	}
	if len(details) != 0 {
		t.Errorf("ByEntity len=%d, want 0", len(details))
	}
}

func TestToAuditLogDetailEntityID(t *testing.T) {
	userID := uuid.New()
	eid := uuid.New()
	l := seedAuditLog(userID, "THESIS_APPROVED", "thesis", eid)
	l.User = &entity.User{ID: userID, FullName: "Budi", Email: "budi@example.com"}

	d := toAuditLogDetail(l)
	if d.EntityID == nil || *d.EntityID != eid {
		t.Errorf("entity_id not mapped correctly: %v", d.EntityID)
	}
	if d.User == nil || d.User.FullName != "Budi" {
		t.Errorf("user brief not mapped correctly: %+v", d.User)
	}
}
