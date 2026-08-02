package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
)

func validYearRequest() AcademicYearRequest {
	return AcademicYearRequest{
		Name:      "2026/2027",
		Semester:  "ganjil",
		StartDate: "2026-09-01",
		EndDate:   "2027-01-31",
	}
}

func TestBuildAcademicYear(t *testing.T) {
	y, err := buildAcademicYear(validYearRequest())
	if err != nil {
		t.Fatalf("buildAcademicYear: %v", err)
	}
	if y.Name != "2026/2027" || y.Semester != "ganjil" {
		t.Errorf("unexpected year: %+v", y)
	}
}

func TestBuildAcademicYearInvalidSemester(t *testing.T) {
	req := validYearRequest()
	req.Semester = "musim"
	_, err := buildAcademicYear(req)
	if !errors.Is(err, ErrInvalidSemester) {
		t.Errorf("expected ErrInvalidSemester, got %v", err)
	}
}

func TestBuildAcademicYearInvalidDates(t *testing.T) {
	tests := []struct {
		name string
		req  AcademicYearRequest
	}{
		{"bad start", func() AcademicYearRequest { r := validYearRequest(); r.StartDate = "01-09-2026"; return r }()},
		{"bad end", func() AcademicYearRequest { r := validYearRequest(); r.EndDate = "not-a-date"; return r }()},
		{"end before start", func() AcademicYearRequest { r := validYearRequest(); r.EndDate = "2026-08-01"; return r }()},
		{"same day", func() AcademicYearRequest { r := validYearRequest(); r.EndDate = "2026-09-01"; return r }()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildAcademicYear(tt.req); err == nil {
				t.Fatal("expected error for invalid date range/format")
			}
		})
	}
}

func TestCreateAcademicYear(t *testing.T) {
	uc := NewAcademicYearUseCase(newFakeAcademicYearRepo())
	y, err := uc.Create(context.Background(), validYearRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if y.ID == uuid.Nil {
		t.Error("expected generated ID")
	}
}

func TestCreateAcademicYearInvalid(t *testing.T) {
	uc := NewAcademicYearUseCase(newFakeAcademicYearRepo())
	req := validYearRequest()
	req.StartDate = "2027-02-01"
	req.EndDate = "2027-01-01"
	_, err := uc.Create(context.Background(), req)
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Errorf("expected ErrInvalidDateRange, got %v", err)
	}
}

func TestUpdateAcademicYear(t *testing.T) {
	repo := newFakeAcademicYearRepo()
	uc := NewAcademicYearUseCase(repo)
	existing := &entity.AcademicYear{Name: "2025/2026", Semester: "ganjil", IsActive: false}
	_ = repo.Create(context.Background(), existing)

	updated, err := uc.Update(context.Background(), existing.ID, validYearRequest())
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "2026/2027" {
		t.Errorf("name = %q, want 2026/2027", updated.Name)
	}
	if updated.ID != existing.ID {
		t.Error("ID must be preserved on update")
	}
}

func TestUpdateAcademicYearNotFound(t *testing.T) {
	uc := NewAcademicYearUseCase(newFakeAcademicYearRepo())
	_, err := uc.Update(context.Background(), uuid.New(), validYearRequest())
	if !errors.Is(err, ErrAcademicYearNotFound) {
		t.Errorf("expected ErrAcademicYearNotFound, got %v", err)
	}
}

func TestActivateAcademicYear(t *testing.T) {
	repo := newFakeAcademicYearRepo()
	uc := NewAcademicYearUseCase(repo)
	a := &entity.AcademicYear{Name: "2026/2027", Semester: "ganjil"}
	b := &entity.AcademicYear{Name: "2027/2028", Semester: "genap"}
	_ = repo.Create(context.Background(), a)
	_ = repo.Create(context.Background(), b)

	if err := uc.Activate(context.Background(), b.ID); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if repo.active != b.ID {
		t.Errorf("active year = %v, want %v (only one active at a time)", repo.active, b.ID)
	}
}

func TestActivateAcademicYearNotFound(t *testing.T) {
	uc := NewAcademicYearUseCase(newFakeAcademicYearRepo())
	err := uc.Activate(context.Background(), uuid.New())
	if !errors.Is(err, ErrAcademicYearNotFound) {
		t.Errorf("expected ErrAcademicYearNotFound, got %v", err)
	}
}
