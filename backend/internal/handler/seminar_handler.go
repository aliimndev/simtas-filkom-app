package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type SeminarHandler struct {
	seminarUseCase *usecase.SeminarUseCase
}

func NewSeminarHandler(uc *usecase.SeminarUseCase) *SeminarHandler {
	return &SeminarHandler{seminarUseCase: uc}
}

// Submit godoc
// @Summary      Ajukan seminar proposal
// @Description  Mengajukan seminar proposal (gate: dokumen seminar approved) (Mahasiswa pemilik only)
// @Tags         Seminar Proposal
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Success      201  {object}  response.APIResponse{data=usecase.SeminarDetail} "Pengajuan dibuat"
// @Failure      422  {object}  response.APIResponse "Gate belum terpenuhi"
// @Failure      409  {object}  response.APIResponse "Sudah ada seminar aktif"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/seminars [post]
func (h *SeminarHandler) Submit(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	detail, err := h.seminarUseCase.Submit(c.Request.Context(), thesisID, actorFromContext(c))
	if err != nil {
		h.respondSeminarError(c, err)
		return
	}
	response.Created(c, detail)
}

// List godoc
// @Summary      Daftar seminar
// @Description  Mengambil daftar seminar (scope per role)
// @Tags         Seminar Proposal
// @Produce      json
// @Param        status query string false "Filter status"
// @Param        examiner_id query string false "Filter penguji"
// @Param        date_from query string false "Filter tanggal mulai (RFC3339)"
// @Param        date_to query string false "Filter tanggal akhir (RFC3339)"
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 20)"
// @Success      200  {object}  response.APIResponse "Daftar seminar"
// @Security     BearerAuth
// @Router       /seminars [get]
func (h *SeminarHandler) List(c *gin.Context) {
	page, perPage := parsePagination(c)
	filter := repository.SeminarFilter{
		Status:  c.Query("status"),
		Page:    page,
		PerPage: perPage,
	}
	if v := c.Query("examiner_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.BadRequest(c, "examiner_id tidak valid")
			return
		}
		filter.ExaminerID = id
	}
	if from := c.Query("date_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			response.BadRequest(c, "date_from harus format RFC3339")
			return
		}
		filter.DateFrom = &t
	}
	if to := c.Query("date_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			response.BadRequest(c, "date_to harus format RFC3339")
			return
		}
		filter.DateTo = &t
	}

	seminars, total, err := h.seminarUseCase.List(
		c.Request.Context(),
		filter,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar seminar")
		return
	}
	response.Paginated(c, seminars, page, perPage, total)
}

