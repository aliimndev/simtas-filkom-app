package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type NotificationHandler struct {
	notificationUseCase *usecase.NotificationUseCase
}

func NewNotificationHandler(uc *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{notificationUseCase: uc}
}

// List godoc
// @Summary      Daftar notifikasi
// @Description  Mengambil notifikasi pengguna yang sedang login (terbaru dahulu)
// @Tags         Notifikasi
// @Produce      json
// @Param        limit query int false "Maksimal item (default 20, max 100)"
// @Success      200  {object}  response.APIResponse "Daftar notifikasi"
// @Failure      401  {object}  response.APIResponse "Belum autentikasi"
// @Security     BearerAuth
// @Router       /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID := userIDFromContext(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	notifs, err := h.notificationUseCase.List(c.Request.Context(), userID, limit)
	if err != nil {
		response.InternalError(c, "Gagal memuat notifikasi")
		return
	}
	response.Success(c, http.StatusOK, notifs)
}

// UnreadCount godoc
// @Summary      Jumlah notifikasi belum dibaca
// @Description  Mengambil jumlah notifikasi yang belum dibaca pengguna
// @Tags         Notifikasi
// @Produce      json
// @Success      200  {object}  response.APIResponse "Jumlah notifikasi belum dibaca"
// @Failure      401  {object}  response.APIResponse "Belum autentikasi"
// @Security     BearerAuth
// @Router       /notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := userIDFromContext(c)

	count, err := h.notificationUseCase.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Gagal memuat jumlah notifikasi")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"unread_count": count})
}

// MarkRead godoc
// @Summary      Tandai notifikasi sudah dibaca
// @Description  Menandai satu notifikasi milik pengguna sebagai sudah dibaca
// @Tags         Notifikasi
// @Produce      json
// @Param        id path string true "ID notifikasi"
// @Success      200  {object}  response.APIResponse "Berhasil ditandai"
// @Failure      401  {object}  response.APIResponse "Belum autentikasi"
// @Failure      404  {object}  response.APIResponse "Notifikasi tidak ditemukan"
// @Security     BearerAuth
// @Router       /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := userIDFromContext(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "id tidak valid")
		return
	}

	if err := h.notificationUseCase.MarkRead(c.Request.Context(), userID, id); err != nil {
		if err == usecase.ErrNotificationNotFound {
			response.Error(c, http.StatusNotFound, "Notifikasi tidak ditemukan", nil)
			return
		}
		response.InternalError(c, "Gagal menandai notifikasi")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"ok": true})
}

// MarkAllRead godoc
// @Summary      Tandai semua notifikasi sudah dibaca
// @Description  Menandai semua notifikasi belum dibaca pengguna sebagai sudah dibaca
// @Tags         Notifikasi
// @Produce      json
// @Success      200  {object}  response.APIResponse "Berhasil ditandai"
// @Failure      401  {object}  response.APIResponse "Belum autentikasi"
// @Security     BearerAuth
// @Router       /notifications/read-all [patch]
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := userIDFromContext(c)

	if err := h.notificationUseCase.MarkAllRead(c.Request.Context(), userID); err != nil {
		response.InternalError(c, "Gagal menandai semua notifikasi")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"ok": true})
}
