package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type AcademicYearHandler struct {
	academicYearUseCase *usecase.AcademicYearUseCase
}

func NewAcademicYearHandler(uc *usecase.AcademicYearUseCase) *AcademicYearHandler {
	return &AcademicYearHandler{academicYearUseCase: uc}
}

// List handles GET /api/v1/academic-years (all authenticated)
func (h *AcademicYearHandler) List(c *gin.Context) {
	years, err := h.academicYearUseCase.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar tahun akademik")
		return
	}
	response.Success(c, http.StatusOK, years)
}

// Create handles POST /api/v1/academic-years (admin only)
func (h *AcademicYearHandler) Create(c *gin.Context) {
	var req usecase.AcademicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: name, semester, start_date, end_date wajib diisi")
		return
	}

	year, err := h.academicYearUseCase.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidSemester) {
			response.BadRequest(c, "Semester harus ganjil atau genap")
			return
		}
		if errors.Is(err, usecase.ErrInvalidDateRange) {
			response.BadRequest(c, "Tanggal akhir harus setelah tanggal mulai")
			return
		}
		response.InternalError(c, "Gagal membuat tahun akademik")
		return
	}
	response.Created(c, year)
}

// Update handles PUT /api/v1/academic-years/:id (admin only)
func (h *AcademicYearHandler) Update(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	var req usecase.AcademicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: name, semester, start_date, end_date wajib diisi")
		return
	}

	year, err := h.academicYearUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrAcademicYearNotFound):
			response.NotFound(c, "Tahun akademik tidak ditemukan")
		case errors.Is(err, usecase.ErrAcademicYearInUse):
			response.Error(c, http.StatusConflict, "Tahun akademik aktif dengan skripsi berjalan tidak dapat diubah", err)
		case errors.Is(err, usecase.ErrInvalidSemester):
			response.BadRequest(c, "Semester harus ganjil atau genap")
		case errors.Is(err, usecase.ErrInvalidDateRange):
			response.BadRequest(c, "Tanggal akhir harus setelah tanggal mulai")
		default:
			response.InternalError(c, "Gagal memperbarui tahun akademik")
		}
		return
	}
	response.Success(c, http.StatusOK, year)
}

// Activate handles PATCH /api/v1/academic-years/:id/activate (admin only)
func (h *AcademicYearHandler) Activate(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	err := h.academicYearUseCase.Activate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrAcademicYearNotFound) {
			response.NotFound(c, "Tahun akademik tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal mengaktifkan tahun akademik")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Tahun akademik berhasil diaktifkan"})
}
