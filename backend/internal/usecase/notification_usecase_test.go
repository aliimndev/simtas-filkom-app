package usecase

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

// fakeNotificationRepo is an in-memory NotificationRepository for usecase tests.
type fakeNotificationRepo struct {
	byUser map[uuid.UUID][]*entity.Notification
}

func newFakeNotificationRepo() *fakeNotificationRepo {
	return &fakeNotificationRepo{byUser: map[uuid.UUID][]*entity.Notification{}}
}

func (f *fakeNotificationRepo) CreateBatch(_ context.Context, userIDs []uuid.UUID, title, message, ntype string, link *string) error {
	for _, uid := range userIDs {
		f.byUser[uid] = append(f.byUser[uid], &entity.Notification{
			ID: uuid.New(), UserID: uid, Title: title, Message: message, Type: ntype, Link: link,
		})
	}
	return nil
}

func (f *fakeNotificationRepo) ListByUser(_ context.Context, userID uuid.UUID, limit int) ([]*entity.Notification, error) {
	all := f.byUser[userID]
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (f *fakeNotificationRepo) UnreadCount(_ context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	for _, notif := range f.byUser[userID] {
		if !notif.IsRead {
			n++
		}
	}
	return n, nil
}

func (f *fakeNotificationRepo) MarkRead(_ context.Context, userID, id uuid.UUID) error {
	for _, notif := range f.byUser[userID] {
		if notif.ID == id {
			notif.IsRead = true
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeNotificationRepo) MarkAllRead(_ context.Context, userID uuid.UUID) error {
	for _, notif := range f.byUser[userID] {
		notif.IsRead = true
	}
	return nil
}

var _ domainRepo.NotificationRepository = (*fakeNotificationRepo)(nil)

func newTestNotificationUseCase() (*NotificationUseCase, *fakeNotificationRepo) {
	repo := newFakeNotificationRepo()
	return NewNotificationUseCase(repo), repo
}

func TestNotificationListAndUnreadCount(t *testing.T) {
	uc, repo := newTestNotificationUseCase()
	userID := uuid.New()
	link := "/theses/abc"
	if err := repo.CreateBatch(context.Background(), []uuid.UUID{userID}, "Judul Baru", "Pesan", "thesis", &link); err != nil {
		t.Fatalf("seed: %v", err)
	}

	notifs, err := uc.List(context.Background(), userID, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(notifs) != 1 || notifs[0].Title != "Judul Baru" || notifs[0].Link == nil {
		t.Errorf("list = %+v", notifs)
	}

	count, err := uc.UnreadCount(context.Background(), userID)
	if err != nil || count != 1 {
		t.Errorf("UnreadCount = %d, %v; want 1", count, err)
	}
}

func TestNotificationMarkReadAndAll(t *testing.T) {
	uc, repo := newTestNotificationUseCase()
	userID := uuid.New()
	otherID := uuid.New()
	if err := repo.CreateBatch(context.Background(), []uuid.UUID{userID, otherID}, "T", "M", "thesis", nil); err != nil {
		t.Fatalf("seed: %v", err)
	}
	ids := make([]uuid.UUID, 0, len(repo.byUser[userID]))
	for _, n := range repo.byUser[userID] {
		ids = append(ids, n.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })

	if err := uc.MarkRead(context.Background(), userID, ids[0]); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	notifs, _ := repo.ListByUser(context.Background(), userID, 10)
	if !notifs[0].IsRead {
		t.Error("notification should be marked read")
	}

	if err := uc.MarkAllRead(context.Background(), otherID); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	other, _ := repo.ListByUser(context.Background(), otherID, 10)
	if !other[0].IsRead {
		t.Error("other user's notification should be read after MarkAllRead")
	}
}

func TestNotificationMarkReadNotFound(t *testing.T) {
	uc, _ := newTestNotificationUseCase()
	err := uc.MarkRead(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("expected ErrNotificationNotFound, got %v", err)
	}
}

func TestToNotificationDetail(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	link := "/theses/x"
	d := toNotificationDetail(&entity.Notification{
		ID: id, Title: "T", Message: "M", Type: "thesis", Link: &link, IsRead: false, CreatedAt: now,
	})
	if d.ID != id || d.Title != "T" || d.IsRead || d.Link == nil || d.CreatedAt.IsZero() {
		t.Errorf("detail = %+v", d)
	}
}
