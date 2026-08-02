package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

// seedUser adds a plain user to the fake repo and returns it.
func seedUser(repo *fakeUserRepo, email string) *entity.User {
	u := &entity.User{Email: email, FullName: "Target User", RoleID: 3, IsActive: true}
	_ = repo.Create(context.Background(), u)
	return u
}

func TestUpdateUser(t *testing.T) {
	uc, repo := newTestUserUseCase()
	target := seedUser(repo, "target@example.com")
	actor := Actor{UserID: uuid.New()}

	fullName := "Nama Baru"
	program := "Teknik Informatika"
	updated, err := uc.Update(context.Background(), target.ID, UpdateUserRequest{
		FullName:     &fullName,
		StudyProgram: &program,
	}, actor)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.FullName != fullName {
		t.Errorf("full_name = %q, want %q", updated.FullName, fullName)
	}
	if updated.StudyProgram == nil || *updated.StudyProgram != program {
		t.Errorf("study_program = %v, want %q", updated.StudyProgram, program)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	uc, _ := newTestUserUseCase()
	name := "X"
	_, err := uc.Update(context.Background(), uuid.New(), UpdateUserRequest{FullName: &name}, Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	uc, repo := newTestUserUseCase()
	target := seedUser(repo, "target@example.com")
	actor := Actor{UserID: uuid.New()}

	if err := uc.Delete(context.Background(), target.ID, actor); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != target.ID {
		t.Errorf("expected soft delete for target user")
	}
	// Soft-deleted users are no longer found.
	if _, err := repo.FindByID(context.Background(), target.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected user to be soft-deleted, got err=%v", err)
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	uc, _ := newTestUserUseCase()
	err := uc.Delete(context.Background(), uuid.New(), Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestActivateUser(t *testing.T) {
	uc, repo := newTestUserUseCase()
	target := seedUser(repo, "target@example.com")
	target.IsActive = false
	repo.activeMap[target.ID] = false
	actor := Actor{UserID: uuid.New()}

	if err := uc.Activate(context.Background(), target.ID, actor); err != nil {
		t.Fatalf("Activate returned error: %v", err)
	}
	if !repo.activeMap[target.ID] {
		t.Error("expected user to be activated")
	}
}

func TestActivateUserNotFound(t *testing.T) {
	uc, _ := newTestUserUseCase()
	err := uc.Activate(context.Background(), uuid.New(), Actor{UserID: uuid.New()})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
