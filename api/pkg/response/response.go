package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Page    int   `json:"page,omitempty"`
	PerPage int   `json:"per_page,omitempty"`
	Total   int64 `json:"total,omitempty"`
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

// Paginated responds with 200 and pagination metadata
func Paginated(c *gin.Context, data interface{}, page, perPage int, total int64) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:    page,
			PerPage: perPage,
			Total:   total,
		},
	})
}

// Error responds with the given HTTP status and an error message.
// The err argument is optional (used for logging context) and never sent to client.
func Error(c *gin.Context, statusCode int, message string, _ error) {
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
