package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

// RefreshCookieName is the httpOnly cookie that carries the refresh token.
// Storing it in an HttpOnly cookie (instead of localStorage) keeps the
// long-lived token out of JavaScript's reach, mitigating XSS token theft.
const RefreshCookieName = "simtas_refresh_token"

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	cfg         *config.Config
	emailSvc    email.EmailService
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase, cfg *config.Config, emailSvc email.EmailService) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase, cfg: cfg, emailSvc: emailSvc}
}

// refreshCookieAttrs returns the cookie attributes for the refresh token. The
// cookie is always HttpOnly; it is marked Secure and SameSite=Strict in
// production, and SameSite=Lax in development (where it is served over HTTP).
func refreshCookieAttrs(cfg *config.Config) (secure bool, sameSite string) {
	if cfg.AppEnv == "production" {
		return true, "Strict"
	}
	return false, "Lax"
}

// setRefreshCookie writes the refresh token as an HttpOnly cookie.
func setRefreshCookie(c *gin.Context, cfg *config.Config, token string) {
	secure, sameSite := refreshCookieAttrs(cfg)
	c.SetCookie(
		RefreshCookieName,
		token,
		int(cfg.JWTRefreshExpiry.Seconds()),
		"/",
		"",      // domain: defaults to the API host
		secure, // Secure only over HTTPS (production)
		true,   // HttpOnly — not readable by JavaScript
	)
	switch sameSite {
	case "Strict":
		c.SetSameSite(http.SameSiteStrictMode)
	case "Lax":
		c.SetSameSite(http.SameSiteLaxMode)
	}
}

// clearRefreshCookie expires the refresh token cookie on logout.
func clearRefreshCookie(c *gin.Context, cfg *config.Config) {
	secure, sameSite := refreshCookieAttrs(cfg)
	c.SetCookie(
		RefreshCookieName,
		"",
		-1,
		"/",
		"",
		secure,
		true,
	)
	switch sameSite {
	case "Strict":
		c.SetSameSite(http.SameSiteStrictMode)
	case "Lax":
		c.SetSameSite(http.SameSiteLaxMode)
	}
}

// Login godoc
// @Summary      Login pengguna
// @Description  Autentikasi dengan email dan password, return JWT access + refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body usecase.LoginRequest true "Kredensial login"
// @Success      200  {object}  response.APIResponse{data=usecase.LoginResponse} "Login sukses"
// @Failure      401  {object}  response.APIResponse "Email atau password salah"
// @Failure      403  {object}  response.APIResponse "Akun terkunci"
// @Failure      429  {object}  response.APIResponse "Terlalu banyak percobaan"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req usecase.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	resp, err := h.authUseCase.Login(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		if errors.Is(err, usecase.ErrAccountLocked) {
			response.Error(c, http.StatusForbidden, err.Error(), err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, err.Error(), err)
			return
		}
		if err.Error() == "akun tidak aktif" {
			response.Error(c, http.StatusForbidden, "Akun tidak aktif", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", err)
		return
	}

	// Issue the refresh token as an HttpOnly cookie so it never reaches
	// JavaScript (mitigates XSS token theft). The JSON body therefore no
	// longer carries the refresh token — only the short-lived access token.
	setRefreshCookie(c, h.cfg, resp.RefreshToken)
	resp.RefreshToken = ""

	response.Success(c, http.StatusOK, resp)
}

