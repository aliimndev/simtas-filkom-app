package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
)

// fakeUserRepo is a minimal in-memory UserRepository for usecase tests.
type fakeUserRepo struct {
	users       map[uuid.UUID]*entity.User
	emails      map[string]uuid.UUID
	roles       map[string]*entity.Role
	softDeleted map[uuid.UUID]bool
	activeMap   map[uuid.UUID]bool
	deleted     []uuid.UUID
	resetPw     []uuid.UUID
	invalidated []uuid.UUID
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:       map[uuid.UUID]*entity.User{},
		emails:      map[string]uuid.UUID{},
		roles:       map[string]*entity.Role{},
		softDeleted: map[uuid.UUID]bool{},
		activeMap:   map[uuid.UUID]bool{},
	}
}

func (f *fakeUserRepo) FindAll(_ context.Context, _ domainRepo.UserFilter) ([]*entity.User, int64, error) {
	users := make([]*entity.User, 0, len(f.users))
	for _, u := range f.users {
		users = append(users, u)
	}
	return users, int64(len(users)), nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	u, ok := f.users[id]
	if !ok || f.softDeleted[id] {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	id, ok := f.emails[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return f.users[id], nil
}

func (f *fakeUserRepo) FindByRole(_ context.Context, role string) ([]*entity.User, error) {
	var users []*entity.User
	for _, u := range f.users {
		if u.Role.Name == role && !f.softDeleted[u.ID] {
			users = append(users, u)
		}
	}
	return users, nil
}

func (f *fakeUserRepo) FindRoleByName(_ context.Context, name string) (*entity.Role, error) {
	r, ok := f.roles[name]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return r, nil
}

func (f *fakeUserRepo) Create(_ context.Context, user *entity.User) error {
	if _, exists := f.emails[user.Email]; exists {
		return errors.New("duplicate email")
	}
	user.ID = uuid.New()
	f.users[user.ID] = user
	f.emails[user.Email] = user.ID
	f.activeMap[user.ID] = user.IsActive
	return nil
}

func (f *fakeUserRepo) Update(_ context.Context, user *entity.User) error {
	f.users[user.ID] = user
	return nil
}

func (f *fakeUserRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	f.softDeleted[id] = true
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeUserRepo) BulkCreate(_ context.Context, users []*entity.User) error {
	for _, u := range users {
		_ = f.Create(context.Background(), u)
	}
	return nil
}

func (f *fakeUserRepo) SetActiveStatus(_ context.Context, id uuid.UUID, isActive bool) error {
	f.activeMap[id] = isActive
	return nil
}

func (f *fakeUserRepo) ResetPassword(_ context.Context, id uuid.UUID, _ string) error {
	f.resetPw = append(f.resetPw, id)
	return nil
}

func (f *fakeUserRepo) ChangePassword(_ context.Context, id uuid.UUID, _ string) error {
	f.resetPw = append(f.resetPw, id)
	return nil
}

func (f *fakeUserRepo) InvalidateUserSessions(_ context.Context, id uuid.UUID) error {
	f.invalidated = append(f.invalidated, id)
	return nil
}

type fakeEmailService struct{}

func (f *fakeEmailService) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (f *fakeEmailService) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
func (f *fakeEmailService) SendThesisSubmitted(context.Context, []string, *entity.Thesis) error {
	return nil
}
func (f *fakeEmailService) SendThesisApproved(context.Context, string, *entity.Thesis) error {
	return nil
}
func (f *fakeEmailService) SendThesisRejected(context.Context, string, *entity.Thesis, string) error {
	return nil
}
func (f *fakeEmailService) SendSupervisorAssigned(context.Context, string, []string, *entity.Thesis) error {
	return nil
}
func (f *fakeEmailService) SendConsultationCreated(context.Context, []string, *entity.ConsultationLog) error {
	return nil
}
func (f *fakeEmailService) SendConsultationApproved(context.Context, string, *entity.ConsultationLog) error {
	return nil
}
func (f *fakeEmailService) SendDocumentUploaded(context.Context, []string, *entity.Document) error {
	return nil
}
func (f *fakeEmailService) SendDocumentReviewed(context.Context, string, *entity.Document, string) error {
	return nil
}
func (f *fakeEmailService) SendSeminarSubmitted(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (f *fakeEmailService) SendSeminarScheduled(context.Context, []string, *entity.Seminar) error {
	return nil
}
func (f *fakeEmailService) SendSeminarFinalized(context.Context, string, *entity.Seminar) error {
	return nil
}
func (f *fakeEmailService) SendDefenseSubmitted(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (f *fakeEmailService) SendDefenseScheduled(context.Context, []string, *entity.ThesisDefense) error {
	return nil
}
func (f *fakeEmailService) SendDefenseFinalized(context.Context, string, *entity.ThesisDefense) error {
	return nil
}
func (f *fakeEmailService) SendGraduated(context.Context, string, *entity.Thesis) error {
	return nil
}
func (f *fakeEmailService) SendArchiveCreated(context.Context, string, *entity.ThesisArchive) error {
	return nil
}

func newTestUserUseCase() (*UserUseCase, *fakeUserRepo) {
	repo := newFakeUserRepo()
	repo.roles["mahasiswa"] = &entity.Role{ID: 3, Name: "mahasiswa"}
	repo.roles["admin_fakultas"] = &entity.Role{ID: 1, Name: "admin_fakultas"}

	auditSvc := audit.NewAuditService(nil) // nil repo → no-op, safe
	uc := NewUserUseCase(repo, &fakeEmailService{}, auditSvc)
	return uc, repo
}

func TestCreateUser(t *testing.T) {
	uc, repo := newTestUserUseCase()
	actor := Actor{UserID: uuid.New()}

	user, err := uc.Create(context.Background(), CreateUserRequest{
		Email:    "Mahasiswa@Example.com",
		FullName: "Nama Lengkap",
		Role:     "mahasiswa",
	}, actor)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if user.Email != "mahasiswa@example.com" {
		t.Errorf("email not normalized: %s", user.Email)
	}
	if user.PasswordHash == "" {
		t.Error("password hash is empty")
	}
	if user.PasswordHash == "mahasiswa@example.com" {
		t.Error("password stored as plaintext")
	}
	if !user.MustChangePassword {
		t.Error("must_change_password should be true for new users")
	}
	if user.RoleID != 3 {
		t.Errorf("role id = %d, want 3", user.RoleID)
	}
	if len(repo.users) != 1 {
		t.Errorf("expected 1 user created, got %d", len(repo.users))
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	uc, repo := newTestUserUseCase()
	_ = repo.Create(context.Background(), &entity.User{
		Email:    "existing@example.com",
		FullName: "Existing",
		RoleID:   3,
	})

	_, err := uc.Create(context.Background(), CreateUserRequest{
		Email:    "existing@example.com",
		FullName: "New User",
		Role:     "mahasiswa",
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestCreateUserInvalidRole(t *testing.T) {
	uc, _ := newTestUserUseCase()

	_, err := uc.Create(context.Background(), CreateUserRequest{
		Email:    "new@example.com",
		FullName: "New User",
		Role:     "tidak_ada",
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrRoleInvalid) {
		t.Errorf("expected ErrRoleInvalid, got %v", err)
	}
}

func TestDeleteSelfRejected(t *testing.T) {
	uc, _ := newTestUserUseCase()
	adminID := uuid.New()

	err := uc.Delete(context.Background(), adminID, Actor{UserID: adminID})
	if !errors.Is(err, ErrCannotDeleteSelf) {
		t.Errorf("expected ErrCannotDeleteSelf, got %v", err)
	}
}

func TestDeactivateSelfRejected(t *testing.T) {
	uc, _ := newTestUserUseCase()
	adminID := uuid.New()

	err := uc.Deactivate(context.Background(), adminID, Actor{UserID: adminID})
	if !errors.Is(err, ErrCannotDeactivateSelf) {
		t.Errorf("expected ErrCannotDeactivateSelf, got %v", err)
	}
}

func TestDeactivateInvalidatesSessions(t *testing.T) {
	uc, repo := newTestUserUseCase()
	target := &entity.User{Email: "target@example.com", FullName: "Target", RoleID: 3}
	_ = repo.Create(context.Background(), target)

	actor := Actor{UserID: uuid.New()}
	if err := uc.Deactivate(context.Background(), target.ID, actor); err != nil {
		t.Fatalf("Deactivate returned error: %v", err)
	}

	if len(repo.invalidated) != 1 || repo.invalidated[0] != target.ID {
		t.Errorf("expected sessions invalidated for target user")
	}
}

func TestResetPassword(t *testing.T) {
	uc, repo := newTestUserUseCase()
	target := &entity.User{Email: "target@example.com", FullName: "Target", RoleID: 3}
	_ = repo.Create(context.Background(), target)

	if err := uc.ResetPassword(context.Background(), target.ID, Actor{UserID: uuid.New()}); err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if len(repo.resetPw) != 1 || repo.resetPw[0] != target.ID {
		t.Errorf("expected password reset for target user")
	}
}

func TestChangeMyPassword(t *testing.T) {
	uc, repo := newTestUserUseCase()

	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), 12)
	target := &entity.User{Email: "target@example.com", FullName: "Target", RoleID: 3, PasswordHash: string(hash)}
	_ = repo.Create(context.Background(), target)

	err := uc.ChangeMyPassword(context.Background(), target.ID, ChangePasswordRequest{
		CurrentPassword: "OldPass123",
		NewPassword:     "NewPass456",
	}, Actor{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("ChangeMyPassword returned error: %v", err)
	}
	if len(repo.resetPw) != 1 || repo.resetPw[0] != target.ID {
		t.Errorf("expected password change + session invalidation for target user")
	}
}

func TestChangeMyPasswordWrongCurrent(t *testing.T) {
	uc, repo := newTestUserUseCase()

	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), 12)
	target := &entity.User{Email: "target@example.com", FullName: "Target", RoleID: 3, PasswordHash: string(hash)}
	_ = repo.Create(context.Background(), target)

	err := uc.ChangeMyPassword(context.Background(), target.ID, ChangePasswordRequest{
		CurrentPassword: "WrongPass",
		NewPassword:     "NewPass456",
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestChangeMyPasswordNotComplex(t *testing.T) {
	uc, repo := newTestUserUseCase()

	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), 12)
	target := &entity.User{Email: "target@example.com", FullName: "Target", RoleID: 3, PasswordHash: string(hash)}
	_ = repo.Create(context.Background(), target)

	err := uc.ChangeMyPassword(context.Background(), target.ID, ChangePasswordRequest{
		CurrentPassword: "OldPass123",
		NewPassword:     "alllowercase",
	}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrPasswordNotComplex) {
		t.Errorf("expected ErrPasswordNotComplex, got %v", err)
	}
}
