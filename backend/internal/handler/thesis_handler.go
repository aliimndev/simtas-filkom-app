package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type ThesisHandler struct {
	thesisUseCase *usecase.ThesisUseCase
}

func NewThesisHandler(uc *usecase.ThesisUseCase) *ThesisHandler {
	return &ThesisHandler{thesisUseCase: uc}
}

// CreateThesis handles POST /api/v1/theses (Mahasiswa only)
func (h *ThesisHandler) CreateThesis(c *gin.Context) {
	var req usecase.CreateThesisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: title, abstract, dan thesis_type wajib diisi")
		return
	}

	studentID := userIDFromContext(c)
	actor := actorFromContext(c)

	thesis, err := h.thesisUseCase.Submit(c.Request.Context(), req, studentID, actor)
	if err != nil {
		h.respondThesisError(c, err)
		return
	}
	response.Created(c, thesis)
}

// ListTheses handles GET /api/v1/theses (all roles, scoped by role)
func (h *ThesisHandler) ListTheses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	var academicYearID, supervisorID uuid.UUID
	var err error
	if v := c.Query("academic_year_id"); v != "" {
		academicYearID, err = uuid.Parse(v)
		if err != nil {
			response.BadRequest(c, "academic_year_id tidak valid")
			return
		}
	}
	if v := c.Query("supervisor_id"); v != "" {
		supervisorID, err = uuid.Parse(v)
		if err != nil {
			response.BadRequest(c, "supervisor_id tidak valid")
			return
		}
	}

	filter := repositoryThesisFilter(
		c.Query("status"),
		academicYearID,
		c.Query("study_program"),
		c.Query("field_of_study"),
		supervisorID,
		c.Query("search"),
		page,
		perPage,
	)

	theses, total, err := h.thesisUseCase.List(
		c.Request.Context(),
		filter,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar thesis")
		return
	}
	response.Paginated(c, theses, page, perPage, total)
}

// GetThesis handles GET /api/v1/theses/:id (all roles, scoped by role)
func (h *ThesisHandler) GetThesis(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	thesis, err := h.thesisUseCase.GetByID(
		c.Request.Context(),
		id,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			response.Forbidden(c, "Akses ditolak: Anda tidak memiliki akses ke thesis ini")
			return
		}
		if errors.Is(err, usecase.ErrThesisNotFound) {
			response.NotFound(c, "Thesis tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal mengambil data thesis")
		return
	}
	response.Success(c, http.StatusOK, thesis)
}

// ReviewThesis handles PUT /api/v1/theses/:id/review (Kaprodi only)
func (h *ThesisHandler) ReviewThesis(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	var req usecase.ReviewThesisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: decision wajib diisi")
		return
	}

	thesis, err := h.thesisUseCase.Review(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		h.respondThesisError(c, err)
		return
	}
	response.Success(c, http.StatusOK, thesis)
}

// AssignSupervisor handles PUT /api/v1/theses/:id/assign-supervisor (Kaprodi only)
func (h *ThesisHandler) AssignSupervisor(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	var req usecase.AssignSupervisorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: supervisor_ids wajib diisi")
		return
	}

	thesis, err := h.thesisUseCase.AssignSupervisor(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		h.respondThesisError(c, err)
		return
	}
	response.Success(c, http.StatusOK, thesis)
}

// CancelThesis handles PATCH /api/v1/theses/:id/cancel (Admin + Kaprodi only)
func (h *ThesisHandler) CancelThesis(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	var req usecase.CancelThesisRequest
	// Body is optional (reason may be omitted).
	_ = c.ShouldBindJSON(&req)

	err := h.thesisUseCase.Cancel(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		h.respondThesisError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Thesis berhasil dibatalkan"})
}

// ListLecturers handles GET /api/v1/lecturers (Kaprodi + Admin only)
func (h *ThesisHandler) ListLecturers(c *gin.Context) {
	lecturers, err := h.thesisUseCase.ListLecturers(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar dosen pembimbing")
		return
	}
	response.Success(c, http.StatusOK, lecturers)
}

// userIDFromContext returns the authenticated user's UUID (uuid.Nil if absent).
func userIDFromContext(c *gin.Context) uuid.UUID {
	v, ok := c.Get("userID")
	if !ok {
		return uuid.Nil
	}
	s, ok := v.(string)
	if !ok {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// roleFromContext returns the authenticated user's role name.
func roleFromContext(c *gin.Context) string {
	v, ok := c.Get("userRole")
	if !ok {
		return ""
	}
	role, _ := v.(string)
	return role
}

// repositoryThesisFilter builds a repository.ThesisFilter from query params.
func repositoryThesisFilter(
	status string,
	academicYearID uuid.UUID,
	studyProgram, fieldOfStudy string,
	supervisorID uuid.UUID,
	search string,
	page, perPage int,
) repository.ThesisFilter {
	return repository.ThesisFilter{
		Status:         status,
		AcademicYearID: academicYearID,
		StudyProgram:   studyProgram,
		FieldOfStudy:   fieldOfStudy,
		SupervisorID:   supervisorID,
		Search:         search,
		Page:           page,
		PerPage:        perPage,
	}
}

// respondThesisError maps use case errors to HTTP responses.
func (h *ThesisHandler) respondThesisError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrThesisNotFound):
		response.NotFound(c, "Thesis tidak ditemukan")
	case errors.Is(err, usecase.ErrForbidden):
		response.Forbidden(c, "Akses ditolak")
	case errors.Is(err, usecase.ErrActiveThesisExists),
		errors.Is(err, usecase.ErrTitleTooShort),
		errors.Is(err, usecase.ErrTitleTooLong),
		errors.Is(err, usecase.ErrAbstractTooShort),
		errors.Is(err, usecase.ErrInvalidThesisType),
		errors.Is(err, usecase.ErrInvalidDecision),
		errors.Is(err, usecase.ErrInvalidSupervisorCount),
		errors.Is(err, usecase.ErrSupervisorNotEligible),
		errors.Is(err, usecase.ErrThesisAlreadyCancelled),
		errors.Is(err, usecase.ErrNoActiveAcademicYear):
		response.BadRequest(c, err.Error())
	case errors.Is(err, usecase.ErrInvalidStateTransition):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), err)
	default:
		response.InternalError(c, "Terjadi kesalahan server")
	}
}