// Logout godoc
// @Summary      Logout pengguna
// @Description  Menonaktifkan (blacklist) access token saat ini
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.APIResponse "Berhasil logout"
// @Failure      401  {object}  response.APIResponse "Token tidak ditemukan"
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("accessToken")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Token tidak ditemukan", nil)
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), token.(string), actorFromContext(c)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal logout", err)
		return
	}

	// Expire the refresh token cookie so a logged-out session cannot refresh.
	clearRefreshCookie(c, h.cfg)

	response.Success(c, http.StatusOK, gin.H{"message": "Berhasil logout"})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Menukar refresh token (HttpOnly cookie) dengan access token baru
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.APIResponse{data=usecase.RefreshTokenResponse} "Access token baru"
// @Failure      401  {object}  response.APIResponse "Refresh token tidak valid"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// The refresh token is read from the HttpOnly cookie set at login, not
	// from the request body — this keeps it out of JavaScript's reach.
	refreshToken, err := c.Cookie(RefreshCookieName)
	if err != nil || refreshToken == "" {
		response.Error(c, http.StatusUnauthorized, "Refresh token tidak ditemukan", nil)
		return
	}

	resp, err := h.authUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrRefreshTokenInvalid),
			errors.Is(err, usecase.ErrTokenBlacklisted):
			// On reuse/theft detection the family was revoked: expire the
			// cookie so the client cannot keep hammering the endpoint.
			clearRefreshCookie(c, h.cfg)
			response.Error(c, http.StatusUnauthorized, "Refresh token tidak valid", err)
		case errors.Is(err, usecase.ErrUserNotFound):
			response.Error(c, http.StatusUnauthorized, "User tidak ditemukan", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", err)
		}
		return
	}

	// Rotation produced a fresh refresh token: rotate the HttpOnly cookie so
	// the client always holds the current family member.
	if resp.RefreshToken != "" {
		setRefreshCookie(c, h.cfg, resp.RefreshToken)
	}
	resp.RefreshToken = ""

	response.Success(c, http.StatusOK, resp)
}

// GetMe godoc
// @Summary      Info user saat ini
// @Description  Mengambil data user yang sedang terautentikasi
// @Tags         Authentication
// @Produce      json
// @Success      200  {object}  response.APIResponse{data=usecase.UserDTO} "Data user"
// @Failure      401  {object}  response.APIResponse "Tidak terautentikasi"
// @Failure      404  {object}  response.APIResponse "User tidak ditemukan"
// @Security     BearerAuth
// @Router       /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User tidak terautentikasi", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User ID tidak valid", nil)
		return
	}

	user, err := h.authUseCase.GetMe(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User tidak ditemukan", err)
		return
	}

	response.Success(c, http.StatusOK, user)
}

// ForgotPassword godoc
// @Summary      Lupa password
// @Description  Mengirim tautan reset password ke email. Selalu return 200 untuk mencegah email enumeration
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body usecase.ForgotPasswordRequest true "Email user"
// @Success      200  {object}  response.APIResponse "Jika email terdaftar, tautan reset telah dikirim"
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req usecase.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	// Always return 200 to prevent email enumeration
	resetToken, fullName, _ := h.authUseCase.ForgotPassword(c.Request.Context(), req.Email)

	resp := gin.H{"message": "Jika email terdaftar, tautan reset telah dikirim"}
	// Development convenience: expose the reset link directly so the flow can
	// be tested end-to-end without a working email provider. Only exposed when
	// the email is actually registered (token non-empty) and never in production.
	if h.cfg.AppEnv != "production" && resetToken != "" {
		resp["reset_url"] = fmt.Sprintf("%s/reset-password?token=%s", h.cfg.FrontendURL, resetToken)
	}

	// Send the reset-link email (the sender never blocks the response; delivery
	// happens on the email worker). Silent failure: forgot-password always
	// returns 200 for anti-enumeration, so a broken email provider must not
	// change the response shape.
	if resetToken != "" {
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.cfg.FrontendURL, resetToken)
		_ = h.emailSvc.SendPasswordResetLink(c.Request.Context(), req.Email, fullName, resetURL)
	}
	response.Success(c, http.StatusOK, resp)
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Mengganti password menggunakan token dari email reset
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body usecase.ResetPasswordRequest true "Token + password baru"
// @Success      200  {object}  response.APIResponse "Password berhasil diubah"
// @Failure      400  {object}  response.APIResponse "Token tidak valid atau password tidak memenuhi syarat"
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req usecase.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	if err := h.authUseCase.ResetPassword(c.Request.Context(), req); err != nil {
		switch {
		case errors.Is(err, usecase.ErrPasswordMismatch):
			response.Error(c, http.StatusBadRequest, "Password tidak cocok", err)
		case errors.Is(err, usecase.ErrPasswordTooShort):
			response.Error(c, http.StatusBadRequest, "Password minimal 8 karakter", err)
		case errors.Is(err, usecase.ErrPasswordNotComplex):
			response.Error(c, http.StatusBadRequest, "Password harus mengandung minimal 1 huruf kapital dan 1 angka", err)
		case errors.Is(err, usecase.ErrInvalidResetToken):
			response.Error(c, http.StatusBadRequest, "Token tidak valid atau sudah kadaluarsa", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", err)
		}
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}
