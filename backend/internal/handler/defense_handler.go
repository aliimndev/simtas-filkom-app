package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type DefenseHandler struct {
	defenseUseCase *usecase.DefenseUseCase
}

func NewDefenseHandler(uc *usecase.DefenseUseCase) *DefenseHandler {
	return &DefenseHandler{defenseUseCase: uc}
}

// Submit godoc
// @Summary      Ajukan sidang skripsi
// @Description  Mengajukan sidang skripsi (gate: seminar lulus + dokumen sidang approved) (Mahasiswa pemilik only)
// @Tags         Thesis Defense
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Success      201  {object}  response.APIResponse{data=usecase.DefenseDetail} "Pengajuan dibuat"
// @Failure      422  {object}  response.APIResponse "Gate belum terpenuhi"
// @Failure      409  {object}  response.APIResponse "Sudah ada sidang aktif"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/defenses [post]
func (h *DefenseHandler) Submit(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	detail, err := h.defenseUseCase.Submit(c.Request.Context(), thesisID, actorFromContext(c))
	if err != nil {
		h.respondDefenseError(c, err)
		return
	}
	response.Created(c, detail)
}

// List godoc
// @Summary      Daftar sidang
// @Description  Mengambil daftar sidang (scope per role)
// @Tags         Thesis Defense
// @Produce      json
// @Param        status query string false "Filter status"
// @Param        examiner_id query string false "Filter penguji"
// @Param        date_from query string false "Filter tanggal mulai (RFC3339)"
// @Param        date_to query string false "Filter tanggal akhir (RFC3339)"
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 20)"
// @Success      200  {object}  response.APIResponse "Daftar sidang"
// @Security     BearerAuth
// @Router       /defenses [get]
func (h *DefenseHandler) List(c *gin.Context) {
	page, perPage := parsePagination(c)
	filter := repository.DefenseFilter{
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

	defenses, total, err := h.defenseUseCase.List(
		c.Request.Context(),
		filter,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar sidang")
		return
	}
	response.Paginated(c, defenses, page, perPage, total)
}

// Get godoc
// @Summary      Detail sidang
// @Description  Mengambil detail sidang lengkap
// @Tags         Thesis Defense
// @Produce      json
// @Param        id path string true "Defense ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.DefenseDetail} "Detail sidang"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /defenses/{id} [get]
func (h *DefenseHandler) Get(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	detail, err := h.defenseUseCase.GetByID(
		c.Request.Context(),
		id,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondDefenseError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Schedule godoc
// @Summary      Jadwalkan sidang
// @Description  Menjadwalkan/menjadwal-ulang sidang dan menunjuk penguji (Admin + Kaprodi only)
// @Tags         Thesis Defense
// @Accept       json
// @Produce      json
// @Param        id path string true "Defense ID (UUID)"
// @Param        body body usecase.ScheduleDefenseRequest true "Jadwal + penguji"
// @Success      200  {object}  response.APIResponse{data=usecase.DefenseDetail} "Sidang terjadwal"
// @Failure      400  {object}  response.APIResponse "Jadwal tidak valid"
// @Failure      409  {object}  response.APIResponse "Konflik ruangan/penguji"
// @Security     BearerAuth
// @Router       /defenses/{id}/schedule [put]
func (h *DefenseHandler) Schedule(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	var req usecase.ScheduleDefenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: scheduled_at, room, dan examiner_ids wajib diisi")
		return
	}
	detail, err := h.defenseUseCase.Schedule(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		h.respondDefenseError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// SubmitScores godoc
// @Summary      Input nilai sidang
// @Description  Menginput nilai sidang per komponen (Dosen Penguji ditugaskan only)
// @Tags         Thesis Defense
// @Accept       json
// @Produce      json
// @Param        id path string true "Defense ID (UUID)"
// @Param        body body usecase.SubmitDefenseScoreRequest true "Nilai per komponen"
// @Success      200  {object}  response.APIResponse{data=usecase.DefenseResult} "Nilai tersimpan / hasil final"
// @Failure      403  {object}  response.APIResponse "Bukan penguji ditugaskan"
// @Failure      409  {object}  response.APIResponse "Sudah pernah input nilai"
// @Security     BearerAuth
// @Router       /defenses/{id}/scores [post]
func (h *DefenseHandler) SubmitScores(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	var req usecase.SubmitDefenseScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: scores wajib diisi")
		return
	}
	result, err := h.defenseUseCase.SubmitScores(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		h.respondDefenseError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

// Result godoc
// @Summary      Hasil sidang
// @Description  Mengambil hasil sidang lengkap (nilai akhir + grade + breakdown per penguji)
// @Tags         Thesis Defense
// @Produce      json
// @Param        id path string true "Defense ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.DefenseResult} "Hasil sidang"
// @Security     BearerAuth
// @Router       /defenses/{id}/result [get]
func (h *DefenseHandler) Result(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	result, err := h.defenseUseCase.Result(
		c.Request.Context(),
		id,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondDefenseError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

// revisionNotesRequest is the shared body for recording post-exam revision
// notes on seminars and defenses (Admin + Kaprodi only).
type revisionNotesRequest struct {
	RevisionNotes string `json:"revision_notes"`
}

// SetRevisionNotes godoc
// @Summary      Catat revisi sidang
// @Description  Mencatat catatan revisi pasca sidang (Admin + Kaprodi only)
// @Tags         Thesis Defense
// @Accept       json
// @Produce      json
// @Param        id path string true "Defense ID (UUID)"
// @Param        body body handler.revisionNotesRequest true "Catatan revisi"
// @Success      200  {object}  response.APIResponse{data=usecase.DefenseDetail} "Catatan tersimpan"
// @Failure      422  {object}  response.APIResponse "Sidang belum selesai dinilai"
// @Security     BearerAuth
// @Router       /defenses/{id}/revision [put]
func (h *DefenseHandler) SetRevisionNotes(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	var req revisionNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid")
		return
	}
	detail, err := h.defenseUseCase.SetRevisionNotes(c.Request.Context(), id, req.RevisionNotes, actorFromContext(c))
	if err != nil {
		h.respondDefenseError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Graduate godoc
// @Summary      Tetapkan yudisium
// @Description  Menetapkan thesis berstatus graduated (Kaprodi only)
// @Tags         Thesis Defense
// @Accept       json
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        body body usecase.GraduationRequest false "Catatan yudisium (opsional)"
// @Success      200  {object}  response.APIResponse "Thesis berhasil dinyatakan lulus"
// @Failure      422  {object}  response.APIResponse "Prasyarat yudisium belum terpenuhi"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/graduation [put]
func (h *DefenseHandler) Graduate(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	var req usecase.GraduationRequest
	_ = c.ShouldBindJSON(&req)

	err := h.defenseUseCase.Graduate(c.Request.Context(), thesisID, req, actorFromContext(c))
	if err != nil {
		h.respondDefenseError(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Thesis berhasil dinyatakan lulus"})
}

// Upcoming godoc
// @Summary      Jadwal mendatang
// @Description  Mengambil seminar dan sidang terjadwal dalam N hari ke depan (Admin + Kaprodi only)
// @Tags         Thesis Defense
// @Produce      json
// @Param        days query int false "Jumlah hari ke depan (default 14)"
// @Success      200  {object}  response.APIResponse "Jadwal seminar + sidang"
// @Security     BearerAuth
// @Router       /schedules/upcoming [get]
func (h *DefenseHandler) Upcoming(c *gin.Context) {
	days := 14
	if v := c.Query("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}
	schedules, err := h.defenseUseCase.Upcoming(c.Request.Context(), days)
	if err != nil {
		response.InternalError(c, "Gagal mengambil jadwal mendatang")
		return
	}
	response.Success(c, http.StatusOK, schedules)
}

// respondDefenseError delegates to the central error catalog (see errors.go).
func (h *DefenseHandler) respondDefenseError(c *gin.Context, err error) {
	respondError(c, err)
}
