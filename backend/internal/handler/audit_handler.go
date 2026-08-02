package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type AuditHandler struct {
	auditUseCase *usecase.AuditUseCase
}

func NewAuditHandler(uc *usecase.AuditUseCase) *AuditHandler {
	return &AuditHandler{auditUseCase: uc}
}

// List godoc
// @Summary      Daftar audit log
// @Description  Mengambil daftar audit log dengan filter dan paginasi (Admin only)
// @Tags         Audit Log
// @Produce      json
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 50, max 200)"
// @Param        action query string false "Filter aksi (contoh: user.create)"
// @Param        entity_type query string false "Filter tipe entitas"
// @Param        user_id query string false "Filter user ID (UUID)"
// @Param        entity_id query string false "Filter entity ID (UUID)"
// @Param        date_from query string false "Filter tanggal mulai (RFC3339)"
// @Param        date_to query string false "Filter tanggal akhir (RFC3339)"
// @Success      200  {object}  response.APIResponse "Daftar audit log"
// @Failure      401  {object}  response.APIResponse "Belum autentikasi"
// @Failure      403  {object}  response.APIResponse "Akses ditolak"
// @Security     BearerAuth
// @Router       /admin/audit-logs [get]
func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}

	filter := domainRepo.AuditFilter{
		Action:     c.Query("action"),
		EntityType: c.Query("entity_type"),
		Page:       page,
		PerPage:    perPage,
	}
	if v := c.Query("user_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.UserID = &id
		}
	}
	if v := c.Query("entity_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.EntityID = &id
		}
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.DateFrom = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.DateTo = &t
		}
	}

	details, total, err := h.auditUseCase.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Gagal memuat audit log")
		return
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	response.SuccessWithMeta(c, details, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ByEntity godoc
// @Summary      Riwayat audit satu entitas
// @Description  Mengambil riwayat lengkap audit log untuk satu entitas (Admin + Kaprodi)
// @Tags         Audit Log
// @Produce      json
// @Param        entity_type path string true "Tipe entitas"
// @Param        entity_id path string true "Entity ID (UUID)"
// @Success      200  {object}  response.APIResponse "Riwayat entitas"
// @Failure      400  {object}  response.APIResponse "entity_id tidak valid"
// @Failure      403  {object}  response.APIResponse "Akses ditolak"
// @Security     BearerAuth
// @Router       /admin/audit-logs/entity/{entity_type}/{entity_id} [get]
func (h *AuditHandler) ByEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityID, err := uuid.Parse(c.Param("entity_id"))
	if err != nil {
		response.BadRequest(c, "entity_id tidak valid")
		return
	}

	details, err := h.auditUseCase.ByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		response.InternalError(c, "Gagal memuat riwayat entitas")
		return
	}
	response.Success(c, http.StatusOK, details)
}
