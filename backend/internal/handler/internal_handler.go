package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

// testEmailSender is implemented by email services that support sending a
// diagnostic test email (ResendEmailService).
type testEmailSender interface {
	SendTestEmail(ctx context.Context, to string) error
}

// InternalHandler exposes development-only endpoints (e.g. email testing).
type InternalHandler struct {
	emailSvc email.EmailService
}

func NewInternalHandler(emailSvc email.EmailService) *InternalHandler {
	return &InternalHandler{emailSvc: emailSvc}
}

type testEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// TestEmail godoc
// @Summary      Kirim email test
// @Description  Mengirim email diagnostik untuk verifikasi konfigurasi email (development only)
// @Tags         Internal
// @Accept       json
// @Produce      json
// @Param        body body handler.testEmailRequest true "Alamat email tujuan"
// @Success      200  {object}  response.APIResponse "Email test diproses"
// @Failure      400  {object}  response.APIResponse "Email tidak valid"
// @Failure      500  {object}  response.APIResponse "Gagal mengirim email test"
// @Router       /internal/test-email [post]
func (h *InternalHandler) TestEmail(c *gin.Context) {
	var req testEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Field 'email' wajib diisi dan harus berupa alamat email valid")
		return
	}

	sender, ok := h.emailSvc.(testEmailSender)
	if !ok {
		response.InternalError(c, "Email service tidak mendukung pengiriman test email")
		return
	}

	if err := sender.SendTestEmail(c.Request.Context(), req.Email); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengirim email test", err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Email test berhasil diproses (cek console / email_logs)"})
}
