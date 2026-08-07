//go:build integration

package handler_test

import (
	"context"
	"testing"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTitleChangeRequestSoftDeleteColumn — regression guard for the
// entity/migration mismatch that broke every GORM query on
// title_change_requests with `column "deleted_at" does not exist` (and the
// actor columns requested_by/reviewed_by/cancelled_by not matching GORM's
// *_id derivation).
//
// entity.TitleChangeRequest declares gorm.DeletedAt, so GORM appends
// `deleted_at IS NULL` to every read; 000014 omitted the column and used the
// wrong actor-column names. Migration 000018 renames the columns to *_id and
// adds deleted_at. Without 000018 this test fails on Create/FindByThesisID;
// with it the round-trip succeeds — proving the migrated schema matches the
// entity. Lives in the handler package (not repository) because it is the
// single package that owns the shared integration test DB; two packages using
// it in parallel race each other's truncate/seed.
func TestTitleChangeRequestSoftDeleteColumn(t *testing.T) {
	db := testutil.SetupTestDB(t)

	student := testutil.SeedUser(t, db, "tc-student@filkom.unida.ac.id", "StudentPass1", "TC Student", "mahasiswa")
	academicYearID := testutil.SeedActiveAcademicYear(t, db)

	thesis := entity.Thesis{
		StudentID:      student.ID,
		AcademicYearID: academicYearID,
		Title:          "Implementasi Sistem Informasi Tugas Akhir",
		ThesisType:     "skripsi",
		Status:         "approved",
	}
	require.NoError(t, db.Create(&thesis).Error)

	repo := repository.NewTitleChangeRequestRepository(db)

	req := entity.TitleChangeRequest{
		ThesisID:       thesis.ID,
		RequestedByID:  student.ID,
		PreviousTitle:  thesis.Title,
		RequestedTitle: "Optimasi Sistem Informasi Tugas Akhir Berbasis Web",
		Reason:         stringPtr("Penyempurnaan judul"),
	}
	require.NoError(t, repo.Create(context.Background(), &req))

	got, err := repo.FindByThesisID(context.Background(), thesis.ID)
	require.NoError(t, err, "FindByThesisID must not fail with a deleted_at column error")
	require.Len(t, got, 1)
	assert.Equal(t, req.RequestedTitle, got[0].RequestedTitle)
	assert.Equal(t, "PENDING", got[0].Status)
}

func stringPtr(s string) *string {
	return &s
}
