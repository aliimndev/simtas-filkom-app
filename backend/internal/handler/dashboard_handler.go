package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type DashboardHandler struct {
	dashboardUseCase *usecase.DashboardUseCase
}

func NewDashboardHandler(uc *usecase.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{dashboardUseCase: uc}
}

// parseDashboardFilter extracts the shared dashboard query params.
func parseDashboardFilter(c *gin.Context) domainRepo.DashboardFilter {
	filter := domainRepo.DashboardFilter{
		Semester:     c.Query("semester"),
		StudyProgram: c.Query("study_program"),
	}
	if v := c.Query("academic_year_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AcademicYearID = &id
		}
	}
	return filter
}

// Summary godoc
// @Summary      Ringkasan dashboard
// @Description  Ringkasan akademik (Admin + Kaprodi only)
// @Tags         Dashboard
// @Produce      json
// @Param        academic_year_id query string false "Filter tahun akademik"
// @Param        semester query string false "Filter semester (ganjil/genap)"
// @Param        study_program query string false "Filter program studi"
// @Success      200  {object}  response.APIResponse{data=usecase.DashboardSummaryResponse} "Ringkasan dashboard"
// @Security     BearerAuth
// @Router       /dashboard/summary [get]
func (h *DashboardHandler) Summary(c *gin.Context) {
	resp, err := h.dashboardUseCase.Summary(c.Request.Context(), parseDashboardFilter(c))
	if err != nil {
		response.InternalError(c, "Gagal memuat ringkasan dashboard")
		return
	}
	response.Success(c, http.StatusOK, resp)
}

// LecturerAnalytics godoc
// @Summary      Analitik dosen
// @Description  Beban bimbingan dan pengujian per dosen (Admin + Kaprodi only)
// @Tags         Dashboard
// @Produce      json
// @Param        academic_year_id query string false "Filter tahun akademik"
// @Param        semester query string false "Filter semester (ganjil/genap)"
// @Param        study_program query string false "Filter program studi"
// @Success      200  {object}  response.APIResponse{data=usecase.LecturerAnalyticsResponse} "Analitik dosen"
// @Security     BearerAuth
// @Router       /dashboard/lecturer-analytics [get]
func (h *DashboardHandler) LecturerAnalytics(c *gin.Context) {
	resp, err := h.dashboardUseCase.LecturerAnalytics(c.Request.Context(), parseDashboardFilter(c))
	if err != nil {
		response.InternalError(c, "Gagal memuat analitik dosen")
		return
	}
	response.Success(c, http.StatusOK, resp)
}

// Operational godoc
// @Summary      Data operasional
// @Description  Pending actions, jadwal mendatang, dan aktivitas (Admin + Kaprodi only)
// @Tags         Dashboard
// @Produce      json
// @Param        days query int false "Jumlah hari jadwal mendatang (default 7)"
// @Success      200  {object}  response.APIResponse{data=usecase.OperationalResponse} "Data operasional"
// @Security     BearerAuth
// @Router       /dashboard/operational [get]
func (h *DashboardHandler) Operational(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days < 1 || days > 90 {
		days = 7
	}
	resp, err := h.dashboardUseCase.Operational(c.Request.Context(), days)
	if err != nil {
		response.InternalError(c, "Gagal memuat data operasional")
		return
	}
	response.Success(c, http.StatusOK, resp)
}

// Student godoc
// @Summary      Dashboard mahasiswa
// @Description  Progress skripsi mahasiswa yang sedang login (Mahasiswa only)
// @Tags         Dashboard
// @Produce      json
// @Success      200  {object}  response.APIResponse{data=usecase.StudentProgressResponse} "Progress mahasiswa"
// @Failure      404  {object}  response.APIResponse "Thesis tidak ditemukan"
// @Security     BearerAuth
// @Router       /dashboard/student [get]
func (h *DashboardHandler) Student(c *gin.Context) {
	resp, err := h.dashboardUseCase.Student(c.Request.Context(), userIDFromContext(c))
	if err != nil {
		if errors.Is(err, usecase.ErrThesisNotFound) {
			response.NotFound(c, "Thesis tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal memuat dashboard mahasiswa")
		return
	}
	response.Success(c, http.StatusOK, resp)
}

// Supervisor godoc
// @Summary      Dashboard dosen pembimbing
// @Description  Ringkasan mahasiswa bimbingan (Dosen Pembimbing only)
// @Tags         Dashboard
// @Produce      json
// @Success      200  {object}  response.APIResponse{data=usecase.SupervisorDashboardResponse} "Ringkasan bimbingan"
// @Security     BearerAuth
// @Router       /dashboard/supervisor [get]
func (h *DashboardHandler) Supervisor(c *gin.Context) {
	resp, err := h.dashboardUseCase.Supervisor(c.Request.Context(), userIDFromContext(c))
	if err != nil {
		response.InternalError(c, "Gagal memuat dashboard dosen pembimbing")
		return
	}
	response.Success(c, http.StatusOK, resp)
}

// Examiner godoc
// @Summary      Dashboard dosen penguji
// @Description  Jadwal pengujian dan pending scores (Dosen Penguji only)
// @Tags         Dashboard
// @Produce      json
// @Success      200  {object}  response.APIResponse{data=usecase.ExaminerDashboardResponse} "Jadwal pengujian"
// @Security     BearerAuth
// @Router       /dashboard/examiner [get]
func (h *DashboardHandler) Examiner(c *gin.Context) {
	resp, err := h.dashboardUseCase.Examiner(c.Request.Context(), userIDFromContext(c))
	if err != nil {
		response.InternalError(c, "Gagal memuat dashboard dosen penguji")
		return
	}
	response.Success(c, http.StatusOK, resp)
}
