package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
)

type DocumentHandler struct {
	documentUseCase *usecase.DocumentUseCase
}

func NewDocumentHandler(uc *usecase.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{documentUseCase: uc}
}

// Upload godoc
// @Summary      Upload dokumen
// @Description  Mengunggah dokumen PDF (Mahasiswa pemilik only)
// @Tags         Documents
// @Accept       multipart/form-data
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        file formData file true "File PDF (max 10 MB)"
// @Param        document_type formData string true "Tipe dokumen (proposal, draft_chapter, seminar_doc, defense_doc, final_thesis, revision_sheet, endorsement_letter)"
// @Param        chapter_number formData int false "Nomor bab 1-5 (wajib untuk draft_chapter)"
// @Success      201  {object}  response.APIResponse{data=usecase.DocumentDetail} "Dokumen diunggah"
// @Failure      400  {object}  response.APIResponse "File tidak valid"
// @Failure      403  {object}  response.APIResponse "Bukan pemilik thesis"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/documents [post]
func (h *DocumentHandler) Upload(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File wajib diunggah pada field 'file'")
		return
	}
	defer file.Close()

	documentType := c.PostForm("document_type")

	var chapterNum *int
	if v := c.PostForm("chapter_number"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil {
			response.BadRequest(c, "chapter_number harus berupa angka")
			return
		}
		chapterNum = &n
	}

	detail, err := h.documentUseCase.Upload(
		c.Request.Context(),
		thesisID,
		file,
		header,
		usecase.UploadDocumentRequest{
			DocumentType:  documentType,
			ChapterNumber: chapterNum,
		},
		actorFromContext(c),
	)
	if err != nil {
		h.respondDocumentError(c, err)
		return
	}
	response.Created(c, detail)
}

// List godoc
// @Summary      Daftar dokumen
// @Description  Mengambil daftar dokumen aktif per tipe (pemilik/pembimbing/penguji/admin/kaprodi)
// @Tags         Documents
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        document_type query string false "Filter tipe dokumen"
// @Param        status query string false "Filter status"
// @Param        page query int false "Halaman (default 1)"
// @Param        per_page query int false "Per halaman (default 20)"
// @Success      200  {object}  response.APIResponse "Daftar dokumen"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/documents [get]
func (h *DocumentHandler) List(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	page, perPage := parsePagination(c)
	filter := repository.DocumentFilter{
		DocumentType: c.Query("document_type"),
		Status:       c.Query("status"),
		Page:         page,
		PerPage:      perPage,
	}

	docs, total, err := h.documentUseCase.List(
		c.Request.Context(),
		thesisID,
		filter,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondDocumentError(c, err)
		return
	}
	response.Paginated(c, docs, page, perPage, total)
}

