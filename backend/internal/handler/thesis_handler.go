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
)

type ThesisHandler struct {
	thesisUseCase *usecase.ThesisUseCase
}

func NewThesisHandler(uc *usecase.ThesisUseCase) *ThesisHandler {
	return &ThesisHandler{thesisUseCase: uc}
}

// CreateThesis godoc
// @Summary      Ajukan judul skripsi
// @Description  Mengajukan judul skripsi baru beserta draft proposal PDF (Mahasiswa only)
// @Tags         Thesis Submission
// @Accept       multipart/form-data
// @Produce      json
// @Param        title formData string true "Judul skripsi"
// @Param        abstract formData string true "Abstrak"
// @Param        field_of_study formData string false "Bidang keahlian"
// @Param        thesis_type formData string true "skripsi atau tugas_akhir"
// @Param        file formData file true "Draft proposal (PDF, maksimal 10 MB)"
// @Success      201  {object}  response.APIResponse{data=usecase.ThesisDetail} "Pengajuan dibuat"
// @Failure      400  {object}  response.APIResponse "Data tidak valid / sudah punya thesis aktif"
// @Security     BearerAuth
// @Router       /theses [post]
func (h *ThesisHandler) CreateThesis(c *gin.Context) {
	var req usecase.CreateThesisRequest
	req.Title = strings.TrimSpace(c.PostForm("title"))
	req.Abstract = strings.TrimSpace(c.PostForm("abstract"))
	req.FieldOfStudy = strings.TrimSpace(c.PostForm("field_of_study"))
	req.ThesisType = strings.TrimSpace(c.PostForm("thesis_type"))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "Request tidak valid: file draft proposal wajib diunggah")
		return
	}
	defer file.Close()
	req.DraftFile = file
	req.DraftHeader = header

	studentID := userIDFromContext(c)
	actor := actorFromContext(c)

	thesis, err := h.thesisUseCase.Submit(c.Request.Context(), req, studentID, actor)
	if err != nil {
		h.respondThesisError(c, err)
		return
	}
	response.Created(c, thesis)
}

// ListTheses godoc
// @Summary      Daftar thesis
// @Description  Mengambil daftar thesis (scope per role: admin/kaprodi semua, dosen pembimbing miliknya, mahasiswa miliknya)
// @Tags         Thesis Submission
// @Produce      json
// @Param        status query string false "Filter status"
// @Param        academic_year_id query string false "Filter tahun akademik"
// @Param        study_program query string false "Filter program studi"
// @Param        field_of_study query string false "Filter bidang keahlian"
// @Param        supervisor_id query string false "Filter dosen pembimbing"
// @Param        search query string false "Cari judul/nama/NIM"
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 20)"
// @Success      200  {object}  response.APIResponse "Daftar thesis"
// @Security     BearerAuth
// @Router       /theses [get]
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

// GetThesis godoc
// @Summary      Detail thesis
// @Description  Mengambil detail thesis lengkap (scope per role)
// @Tags         Thesis Submission
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.ThesisDetail} "Detail thesis"
// @Failure      403  {object}  response.APIResponse "Akses ditolak"
// @Failure      404  {object}  response.APIResponse "Thesis tidak ditemukan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id} [get]
func (h *ThesisHandler) GetThesis(c *gin.Context) {
	id, ok := parseThesisIDParam(c)
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

// ReviewThesis godoc
// @Summary      Review pengajuan judul
// @Description  Menyetujui atau menolak pengajuan judul (Kaprodi only)
// @Tags         Thesis Submission
// @Accept       json
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        body body usecase.ReviewThesisRequest true "Keputusan review"
// @Success      200  {object}  response.APIResponse{data=usecase.ThesisDetail} "Hasil review"
// @Failure      400  {object}  response.APIResponse "Decision tidak valid"
// @Failure      422  {object}  response.APIResponse "Status bukan submitted"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/review [put]
func (h *ThesisHandler) ReviewThesis(c *gin.Context) {
	id, ok := parseThesisIDParam(c)
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

// AssignSupervisor godoc
// @Summary      Tunjuk dosen pembimbing
// @Description  Menetapkan 1-2 dosen pembimbing untuk thesis (Kaprodi only)
// @Tags         Thesis Submission
// @Accept       json
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        body body usecase.AssignSupervisorRequest true "Daftar supervisor"
// @Success      200  {object}  response.APIResponse{data=usecase.ThesisDetail} "Thesis dengan supervisor"
// @Failure      400  {object}  response.APIResponse "Supervisor tidak valid"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/assign-supervisor [put]
func (h *ThesisHandler) AssignSupervisor(c *gin.Context) {
	id, ok := parseThesisIDParam(c)
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

// CancelThesis godoc
// @Summary      Batalkan thesis
// @Description  Membatalkan thesis (Admin + Kaprodi only)
// @Tags         Thesis Submission
// @Accept       json
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        body body usecase.CancelThesisRequest false "Alasan pembatalan (opsional)"
// @Success      200  {object}  response.APIResponse "Thesis berhasil dibatalkan"
// @Failure      400  {object}  response.APIResponse "Thesis sudah dibatalkan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/cancel [patch]
func (h *ThesisHandler) CancelThesis(c *gin.Context) {
	id, ok := parseThesisIDParam(c)
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

// ListLecturers godoc
// @Summary      Daftar dosen pembimbing
// @Description  Mengambil daftar dosen pembimbing aktif dengan beban bimbingan (Kaprodi + Admin only)
// @Tags         Thesis Submission
// @Produce      json
// @Success      200  {object}  response.APIResponse "Daftar dosen pembimbing"
// @Security     BearerAuth
// @Router       /lecturers [get]
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

// respondThesisError delegates to the central error catalog (see errors.go).
func (h *ThesisHandler) respondThesisError(c *gin.Context, err error) {
	respondError(c, err)
}
