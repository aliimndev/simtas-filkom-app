package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/api/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/api/pkg/response"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req usecase.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	resp, err := h.authUseCase.Login(c.Request.Context(), req)
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

	response.Success(c, http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("accessToken")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Token tidak ditemukan", nil)
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), token.(string)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal logout", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Berhasil logout"})
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req usecase.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	resp, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrRefreshTokenInvalid),
			errors.Is(err, usecase.ErrTokenBlacklisted):
			response.Error(c, http.StatusUnauthorized, "Refresh token tidak valid", err)
		case errors.Is(err, usecase.ErrUserNotFound):
			response.Error(c, http.StatusUnauthorized, "User tidak ditemukan", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// GetMe handles GET /api/v1/auth/me
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

// ForgotPassword handles POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req usecase.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	// Always return 200 to prevent email enumeration
	_, _ = h.authUseCase.ForgotPassword(c.Request.Context(), req.Email)

	// TODO: Send email with reset link when email service is implemented (Job 08)
	response.Success(c, http.StatusOK, gin.H{
		"message": "Jika email terdaftar, tautan reset telah dikirim",
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password
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
