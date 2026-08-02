package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

type ArchiveHandler struct {
	archiveUseCase *usecase.ArchiveUseCase
}

func NewArchiveHandler(uc *usecase.ArchiveUseCase) *ArchiveHandler {
	return &ArchiveHandler{archiveUseCase: uc}
}

// Create godoc
// @Summary      Arsipkan skripsi
// @Description  Mengarsipkan skripsi final setelah yudisium (Mahasiswa pemilik + Admin)
// @Tags         Archives
// @Accept       multipart/form-data
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        file formData file true "File PDF skripsi final (max 25 MB)"
// @Param        abstract_id formData string true "Abstrak Bahasa Indonesia (min 50 kata)"
// @Param        abstract_en formData string false "Abstrak Bahasa Inggris"
// @Param        keywords formData string true "Kata kunci dipisah koma (min 3)"
// @Param        graduation_year formData int true "Tahun lulus"
// @Success      201  {object}  response.APIResponse{data=usecase.ArchiveDetail} "Arsip dibuat"
// @Failure      422  {object}  response.APIResponse "Thesis belum graduated"
// @Failure      409  {object}  response.APIResponse "Arsip sudah ada"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/archive [post]
func (h *ArchiveHandler) Create(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File wajib diunggah pada field 'file'")
		return
	}
	defer file.Close()

	graduationYear := 0
	if v := c.PostForm("graduation_year"); v != "" {
		graduationYear, err = strconv.Atoi(v)
		if err != nil {
			response.BadRequest(c, "graduation_year harus berupa angka")
			return
		}
	}

	keywords := splitKeywords(c.PostForm("keywords"))

	detail, err := h.archiveUseCase.Create(
		c.Request.Context(),
		thesisID,
		file,
		header,
		usecase.CreateArchiveRequest{
			AbstractID:     c.PostForm("abstract_id"),
			AbstractEN:     c.PostForm("abstract_en"),
			Keywords:       keywords,
			GraduationYear: graduationYear,
		},
		actorFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondArchiveError(c, err)
		return
	}
	response.Created(c, detail)
}

// Search godoc
// @Summary      Cari arsip
// @Description  Mencari arsip skripsi (full-text search + filter) — semua user terautentikasi
// @Tags         Archives
// @Produce      json
// @Param        q query string false "Full-text search"
// @Param        year query int false "Filter tahun lulus"
// @Param        field_of_study query string false "Filter bidang keahlian"
// @Param        study_program query string false "Filter program studi"
// @Param        supervisor_id query string false "Filter dosen pembimbing"
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 20)"
// @Success      200  {object}  response.APIResponse "Daftar arsip"
// @Security     BearerAuth
// @Router       /archives [get]
func (h *ArchiveHandler) Search(c *gin.Context) {
	page, perPage := parsePagination(c)
	filter := repository.ArchiveFilter{
		Query:          c.Query("q"),
		GraduationYear: parseYear(c.Query("year")),
		FieldOfStudy:   c.Query("field_of_study"),
		StudyProgram:   c.Query("study_program"),
		Page:           page,
		PerPage:        perPage,
	}
	if v := c.Query("supervisor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.BadRequest(c, "supervisor_id tidak valid")
			return
		}
		filter.SupervisorID = &id
	}

	archives, total, err := h.archiveUseCase.Search(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Gagal mengambil arsip")
		return
	}
	response.Paginated(c, archives, page, perPage, total)
}

// Get godoc
// @Summary      Detail arsip
// @Description  Mengambil detail lengkap arsip (semua user terautentikasi)
// @Tags         Archives
// @Produce      json
// @Param        id path string true "Archive ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.ArchiveDetail} "Detail arsip"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /archives/{id} [get]
func (h *ArchiveHandler) Get(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	detail, err := h.archiveUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.respondArchiveError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// GetByThesis godoc
// @Summary      Detail arsip per thesis
// @Description  Shortcut untuk mengambil arsip berdasarkan thesis ID
// @Tags         Archives
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.ArchiveDetail} "Detail arsip"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/archive [get]
func (h *ArchiveHandler) GetByThesis(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	detail, err := h.archiveUseCase.GetByThesisID(c.Request.Context(), thesisID)
	if err != nil {
		h.respondArchiveError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Download godoc
// @Summary      Download arsip
// @Description  Menghasilkan presigned URL download arsip (expired 30 menit, role-based access)
// @Tags         Archives
// @Produce      json
// @Param        id path string true "Archive ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.ArchiveDownloadResult} "Presigned URL"
// @Failure      403  {object}  response.APIResponse "Mahasiswa non-pemilik ditolak"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /archives/{id}/download [get]
func (h *ArchiveHandler) Download(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	result, err := h.archiveUseCase.Download(
		c.Request.Context(),
		id,
		userIDFromContext(c),
		roleFromContext(c),
		actorFromContext(c),
	)
	if err != nil {
		h.respondArchiveError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

// Stats godoc
// @Summary      Statistik arsip
// @Description  Mengambil statistik arsip agregat (Admin + Kaprodi only)
// @Tags         Archives
// @Produce      json
// @Success      200  {object}  response.APIResponse "Statistik arsip"
// @Security     BearerAuth
// @Router       /archives/stats [get]
func (h *ArchiveHandler) Stats(c *gin.Context) {
	stats, err := h.archiveUseCase.Stats(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil statistik arsip")
		return
	}
	response.Success(c, http.StatusOK, stats)
}

// respondArchiveError maps archive use case errors to HTTP responses.
func (h *ArchiveHandler) respondArchiveError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrArchiveNotFound),
		errors.Is(err, usecase.ErrThesisNotFound):
		response.NotFound(c, "Arsip atau thesis tidak ditemukan")
	case errors.Is(err, usecase.ErrForbidden),
		errors.Is(err, usecase.ErrArchiveDownloadDenied):
		response.Forbidden(c, "Akses ditolak")
	case errors.Is(err, utils.ErrNotPDF):
		response.BadRequest(c, "Hanya file PDF yang diizinkan")
	case errors.Is(err, utils.ErrFileTooLarge):
		response.BadRequest(c, "Ukuran file maksimal 25 MB")
	case errors.Is(err, utils.ErrFileEmpty),
		errors.Is(err, usecase.ErrArchiveFileRequired):
		response.BadRequest(c, "File wajib diunggah")
	case errors.Is(err, usecase.ErrArchiveAbstractShort),
		errors.Is(err, usecase.ErrArchiveKeywordsShort),
		errors.Is(err, usecase.ErrArchiveInvalidYear):
		response.BadRequest(c, err.Error())
	case errors.Is(err, usecase.ErrArchiveExists):
		response.Error(c, http.StatusConflict, err.Error(), err)
	// Gate failure is a state violation → 422 per docs/phase-3-supporting-features/10-archive-module.md.
	case errors.Is(err, usecase.ErrArchiveThesisNotGrad):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), err)
	default:
		response.InternalError(c, "Terjadi kesalahan server")
	}
}

// splitKeywords splits a comma-separated keyword string, trimming whitespace
// and dropping empties.
func splitKeywords(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseYear safely parses a query year (0 = no filter).
func parseYear(raw string) int {
	if raw == "" {
		return 0
	}
	y, err := strconv.Atoi(raw)
	if err != nil || y < 0 {
		return 0
	}
	return y
}