// Get godoc
// @Summary      Detail seminar
// @Description  Mengambil detail seminar lengkap
// @Tags         Seminar Proposal
// @Produce      json
// @Param        id path string true "Seminar ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.SeminarDetail} "Detail seminar"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /seminars/{id} [get]
func (h *SeminarHandler) Get(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	detail, err := h.seminarUseCase.GetByID(
		c.Request.Context(),
		id,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondSeminarError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Schedule godoc
// @Summary      Jadwalkan seminar
// @Description  Menjadwalkan/menjadwal-ulang seminar dan menunjuk penguji (Admin + Kaprodi only)
// @Tags         Seminar Proposal
// @Accept       json
// @Produce      json
// @Param        id path string true "Seminar ID (UUID)"
// @Param        body body usecase.ScheduleSeminarRequest true "Jadwal + penguji"
// @Success      200  {object}  response.APIResponse{data=usecase.SeminarDetail} "Seminar terjadwal"
// @Failure      400  {object}  response.APIResponse "Jadwal tidak valid"
// @Failure      409  {object}  response.APIResponse "Konflik ruangan/penguji"
// @Security     BearerAuth
// @Router       /seminars/{id}/schedule [put]
func (h *SeminarHandler) Schedule(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	var req usecase.ScheduleSeminarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: scheduled_at, room, dan examiner_ids wajib diisi")
		return
	}
	detail, err := h.seminarUseCase.Schedule(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		h.respondSeminarError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// SubmitScores godoc
// @Summary      Input nilai seminar
// @Description  Menginput nilai seminar per komponen (Dosen Penguji ditugaskan only)
// @Tags         Seminar Proposal
// @Accept       json
// @Produce      json
// @Param        id path string true "Seminar ID (UUID)"
// @Param        body body usecase.SubmitScoreRequest true "Nilai per komponen"
// @Success      200  {object}  response.APIResponse{data=usecase.SeminarResult} "Nilai tersimpan / hasil final"
// @Failure      403  {object}  response.APIResponse "Bukan penguji ditugaskan"
// @Failure      409  {object}  response.APIResponse "Sudah pernah input nilai"
// @Security     BearerAuth
// @Router       /seminars/{id}/scores [post]
func (h *SeminarHandler) SubmitScores(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	var req usecase.SubmitScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: scores wajib diisi")
		return
	}
	result, err := h.seminarUseCase.SubmitScores(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		h.respondSeminarError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

// Result godoc
// @Summary      Hasil seminar
// @Description  Mengambil hasil seminar lengkap (nilai akhir + breakdown per penguji)
// @Tags         Seminar Proposal
// @Produce      json
// @Param        id path string true "Seminar ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.SeminarResult} "Hasil seminar"
// @Security     BearerAuth
// @Router       /seminars/{id}/result [get]
func (h *SeminarHandler) Result(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	result, err := h.seminarUseCase.Result(
		c.Request.Context(),
		id,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondSeminarError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

// SetRevisionNotes godoc
// @Summary      Catat revisi seminar
// @Description  Mencatat catatan revisi pasca seminar (Admin + Kaprodi only)
// @Tags         Seminar Proposal
// @Accept       json
// @Produce      json
// @Param        id path string true "Seminar ID (UUID)"
// @Param        body body handler.revisionNotesRequest true "Catatan revisi"
// @Success      200  {object}  response.APIResponse{data=usecase.SeminarDetail} "Catatan tersimpan"
// @Failure      422  {object}  response.APIResponse "Seminar belum passed"
// @Security     BearerAuth
// @Router       /seminars/{id}/revision [put]
func (h *SeminarHandler) SetRevisionNotes(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	var req revisionNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid")
		return
	}
	detail, err := h.seminarUseCase.SetRevisionNotes(c.Request.Context(), id, req.RevisionNotes, actorFromContext(c))
	if err != nil {
		h.respondSeminarError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// respondSeminarError maps seminar use case errors to HTTP responses.
func (h *SeminarHandler) respondSeminarError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrSeminarNotFound):
		response.NotFound(c, "Seminar tidak ditemukan")
	case errors.Is(err, usecase.ErrForbidden),
		errors.Is(err, usecase.ErrSeminarNotExaminer):
		response.Forbidden(c, "Akses ditolak")
	case errors.Is(err, usecase.ErrSeminarInvalidDecision),
		errors.Is(err, usecase.ErrSeminarScheduleLeadTime),
		errors.Is(err, usecase.ErrSeminarMinExaminers),
		errors.Is(err, usecase.ErrSeminarInvalidExaminer),
		errors.Is(err, usecase.ErrSeminarIncompleteScore),
		errors.Is(err, usecase.ErrSeminarInvalidScore):
		response.BadRequest(c, err.Error())
	case errors.Is(err, usecase.ErrSeminarActiveExists),
		errors.Is(err, usecase.ErrSeminarRoomConflict),
		errors.Is(err, usecase.ErrSeminarAlreadyScored):
		response.Error(c, http.StatusConflict, err.Error(), err)
	// Gate failure is a state violation → 422 per docs/phase-2-core-backend/08-seminar-module.md.
	case errors.Is(err, usecase.ErrSeminarGateNotMet),
		errors.Is(err, usecase.ErrSeminarNotScheduled),
		errors.Is(err, usecase.ErrSeminarNotPassed):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), err)
	default:
		response.InternalError(c, "Terjadi kesalahan server")
	}
}