// Get godoc
// @Summary      Detail dokumen
// @Description  Mengambil detail satu dokumen
// @Tags         Documents
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        id path string true "Document ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.DocumentDetail} "Detail dokumen"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/documents/{id} [get]
func (h *DocumentHandler) Get(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	detail, err := h.documentUseCase.GetByID(
		c.Request.Context(),
		thesisID,
		id,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondDocumentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// Download godoc
// @Summary      Download dokumen
// @Description  Menghasilkan presigned URL untuk download dokumen (expired 15 menit)
// @Tags         Documents
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        id path string true "Document ID (UUID)"
// @Success      200  {object}  response.APIResponse{data=usecase.DownloadResult} "Presigned URL"
// @Failure      404  {object}  response.APIResponse "Tidak ditemukan"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/documents/{id}/download [get]
func (h *DocumentHandler) Download(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	result, err := h.documentUseCase.Download(
		c.Request.Context(),
		thesisID,
		id,
		userIDFromContext(c),
		roleFromContext(c),
		actorFromContext(c),
	)
	if err != nil {
		h.respondDocumentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

// History godoc
// @Summary      Riwayat versi dokumen
// @Description  Mengambil semua versi dokumen untuk satu tipe (riwayat upload)
// @Tags         Documents
// @Produce      json
// @Param        thesis_id path string true "Thesis ID (UUID)"
// @Param        document_type query string true "Tipe dokumen"
// @Param        chapter_number query int false "Nomor bab (untuk draft_chapter)"
// @Success      200  {object}  response.APIResponse "Riwayat versi"
// @Security     BearerAuth
// @Router       /theses/{thesis_id}/documents/history [get]
func (h *DocumentHandler) History(c *gin.Context) {
	thesisID, ok := parseThesisIDParam(c)
	if !ok {
		return
	}

	docType := c.Query("document_type")
	if !entity.ValidDocumentType(docType) {
		response.BadRequest(c, "document_type tidak valid")
		return
	}

	var chapterNum *int
	if v := c.Query("chapter_number"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil {
			response.BadRequest(c, "chapter_number harus berupa angka")
			return
		}
		chapterNum = &n
	}

	history, err := h.documentUseCase.History(
		c.Request.Context(),
		thesisID,
		docType,
		chapterNum,
		userIDFromContext(c),
		roleFromContext(c),
	)
	if err != nil {
		h.respondDocumentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, history)
}

// documentReviewRequest is the body for a supervisor's document review decision.
type documentReviewRequest struct {
	Decision string `json:"decision" binding:"required"`
	Notes    string `json:"notes"`
}

// Review godoc
// @Summary      Review dokumen
// @Description  Menyetujui atau meminta revisi dokumen (Dosen Pembimbing thesis terkait only)
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Param        id path string true "Document ID (UUID)"
// @Param        body body handler.documentReviewRequest true "Keputusan review"
// @Success      200  {object}  response.APIResponse{data=usecase.DocumentDetail} "Hasil review"
// @Failure      403  {object}  response.APIResponse "Bukan pembimbing thesis ini"
// @Failure      422  {object}  response.APIResponse "Dokumen tidak pending"
// @Security     BearerAuth
// @Router       /documents/{id}/review [patch]
func (h *DocumentHandler) Review(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}

	var req documentReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request tidak valid: decision wajib diisi")
		return
	}

	detail, err := h.documentUseCase.Review(c.Request.Context(), id, req.Decision, req.Notes, actorFromContext(c))
	if err != nil {
		h.respondDocumentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, detail)
}

// parsePagination reads page/per_page with safe defaults (shared helper).
func parsePagination(c *gin.Context) (int, int) {
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
	return page, perPage
}

// respondDocumentError maps document use case errors to HTTP responses.
func (h *DocumentHandler) respondDocumentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrThesisNotFound),
		errors.Is(err, usecase.ErrDocumentNotFound):
		response.NotFound(c, "Dokumen atau thesis tidak ditemukan")
	case errors.Is(err, usecase.ErrForbidden),
		errors.Is(err, usecase.ErrNotDocumentReviewer):
		response.Forbidden(c, "Akses ditolak")
	case errors.Is(err, utils.ErrNotPDF):
		response.BadRequest(c, "Hanya file PDF yang diizinkan")
	case errors.Is(err, utils.ErrFileTooLarge):
		response.BadRequest(c, "Ukuran file maksimal 10 MB")
	case errors.Is(err, utils.ErrFileEmpty),
		errors.Is(err, usecase.ErrDocumentFileRequired):
		response.BadRequest(c, "File wajib diunggah")
	case errors.Is(err, usecase.ErrInvalidDocumentType),
		errors.Is(err, usecase.ErrInvalidChapterNumber),
		errors.Is(err, usecase.ErrChapterNumberRequired),
		errors.Is(err, usecase.ErrDocumentThesisNotLive):
		response.BadRequest(c, err.Error())
	case errors.Is(err, usecase.ErrInvalidReviewDecision):
		response.BadRequest(c, "Decision harus approved atau revision_required")
	case errors.Is(err, usecase.ErrDocumentNotPending):
		response.Error(c, http.StatusUnprocessableEntity, "Dokumen tidak dalam status menunggu review", err)
	default:
		response.InternalError(c, "Terjadi kesalahan server")
	}
}
