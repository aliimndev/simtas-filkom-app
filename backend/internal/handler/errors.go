package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

// respondError is the single source of truth mapping use case sentinel errors to
// HTTP responses (Refactor 2: central error catalog). Per-handler helpers
// delegate to it so the status-code mapping is defined once and every endpoint
// emits the standardized APIResponse shape (Refactor 3).
//
// Order matters: ErrThesisNotFound and ErrForbidden are intentionally checked
// before domain-specific errors because they are shared by both thesis and
// defense flows.
func respondError(c *gin.Context, err error) {
	switch {
	// ── 404 Not Found ─────────────────────────────────────────────────────
	case errors.Is(err, usecase.ErrThesisNotFound):
		response.NotFound(c, "Thesis tidak ditemukan")
	case errors.Is(err, usecase.ErrDefenseNotFound):
		response.NotFound(c, "Sidang tidak ditemukan")

	// ── 403 Forbidden ────────────────────────────────────────────────────
	case errors.Is(err, usecase.ErrForbidden),
		errors.Is(err, usecase.ErrDefenseNotExaminer):
		response.Forbidden(c, "Akses ditolak")

	// ── 409 Conflict ─────────────────────────────────────────────────────
	case errors.Is(err, usecase.ErrDefenseActiveExists),
		errors.Is(err, usecase.ErrDefenseRoomConflict),
		errors.Is(err, usecase.ErrDefenseAlreadyScored):
		response.Error(c, http.StatusConflict, err.Error(), err)

	// ── 422 Unprocessable Entity (state / gate violations) ───────────────
	case errors.Is(err, usecase.ErrInvalidStateTransition),
		errors.Is(err, usecase.ErrThesisCannotCancel),
		errors.Is(err, usecase.ErrDefenseGateNotMet),
		errors.Is(err, usecase.ErrGraduationGateNotMet),
		errors.Is(err, usecase.ErrDefenseNotScheduled),
		errors.Is(err, usecase.ErrDefenseNotFinalized):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), err)

	// ── 400 Bad Request (validation) ─────────────────────────────────────
	case errors.Is(err, usecase.ErrActiveThesisExists),
		errors.Is(err, usecase.ErrTitleTooShort),
		errors.Is(err, usecase.ErrTitleTooLong),
		errors.Is(err, usecase.ErrAbstractTooShort),
		errors.Is(err, usecase.ErrInvalidThesisType),
		errors.Is(err, usecase.ErrInvalidDecision),
		errors.Is(err, usecase.ErrInvalidSupervisorCount),
		errors.Is(err, usecase.ErrSupervisorNotEligible),
		errors.Is(err, usecase.ErrThesisAlreadyCancelled),
		errors.Is(err, usecase.ErrNoActiveAcademicYear),
		errors.Is(err, usecase.ErrDefenseScheduleLeadTime),
		errors.Is(err, usecase.ErrDefenseMinExaminers),
		errors.Is(err, usecase.ErrDefenseInvalidExaminer),
		errors.Is(err, usecase.ErrDefenseIncompleteScore),
		errors.Is(err, usecase.ErrDefenseInvalidScore):
		response.BadRequest(c, err.Error())

	default:
		response.InternalError(c, "Terjadi kesalahan server")
	}
}
