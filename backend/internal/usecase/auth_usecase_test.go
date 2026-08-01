package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

func TestIsPasswordComplex(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"uppercase and digit", "Password1", true},
		{"uppercase only", "PASSWORD", false},
		{"digit only", "password1", false},
		{"neither", "password", false},
		{"digit at start uppercase at end", "1passwordA", true},
		{"unicode uppercase treated as upper", "Pässwörd1", true},
		{"symbols no digit", "P@ssword!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPasswordComplex(tt.password); got != tt.want {
				t.Errorf("isPasswordComplex(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

type fakeAuthRepo struct{}

func (f *fakeAuthRepo) FindUserByEmail(_ context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (f *fakeAuthRepo) FindUserByID(_ context.Context, _ uuid.UUID) (*entity.User, error) {
	return nil, nil
}
func (f *fakeAuthRepo) UpdateLoginAttempt(_ context.Context, _ uuid.UUID, _ int, _ *time.Time) error {
	return nil
}
func (f *fakeAuthRepo) UpdateLastLogin(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeAuthRepo) BlacklistToken(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (f *fakeAuthRepo) IsTokenBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (f *fakeAuthRepo) CreatePasswordResetToken(_ context.Context, _ *entity.PasswordResetToken) error {
	return nil
}
func (f *fakeAuthRepo) FindPasswordResetTokenByToken(_ context.Context, _ string) (*entity.PasswordResetToken, error) {
	return nil, nil
}
func (f *fakeAuthRepo) MarkPasswordResetTokenUsed(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (f *fakeAuthRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (f *fakeAuthRepo) ClearMustChangePassword(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (f *fakeAuthRepo) GetUserTokenVersion(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func TestResetPasswordValidation(t *testing.T) {
	uc := NewAuthUseCase(&fakeAuthRepo{}, nil)

	tests := []struct {
		name string
		req  ResetPasswordRequest
		want error
	}{
		{"password mismatch", ResetPasswordRequest{Token: "t", NewPassword: "StrongPass1", ConfirmPassword: "Different1"}, ErrPasswordMismatch},
		{"too short", ResetPasswordRequest{Token: "t", NewPassword: "Short1", ConfirmPassword: "Short1"}, ErrPasswordTooShort},
		{"missing uppercase", ResetPasswordRequest{Token: "t", NewPassword: "weakpass1", ConfirmPassword: "weakpass1"}, ErrPasswordNotComplex},
		{"missing digit", ResetPasswordRequest{Token: "t", NewPassword: "WeakPassw", ConfirmPassword: "WeakPassw"}, ErrPasswordNotComplex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.ResetPassword(context.Background(), tt.req)
			if err != tt.want {
				t.Errorf("ResetPassword error = %v, want %v", err, tt.want)
			}
		})
	}
}
