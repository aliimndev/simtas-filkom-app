package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a typed HTTP error carrying an HTTP status code and a
// user-facing message. The wrapped Err is only used for logging.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// Standard application errors (Job 13).
var (
	ErrNotFound      = &AppError{Code: http.StatusNotFound, Message: "Data tidak ditemukan"}
	ErrUnauthorized  = &AppError{Code: http.StatusUnauthorized, Message: "Akses tidak diizinkan"}
	ErrForbidden     = &AppError{Code: http.StatusForbidden, Message: "Akses ditolak"}
	ErrBadRequest    = &AppError{Code: http.StatusBadRequest, Message: "Request tidak valid"}
	ErrConflict      = &AppError{Code: http.StatusConflict, Message: "Data sudah ada"}
	ErrUnprocessable = &AppError{Code: http.StatusUnprocessableEntity, Message: "Proses tidak dapat dilakukan"}
	ErrInternal      = &AppError{Code: http.StatusInternalServerError, Message: "Terjadi kesalahan pada server"}
)

// New wraps err in an AppError with the given code and message.
func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// CodeOf extracts the HTTP status code from err, defaulting to 500.
func CodeOf(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// MessageOf extracts a user-facing message from err.
func MessageOf(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "Terjadi kesalahan pada server"
}
