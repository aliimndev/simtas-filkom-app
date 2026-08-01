package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

// ListUsers handles GET /api/v1/admin/users
func (h *UserHandler) ListUsers(c *gin.Context) {
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

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			response.BadRequest(c, "Parameter is_active harus true atau false")
			return
		}
		isActive = &b
	}

	filter := repository.UserFilter{
		Role:         c.Query("role"),
		IsActive:     isActive,
		StudyProgram: c.Query("study_program"),
		Search:       c.Query("search"),
		Page:         page,
		PerPage:      perPage,
	}

	users, total, err := h.userUseCase.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar user")
		return
	}
	response.Paginated(c, users, page, perPage, total)
}

// GetUser handles GET /api/v1/admin/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	user, err := h.userUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			response.NotFound(c, "User tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal mengambil data user")
		return
	}
	response.Success(c, http.StatusOK, user)
}

// CreateUser handles POST /api/v1/admin/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req usecase.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: email, nama lengkap, dan role wajib diisi")
		return
	}

	actor := actorFromContext(c)
	user, err := h.userUseCase.Create(c.Request.Context(), req, actor)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyExists):
			response.Error(c, http.StatusConflict, "Email sudah terdaftar", err)
		case errors.Is(err, usecase.ErrRoleInvalid):
			response.BadRequest(c, "Role tidak valid")
		case errors.Is(err, usecase.ErrFullNameRequired),
			errors.Is(err, usecase.ErrInvalidEmailFormat):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "Gagal membuat user")
		}
		return
	}
	response.Created(c, user)
}

// UpdateUser handles PUT /api/v1/admin/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	var req usecase.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid")
		return
	}

	actor := actorFromContext(c)
	user, err := h.userUseCase.Update(c.Request.Context(), id, req, actor)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			response.NotFound(c, "User tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal memperbarui user")
		return
	}
	response.Success(c, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/admin/users/:id (soft delete)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	actor := actorFromContext(c)
	err := h.userUseCase.Delete(c.Request.Context(), id, actor)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrCannotDeleteSelf):
			response.BadRequest(c, "Tidak dapat menghapus akun sendiri")
		case errors.Is(err, usecase.ErrUserNotFound):
			response.NotFound(c, "User tidak ditemukan")
		default:
			response.InternalError(c, "Gagal menghapus user")
		}
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "User berhasil dihapus"})
}

// ActivateUser handles PATCH /api/v1/admin/users/:id/activate
func (h *UserHandler) ActivateUser(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	actor := actorFromContext(c)
	err := h.userUseCase.Activate(c.Request.Context(), id, actor)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			response.NotFound(c, "User tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal mengaktifkan user")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "User berhasil diaktifkan"})
}

// DeactivateUser handles PATCH /api/v1/admin/users/:id/deactivate
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	actor := actorFromContext(c)
	err := h.userUseCase.Deactivate(c.Request.Context(), id, actor)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrCannotDeactivateSelf):
			response.BadRequest(c, "Tidak dapat menonaktifkan akun sendiri")
		case errors.Is(err, usecase.ErrUserNotFound):
			response.NotFound(c, "User tidak ditemukan")
		default:
			response.InternalError(c, "Gagal menonaktifkan user")
		}
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "User berhasil dinonaktifkan"})
}

// ResetPassword handles POST /api/v1/admin/users/:id/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	actor := actorFromContext(c)
	err := h.userUseCase.ResetPassword(c.Request.Context(), id, actor)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			response.NotFound(c, "User tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal mereset password")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Password berhasil direset dan dikirim ke email user"})
}

// ImportTemplate handles GET /api/v1/admin/users/import-template
func (h *UserHandler) ImportTemplate(c *gin.Context) {
	data, err := h.userUseCase.BuildImportTemplate()
	if err != nil {
		response.InternalError(c, "Gagal membuat template import")
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="user_import_template.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ImportUsers handles POST /api/v1/admin/users/import
func (h *UserHandler) ImportUsers(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File wajib diunggah pada field 'file'")
		return
	}

	if fileHeader.Size > usecase.MaxImportFileSize {
		response.BadRequest(c, "Ukuran file melebihi 5 MB")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c, "Gagal membaca file")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.InternalError(c, "Gagal membaca file")
		return
	}

	actor := actorFromContext(c)
	result, err := h.userUseCase.ImportUsers(c.Request.Context(), fileHeader.Filename, data, actor)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidFileFormat):
			response.BadRequest(c, "Format file tidak didukung, gunakan .csv atau .xlsx")
		case errors.Is(err, usecase.ErrFileTooLarge):
			response.BadRequest(c, "Ukuran file melebihi 5 MB")
		default:
			response.InternalError(c, "Gagal mengimpor user")
		}
		return
	}
	response.Success(c, http.StatusOK, result)
}

func parseUUIDParam(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return uuid.Nil, false
	}
	return id, true
}

func actorFromContext(c *gin.Context) usecase.Actor {
	userID := uuid.Nil
	if v, ok := c.Get("userID"); ok {
		if s, ok := v.(string); ok {
			userID, _ = uuid.Parse(s)
		}
	}

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	return usecase.Actor{
		UserID:    userID,
		IPAddress: &ip,
		UserAgent: &ua,
	}
}
