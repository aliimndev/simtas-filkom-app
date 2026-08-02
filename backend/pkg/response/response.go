package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/apperror"
)

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Success responds with HTTP 200 (or the given statusCode) and wraps data in APIResponse.
// Callers that previously called Success(c, message, data) now pass statusCode explicitly.
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Created responds with HTTP 201
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta responds with 200 and both data and explicit pagination metadata.
// Used when the data shape differs from a plain list (e.g. list + summary).
func SuccessWithMeta(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

// Paginated responds with 200 and pagination metadata including total_pages
func Paginated(c *gin.Context, data interface{}, page, perPage int, total int64) {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error responds with the given HTTP status and an error message.
// When err is an *apperror.AppError its Code/Message take precedence, which
// keeps every handler's error output consistent (Job 13). The err argument is
// optional and never sent to the client.
func Error(c *gin.Context, statusCode int, message string, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		statusCode = apperror.CodeOf(err)
		message = apperror.MessageOf(err)
	}
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message, nil)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, nil)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, nil)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, nil)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, nil)
}
