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

type ConsultationHandler struct {
	consultationUseCase *usecase.ConsultationUseCase
}

func NewConsultationHandler(uc *usecase.ConsultationUseCase) *ConsultationHandler {
	return &ConsultationHandler{consultationUseCase: uc}
}

// Create godoc
// @Summary      Catat log bimbingan
// @Description  Mencatat log konsultasi bimbingan (Mahasiswa pemilik + Dosen Pembimbing)
// @Tags         Supervision
// @Accept       json
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        body body usecase.CreateConsultationRequest true "Data log konsultasi"
// @Success      201  {object}  response.APIResponse{data=usecase.ConsultationDetail} "Log dibuat"
// @Failure      400  {object}  response.APIResponse "Data tidak valid / tanggal masa depan"
// @Failure      403  {object}  response.APIResponse "Bukan pemilik/pembimbing"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/consultations [post]
func (h *ConsultationHandler) Create(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	var req usecase.CreateConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: consultation_date dan topics_discussed wajib diisi")
		return
	}

	detail, err := h.consultationUseCase.Create(
		c.Request.Context(),
		thesisID,
		req,
		actorFromContext(c),
	)
	if err != nil {
		h.respondConsultationError(c, err)
		return
	}
	response.Created(c, detail)
}

// List godoc
// @Summary      Daftar log bimbingan
// @Description  Mengambil daftar log konsultasi dengan summary (pemilik/pembimbing/admin/kaprodi)
// @Tags         Supervision
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        status query string false "Filter status (pending/approved)"
// @Param        date_from query string false "Filter tanggal mulai (YYYY-MM-DD)"
// @Param        date_to query string false "Filter tanggal akhir (YYYY-MM-DD)"
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 20)"
// @Success      200  {object}  response.APIResponse "Daftar + summary log konsultasi"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/consultations [get]
func (h *ConsultationHandler) List(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

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

	filter := repository.ConsultationFilter{
		Status:   c.Query("status"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Page:     page,
		PerPage:  perPage,
	}

	result, total, err := h.consultationUseCase.List(
		c.Request.Context(),
		thesisID,
		filter,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondConsultationError(c, err)
		return
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	response.SuccessWithMeta(c, result, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// Get godoc
// @Summary      Detail log bimbingan
// @Description  Mengambil detail satu log konsultasi
// @Tags         Supervision
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        id path string true "Consultation ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.ConsultationDetail} "Detail log"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/consultations/{id} [get]
func (h *ConsultationHandler) Get(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	detail, err := h.consultationUseCase.GetByID(
		c.Request.Context(),
		thesisID,
		id,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondConsultationError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Update godoc
// @Summary      Update log bimbingan
// @Description  Memperbarui log konsultasi (pembuat log, selama status pending)
// @Tags         Supervision
// @Accept       json
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        id path string true "Consultation ID (UUID)"
// @Param        body body usecase.UpdateConsultationRequest true "Data yang diperbarui"
// @Success      200  {object}  response.APIResponse{data=usecase.ConsultationDetail} "Log diperbarui"
// @Failure      422  {object}  response.APIResponse "Log sudah di-approve"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/consultations/{id} [put]
func (h *ConsultationHandler) Update(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	var req usecase.UpdateConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid")
		return
	}

	detail, err := h.consultationUseCase.Update(c.Request.Context(), thesisID, id, req, actorFromContext(c))
	if err != nil {
		h.respondConsultationError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Approve godoc
// @Summary      Setujui log bimbingan
// @Description  Menyetujui log konsultasi (Dosen Pembimbing thesis ini only)
// @Tags         Supervision
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        id path string true "Consultation ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.ConsultationDetail} "Log disetujui"
// @Failure      403  {object}  response.APIResponse "Bukan pembimbing thesis ini"
// @Failure      422  {object}  response.APIResponse "Log sudah di-approve"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/consultations/{id}/approve [patch]
func (h *ConsultationHandler) Approve(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	detail, err := h.consultationUseCase.Approve(
		c.Request.Context(),
		thesisID,
		id,
		userIDFromContext(c),
		actorFromContext(c),
	)
	if err != nil {
		h.respondConsultationError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Delete godoc
// @Summary      Hapus log bimbingan
// @Description  Menghapus log konsultasi (pembuat log, selama status pending)
// @Tags         Supervision
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        id path string true "Consultation ID (UUID)"
// @Success      200  {object}  response.APIResponse "Log berhasil dihapus"
// @Failure      422  {object}  response.APIResponse "Log sudah di-approve"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/consultations/{id} [delete]
func (h *ConsultationHandler) Delete(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	err := h.consultationUseCase.Delete(c.Request.Context(), thesisID, id, userIDFromContext(c), actorFromContext(c))
	if err != nil {
		h.respondConsultationError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Log konsultasi berhasil dihapus"})
}

// Summary godoc
// @Summary      Ringkasan bimbingan
// @Description  Mengambil ringkasan statistik log konsultasi tanpa pagination
// @Tags         Supervision
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.ConsultationSummary} "Ringkasan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/consultations/summary [get]
func (h *ConsultationHandler) Summary(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	summary, err := h.consultationUseCase.Summary(
		c.Request.Context(),
		thesisID,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondConsultationError(c, err)
		return
	}
	response.Success(c, http.StatusOK, summary)
}

func parseThesisIDParam(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("thesis_id"))
	if err != nil {
		response.BadRequest(c, "thesis_id tidak valid")
		return uuid.Nil, false
	}
	return id, true
}

// respondConsultationError maps consultation use case errors to HTTP responses.
func (h *ConsultationHandler) respondConsultationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrThesisNotFound):
		response.NotFound(c, "Thesis tidak ditemukan")
	case errors.Is(err, usecase.ErrConsultationNotFound):
		response.NotFound(c, "Log konsultasi tidak ditemukan")
	case errors.Is(err, usecase.ErrForbidden):
		response.Forbidden(c, "Akses ditolak")
	case errors.Is(err, usecase.ErrNotConsultationCreator):
		response.Forbidden(c, "Hanya pembuat log yang dapat mengubah log ini")
	case errors.Is(err, usecase.ErrNotSupervisorOfThesis):
		response.Forbidden(c, "Hanya dosen pembimbing thesis ini yang dapat menyetujui")
	case errors.Is(err, usecase.ErrConsultationDateFuture),
		errors.Is(err, usecase.ErrInvalidDateFormat),
		errors.Is(err, usecase.ErrTopicsDiscussedRequired),
		errors.Is(err, usecase.ErrThesisNotInProgress):
		response.BadRequest(c, err.Error())
	case errors.Is(err, usecase.ErrConsultationAlreadyDone):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), err)
	default:
		response.InternalError(c, "Terjadi kesalahan server")
	}
}
