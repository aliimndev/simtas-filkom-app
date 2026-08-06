package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
	"gorm.io/gorm"
)

// rotationAuthRepo is an in-memory AuthRepository that also implements the
// refresh-token family API, simulating the row-locked rotation semantics of
// the real repository.
type rotationAuthRepo struct {
	*authUserRepo
	families map[string]*entity.RefreshTokenFamily // token_jti -> family
	user     *entity.User
}

func newRotationAuthRepo(user *entity.User) *rotationAuthRepo {
	return &rotationAuthRepo{
		authUserRepo: &authUserRepo{user: user},
		families:     map[string]*entity.RefreshTokenFamily{},
		user:         user,
	}
}

func (f *rotationAuthRepo) CreateRefreshTokenFamily(_ context.Context, family *entity.RefreshTokenFamily) error {
	if f.user != nil {
		family.UserID = f.user.ID
	}
	f.families[family.TokenJTI] = family
	return nil
}

func (f *rotationAuthRepo) FindRefreshTokenFamilyByJTI(_ context.Context, jti string) (*entity.RefreshTokenFamily, error) {
	fam, ok := f.families[jti]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return fam, nil
}

func (f *rotationAuthRepo) RotateRefreshTokenFamily(_ context.Context, oldJTI, newJTI string, newExpiresAt time.Time) (bool, error) {
	fam, ok := f.families[oldJTI]
	if !ok {
		return false, nil
	}
	delete(f.families, oldJTI)
	fam.TokenJTI = newJTI
	fam.ExpiresAt = newExpiresAt
	f.families[newJTI] = fam
	return true, nil
}

func (f *rotationAuthRepo) RevokeRefreshTokenFamiliesByUser(_ context.Context, userID uuid.UUID) error {
	if f.user != nil && f.user.ID != userID {
		return nil
	}
	f.families = map[string]*entity.RefreshTokenFamily{}
	return nil
}

func newRotationAuthUC(user *entity.User) (*AuthUseCase, *jwt.JWTManager, *rotationAuthRepo) {
	repo := newRotationAuthRepo(user)
	jwtMgr := jwt.NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	auditRepo := &chanAuditRepo{actions: make(chan string, 4)}
	uc := NewAuthUseCase(repo, jwtMgr, audit.NewAuditService(auditRepo))
	return uc, jwtMgr, repo
}

// TestRefreshTokenRotation — a successful refresh must return a NEW refresh
// token and rotate the family so the old token can no longer be used.
func TestRefreshTokenRotation(t *testing.T) {
	user := newTestUser("Password123")
	uc, jwtMgr, repo := newRotationAuthUC(user)

	// Seed a family the way Login would: one current JTI.
	token, jti, err := jwtMgr.GenerateRefreshToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if err := repo.CreateRefreshTokenFamily(context.Background(), &entity.RefreshTokenFamily{
		UserID:    user.ID,
		FamilyID:  uuid.New(),
		TokenJTI:  jti,
		ExpiresAt: time.Now().Add(jwtMgr.RefreshTokenExpiry()),
	}); err != nil {
		t.Fatalf("seed family: %v", err)
	}

	resp, err := uc.RefreshToken(context.Background(), token)
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected a new access token")
	}
	if resp.RefreshToken == "" {
		t.Fatal("expected a rotated refresh token")
	}
	if resp.RefreshToken == token {
		t.Error("rotated refresh token must differ from the presented token")
	}

	// The old JTI must no longer resolve to a family (it was rotated away).
	if _, err := repo.FindRefreshTokenFamilyByJTI(context.Background(), jti); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("old JTI should be rotated out of the family, got err=%v", err)
	}
}

// TestRefreshTokenReuseDetection — presenting an already-rotated token must be
// treated as theft: the whole family is revoked and the refresh fails.
func TestRefreshTokenReuseDetection(t *testing.T) {
	user := newTestUser("Password123")
	uc, jwtMgr, repo := newRotationAuthUC(user)

	oldToken, oldJTI, err := jwtMgr.GenerateRefreshToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if err := repo.CreateRefreshTokenFamily(context.Background(), &entity.RefreshTokenFamily{
		UserID:    user.ID,
		FamilyID:  uuid.New(),
		TokenJTI:  oldJTI,
		ExpiresAt: time.Now().Add(jwtMgr.RefreshTokenExpiry()),
	}); err != nil {
		t.Fatalf("seed family: %v", err)
	}

	// First refresh succeeds and rotates.
	if _, err := uc.RefreshToken(context.Background(), oldToken); err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	// Replaying the OLD token is detected as reuse and must fail.
	if _, err := uc.RefreshToken(context.Background(), oldToken); !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Errorf("replayed token should fail with ErrRefreshTokenInvalid, got %v", err)
	}

	// The whole family is revoked: no families remain for the user.
	if len(repo.families) != 0 {
		t.Errorf("expected all families revoked after reuse, got %d remaining", len(repo.families))
	}
}

// TestRefreshTokenMissingFamily — a validly-signed token with no family row
// (never issued, or family already expired) must fail closed.
func TestRefreshTokenMissingFamily(t *testing.T) {
	user := newTestUser("Password123")
	uc, jwtMgr, _ := newRotationAuthUC(user)

	token, _, err := jwtMgr.GenerateRefreshToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	// No family row created for this token.
	if _, err := uc.RefreshToken(context.Background(), token); !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Errorf("expected ErrRefreshTokenInvalid for orphan token, got %v", err)
	}
}
