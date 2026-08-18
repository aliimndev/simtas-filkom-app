package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

// UserResponse is the admin-facing representation of a user. It flattens the
// role relation into a plain string (role name) so the payload is consistent
// with the rest of the API (e.g. /auth/me and login return role as a string).
type UserResponse struct {
	ID                 uuid.UUID  `json:"id"`
	Email              string     `json:"email"`
	FullName           string     `json:"full_name"`
	Role               string     `json:"role"`
	NimNidn            *string    `json:"nim_nidn,omitempty"`
	StudyProgram       *string    `json:"study_program,omitempty"`
	PlaceOfBirth       *string    `json:"place_of_birth,omitempty"`
	Address            *string    `json:"address,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	BirthDate          *string    `json:"birth_date,omitempty"`
	Faculty            *string    `json:"faculty,omitempty"`
	Semester           *int       `json:"semester,omitempty"`
	ProfilePhotoURL    *string    `json:"profile_photo_url,omitempty"`
	IsActive           bool       `json:"is_active"`
	MustChangePassword bool       `json:"must_change_password"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func toUserResponse(u *entity.User) UserResponse {
	return UserResponse{
		ID:                 u.ID,
		Email:              u.Email,
		FullName:           u.FullName,
		Role:               u.Role.Name,
		NimNidn:            u.NimNidn,
		StudyProgram:       u.StudyProgram,
		PlaceOfBirth:       u.PlaceOfBirth,
		Address:            u.Address,
		Phone:              u.Phone,
		BirthDate:          u.BirthDate,
		Faculty:            u.Faculty,
		Semester:           u.Semester,
		ProfilePhotoURL:    u.ProfilePhotoURL,
		IsActive:           u.IsActive,
		MustChangePassword: u.MustChangePassword,
		LastLoginAt:        u.LastLoginAt,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}
}

func toUserResponses(users []*entity.User) []UserResponse {
	out := make([]UserResponse, 0, len(users))
	for _, u := range users {
		if u == nil {
			continue
		}
		out = append(out, toUserResponse(u))
	}
	return out
}

func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

// ListUsers godoc
// @Summary      Daftar user
// @Description  Mengambil daftar user dengan filter dan pagination (Admin only)
// @Tags         User Management
// @Produce      json
// @Param        role query string false "Filter role"
// @Param        is_active query boolean false "Filter status aktif"
// @Param        study_program query string false "Filter program studi"
// @Param        search query string false "Cari nama/email/NIM"
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 20, max 100)"
// @Success      200  {object}  response.APIResponse "Daftar user"
// @Failure      400  {object}  response.APIResponse "Parameter tidak valid"
// @Security     BearerAuth
// @Router       /admin/users [get]
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
	response.Paginated(c, toUserResponses(users), page, perPage, total)
}

// GetUser godoc
// @Summary      Detail user
// @Description  Mengambil detail user berdasarkan ID (Admin only)
// @Tags         User Management
// @Produce      json
// @Param        id path string true "User ID (UUID)"
// @Success      200  {object}  response.APIResponse "Detail user"
// @Failure      404  {object}  response.APIResponse "User tidak ditemukan"
// @Security     BearerAuth
// @Router       /admin/users/{id} [get]
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
	response.Success(c, http.StatusOK, toUserResponse(user))
}

// CreateUser godoc
// @Summary      Buat user
// @Description  Membuat user baru (password otomatis dikirim via email) (Admin only)
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        body body usecase.CreateUserRequest true "Data user"
// @Success      201  {object}  response.APIResponse "User dibuat"
// @Failure      400  {object}  response.APIResponse "Request tidak valid"
// @Failure      409  {object}  response.APIResponse "Email sudah terdaftar"
// @Security     BearerAuth
// @Router       /admin/users [post]
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
	response.Created(c, toUserResponse(user))
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Memperbarui data user (email dan role tidak dapat diubah) (Admin only)
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID (UUID)"
// @Param        body body usecase.UpdateUserRequest true "Data yang diperbarui"
// @Success      200  {object}  response.APIResponse "User diperbarui"
// @Failure      404  {object}  response.APIResponse "User tidak ditemukan"
// @Security     BearerAuth
// @Router       /admin/users/{id} [put]
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
	response.Success(c, http.StatusOK, toUserResponse(user))
}

