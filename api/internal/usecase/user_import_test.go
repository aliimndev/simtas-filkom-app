package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/aliimndev/simtas-filkom-app/api/internal/domain/entity"
)

func TestParseCSV(t *testing.T) {
	csv := "email,full_name,nim_nidn,role,study_program\n" +
		"ali@example.com,Ali,NIM001,mahasiswa,Teknik Informatika\n" +
		"budi@example.com,Budi,,dosen_pembimbing,\n"

	rows, err := parseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("parseCSV returned error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Row numbers are 1-based file rows; header is row 1.
	if rows[0].Row != 2 {
		t.Errorf("rows[0].Row = %d, want 2", rows[0].Row)
	}
	if rows[0].Email != "ali@example.com" {
		t.Errorf("rows[0].Email = %q", rows[0].Email)
	}
	if rows[1].Role != "dosen_pembimbing" {
		t.Errorf("rows[1].Role = %q", rows[1].Role)
	}
}

func TestParseCSVHeaderOnly(t *testing.T) {
	rows, err := parseCSV([]byte("email,full_name,nim_nidn,role,study_program\n"))
	if err != nil {
		t.Fatalf("parseCSV returned error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestImportUsersCSV(t *testing.T) {
	uc, repo := newTestUserUseCase()

	csv := "email,full_name,nim_nidn,role,study_program\n" +
		"ali@example.com,Ali,NIM001,mahasiswa,Teknik Informatika\n" +
		"budi@example.com,Budi,,mahasiswa,Teknik Informatika\n" +
		"invalid-email,Nama Salah,,mahasiswa,\n"

	result, err := uc.ImportUsers(context.Background(), "users.csv", []byte(csv), Actor{})
	if err != nil {
		t.Fatalf("ImportUsers returned error: %v", err)
	}

	if result.TotalRows != 3 {
		t.Errorf("TotalRows = %d, want 3", result.TotalRows)
	}
	if result.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", result.SuccessCount)
	}
	if result.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", result.ErrorCount)
	}
	if len(result.Errors) != 1 || result.Errors[0].Row != 4 {
		t.Errorf("errors = %+v, want 1 error at row 4", result.Errors)
	}
	if result.Errors[0].Reason != "Format email tidak valid" {
		t.Errorf("reason = %q, want 'Format email tidak valid'", result.Errors[0].Reason)
	}
	if len(repo.users) != 2 {
		t.Errorf("expected 2 users created, got %d", len(repo.users))
	}
}

func TestImportUsersDuplicateEmailInDB(t *testing.T) {
	uc, repo := newTestUserUseCase()
	// Pre-register one email in the DB (as if it existed from a previous import)
	_ = repo.Create(context.Background(), &entity.User{
		Email:    "existing@example.com",
		FullName: "Existing User",
		RoleID:   3,
	})

	csv := "email,full_name,nim_nidn,role,study_program\n" +
		"new@example.com,New User,,mahasiswa,\n" +
		"existing@example.com,Existing Duplicate,,mahasiswa,\n"

	result, err := uc.ImportUsers(context.Background(), "users.csv", []byte(csv), Actor{})
	if err != nil {
		t.Fatalf("ImportUsers returned error: %v", err)
	}

	if result.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", result.SuccessCount)
	}
	if result.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", result.ErrorCount)
	}
	if len(result.Errors) != 1 || result.Errors[0].Row != 3 {
		t.Errorf("errors = %+v, want 1 error at row 3", result.Errors)
	}
	if result.Errors[0].Reason != "Email sudah terdaftar" {
		t.Errorf("reason = %q, want 'Email sudah terdaftar'", result.Errors[0].Reason)
	}
	// Only the genuinely new user should have been created
	if len(repo.users) != 2 { // 1 pre-seeded + 1 new
		t.Errorf("expected 2 users total, got %d", len(repo.users))
	}
}

func TestImportUsersInvalidFileFormat(t *testing.T) {
	uc, _ := newTestUserUseCase()

	_, err := uc.ImportUsers(context.Background(), "users.txt", []byte("x"), Actor{})
	if !errors.Is(err, ErrInvalidFileFormat) {
		t.Errorf("expected ErrInvalidFileFormat, got %v", err)
	}
}

func TestImportUsersFileTooLarge(t *testing.T) {
	uc, _ := newTestUserUseCase()

	big := make([]byte, MaxImportFileSize+1)
	_, err := uc.ImportUsers(context.Background(), "users.csv", big, Actor{})
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}
