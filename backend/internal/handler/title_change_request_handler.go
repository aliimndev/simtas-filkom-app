package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type TitleChangeRequestHandler struct {
	titleChangeUseCase *usecase.TitleChangeRequestUseCase
}

func NewTitleChangeRequestHandler(uc *usecase.TitleChangeRequestUseCase) *TitleChangeRequestHandler {
	return &TitleChangeRequestHandler{titleChangeUseCase: uc}
}

// Create godoc
// @Summary      Ajukan perubahan judul
// @Description  Mengajukan perubahan judul skripsi (Mahasiswa pemilik, thesis approved/in_progress dengan pembimbing)
// @Tags         Title Change Request
// @Accept       json
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        body body usecase.CreateTitleChangeRequest true "Data perubahan judul"
// @Success      201  {object}  response.APIResponse{data=usecase.TitleChangeRequestDetail} "Permintaan dibuat"
// @Failure      400  {object}  response.APIResponse "Validasi gagal"
// @Failure      403  {object}  response.APIResponse "Bukan pemilik thesis"
// @Failure      404  {object}  response.APIResponse "Thesis tidak ditemukan"
// @Failure      409  {object}  response.APIResponse "Sudah ada permintaan pending"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/title-change-requests [post]
func (h *TitleChangeRequestHandler) Create(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	var req usecase.CreateTitleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: requested_title wajib diisi")
		return
	}

	detail, err := h.titleChangeUseCase.Submit(
		c.Request.Context(),
		thesisID,
		req,
		actorFromContext(c),
	)
	if err != nil {
		h.respondTitleChangeError(c, err)
		return
	}
	response.Created(c, detail)
}

// List godoc
// @Summary      Riwayat perubahan judul
// @Description  Mengambil daftar permintaan perubahan judul sebuah thesis (pemilik, pembimbing, Kaprodi, Admin)
// @Tags         Title Change Request
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Success      200  {object}  response.APIResponse "Daftar permintaan"
// @Failure      403  {object}  response.APIResponse "Akses ditolak"
// @Failure      404  {object}  response.APIResponse "Thesis tidak ditemukan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/title-change-requests [get]
func (h *TitleChangeRequestHandler) List(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	list, err := h.titleChangeUseCase.List(
		c.Request.Context(),
		thesisID,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondTitleChangeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, list)
}

// Cancel godoc
// @Summary      Batalkan permintaan perubahan judul
// @Description  Membatalkan permintaan perubahan judul yang masih PENDING (Mahasiswa pemilik)
// @Tags         Title Change Request
// @Produce      json
// @Param        id path string true "Title change request ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.TitleChangeRequestDetail} "Permintaan dibatalkan"
// @Failure      400  {object}  response.APIResponse "Bukan status pending"
// @Failure      403  {object}  response.APIResponse "Bukan pemilik permintaan"
// @Failure      404  {object}  response.APIResponse "Permintaan tidak ditemukan"
// @Security     BearerAuth
// @Router       /title-change-requests/{id}/cancel [post]
func (h *TitleChangeRequestHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id tidak valid")
		return
	}

	detail, err := h.titleChangeUseCase.Cancel(
		c.Request.Context(),
		id,
		actorFromContext(c),
	)
	if err != nil {
		h.respondTitleChangeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// ListPending godoc
// @Summary      Antrian review perubahan judul
// @Description  Daftar permintaan perubahan judul PENDING untuk mahasiswa bimbingan (Dosen Pembimbing)
// @Tags         Title Change Request
// @Produce      json
// @Success      200  {object}  response.APIResponse "Antrian review"
// @Failure      403  {object}  response.APIResponse "Hanya Dosen Pembimbing"
// @Security     BearerAuth
// @Router       /title-change-requests [get]
func (h *TitleChangeRequestHandler) ListPending(c *gin.Context) {
	list, err := h.titleChangeUseCase.ListPendingForSupervisor(
		c.Request.Context(),
		userIDFromContext(c),
	)
	if err != nil {
		h.respondTitleChangeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, list)
}

// Approve godoc
// @Summary      Setujui perubahan judul
// @Description  Menyetujui permintaan perubahan judul (Dosen Pembimbing assigned; PENDING only) dan meng-update judul thesis secara atomik
// @Tags         Title Change Request
// @Accept       json
// @Produce      json
// @Param        id path string true "Title change request ID (UUID)"
// @Param        body body usecase.ReviewTitleChangeRequest false "Catatan persetujuan (opsional)"
// @Success      200  {object}  response.APIResponse{data=usecase.TitleChangeRequestDetail} "Permintaan disetujui"
// @Failure      400  {object}  response.APIResponse "Bukan status pending"
// @Failure      403  {object}  response.APIResponse "Bukan dosen pembimbing thesis ini"
// @Failure      404  {object}  response.APIResponse "Permintaan tidak ditemukan"
// @Security     BearerAuth
// @Router       /title-change-requests/{id}/approve [post]
func (h *TitleChangeRequestHandler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id tidak valid")
		return
	}

	var req usecase.ReviewTitleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.BadRequest(c, "Request tidak valid")
		return
	}

	detail, err := h.titleChangeUseCase.Approve(
		c.Request.Context(),
		id,
		req,
		actorFromContext(c),
	)
	if err != nil {
		h.respondTitleChangeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Reject godoc
// @Summary      Tolak perubahan judul
// @Description  Menolak permintaan perubahan judul (Dosen Pembimbing assigned; PENDING only) dengan catatan wajib
// @Tags         Title Change Request
// @Accept       json
// @Produce      json
// @Param        id path string true "Title change request ID (UUID)"
// @Param        body body usecase.ReviewTitleChangeRequest true "Alasan penolakan (wajib)"
// @Success      200  {object}  response.APIResponse{data=usecase.TitleChangeRequestDetail} "Permintaan ditolak"
// @Failure      400  {object}  response.APIResponse "Bukan status pending / catatan kosong"
// @Failure      403  {object}  response.APIResponse "Bukan dosen pembimbing thesis ini"
// @Failure      404  {object}  response.APIResponse "Permintaan tidak ditemukan"
// @Security     BearerAuth
// @Router       /title-change-requests/{id}/reject [post]
func (h *TitleChangeRequestHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id tidak valid")
		return
	}

	var req usecase.ReviewTitleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: review_notes wajib diisi")
		return
	}

	detail, err := h.titleChangeUseCase.Reject(
		c.Request.Context(),
		id,
		req,
		actorFromContext(c),
	)
	if err != nil {
		h.respondTitleChangeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// respondTitleChangeError maps title change use case errors to HTTP responses.
func (h *TitleChangeRequestHandler) respondTitleChangeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrThesisNotFound),
		errors.Is(err, usecase.ErrTitleChangeNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, usecase.ErrTitleChangeForbidden),
		errors.Is(err, usecase.ErrTitleChangeNotSupervisor):
		response.Forbidden(c, err.Error())
	case errors.Is(err, usecase.ErrTitleChangeNotEligible),
		errors.Is(err, usecase.ErrNoSupervisorAssigned),
		errors.Is(err, usecase.ErrTitleChangeNotPending),
		errors.Is(err, usecase.ErrTitleChangeTitleTooShort),
		errors.Is(err, usecase.ErrTitleChangeTitleTooLong),
		errors.Is(err, usecase.ErrTitleChangeReviewNotesReq):
		response.BadRequest(c, err.Error())
	case errors.Is(err, usecase.ErrPendingTitleChangeExists):
		response.Error(c, http.StatusConflict, err.Error(), err)
	default:
		response.InternalError(c, "Terjadi kesalahan server")
	}
}