// UpdateMyProfile godoc
// @Summary      Update profil sendiri
// @Description  Memperbarui data profil pengguna yang sedang login (Nama, NIM, jurusan, dll). Email & role tidak dapat diubah.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        body body usecase.UpdateUserRequest true "Data profil"
// @Success      200  {object}  response.APIResponse "Profil diperbarui"
// @Security     BearerAuth
// @Router       /users/me [patch]
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
	actor := actorFromContext(c)
	if actor.UserID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "Sesi tidak valid", nil)
		return
	}

	var req usecase.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid")
		return
	}

	user, err := h.userUseCase.Update(c.Request.Context(), actor.UserID, req, actor)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			response.NotFound(c, "User tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal memperbarui profil")
		return
	}
	response.Success(c, http.StatusOK, toUserResponse(user))
}
// @Summary      Hapus user (soft delete)
// @Description  Menghapus user secara soft delete (Admin only)
// @Tags         User Management
// @Produce      json
// @Param        id path string true "User ID (UUID)"
// @Success      200  {object}  response.APIResponse "User berhasil dihapus"
// @Failure      400  {object}  response.APIResponse "Tidak dapat menghapus akun sendiri"
// @Failure      404  {object}  response.APIResponse "User tidak ditemukan"
// @Security     BearerAuth
// @Router       /admin/users/{id} [delete]
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

// ActivateUser godoc
// @Summary      Aktifkan user
// @Description  Mengaktifkan akun user (Admin only)
// @Tags         User Management
// @Produce      json
// @Param        id path string true "User ID (UUID)"
// @Success      200  {object}  response.APIResponse "User berhasil diaktifkan"
// @Failure      404  {object}  response.APIResponse "User tidak ditemukan"
// @Security     BearerAuth
// @Router       /admin/users/{id}/activate [patch]
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

// DeactivateUser godoc
// @Summary      Nonaktifkan user
// @Description  Menonaktifkan akun user dan invalidasi session (Admin only)
// @Tags         User Management
// @Produce      json
// @Param        id path string true "User ID (UUID)"
// @Success      200  {object}  response.APIResponse "User berhasil dinonaktifkan"
// @Failure      400  {object}  response.APIResponse "Tidak dapat menonaktifkan akun sendiri"
// @Failure      404  {object}  response.APIResponse "User tidak ditemukan"
// @Security     BearerAuth
// @Router       /admin/users/{id}/deactivate [patch]
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

// ResetPassword godoc
// @Summary      Reset password user
// @Description  Men-generate password baru dan mengirim via email (Admin only)
// @Tags         User Management
// @Produce      json
// @Param        id path string true "User ID (UUID)"
// @Success      200  {object}  response.APIResponse "Password berhasil direset"
// @Failure      404  {object}  response.APIResponse "User tidak ditemukan"
// @Security     BearerAuth
// @Router       /admin/users/{id}/reset-password [post]
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

// ChangeMyPassword godoc
// @Summary      Ganti password sendiri
// @Description  Mengganti password user yang sedang login
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        body body usecase.ChangePasswordRequest true "Password lama + baru"
// @Success      200  {object}  response.APIResponse "Password berhasil diubah"
// @Failure      400  {object}  response.APIResponse "Password saat ini salah / tidak memenuhi syarat"
// @Security     BearerAuth
// @Router       /users/me/password [put]
func (h *UserHandler) ChangeMyPassword(c *gin.Context) {
	var req usecase.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: current_password dan new_password wajib diisi")
		return
	}

	actor := actorFromContext(c)
	if actor.UserID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "Sesi tidak valid", nil)
		return
	}

	err := h.userUseCase.ChangeMyPassword(c.Request.Context(), actor.UserID, req, actor)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			response.NotFound(c, "User tidak ditemukan")
		case errors.Is(err, usecase.ErrPasswordMismatch):
			response.Error(c, http.StatusBadRequest, "Password saat ini salah", err)
		case errors.Is(err, usecase.ErrPasswordTooShort),
			errors.Is(err, usecase.ErrPasswordNotComplex):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "Gagal mengganti password")
		}
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}

// ImportTemplate godoc
// @Summary      Download template import user
// @Description  Mendownload template Excel untuk import user massal (Admin only)
// @Tags         User Management
// @Produce      octet-stream
// @Success      200  {file}  binary "Template Excel"
// @Security     BearerAuth
// @Router       /admin/users/import-template [get]
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

// ImportUsers godoc
// @Summary      Import user massal
// @Description  Import user dari file .csv atau .xlsx (Admin only)
// @Tags         User Management
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File CSV/Excel (max 5 MB)"
// @Success      200  {object}  response.APIResponse{data=usecase.ImportResult} "Hasil import"
// @Failure      400  {object}  response.APIResponse "File tidak valid"
// @Security     BearerAuth
// @Router       /admin/users/import [post]
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
