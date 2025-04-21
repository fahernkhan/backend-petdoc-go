package admin

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidRoleTransition = errors.New("invalid role transition")
	ErrDatabaseOperation     = errors.New("database operation failed")
)

type AdminError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e AdminError) Error() string {
	return fmt.Sprintf("admin error [%d]: %s", e.Code, e.Message)
}

func MapAdminError(err error) AdminError {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return AdminError{http.StatusNotFound, "User not found", "User dengan ID tersebut tidak ditemukan"}
	case errors.Is(err, ErrInvalidRoleTransition):
		return AdminError{http.StatusBadRequest, "Invalid role transition", "Hanya user dengan role 'user' yang bisa diubah menjadi admin"}
	case errors.Is(err, ErrDatabaseOperation):
		return AdminError{http.StatusInternalServerError, "Database error", "Terjadi kesalahan pada operasi database"}
	default:
		return AdminError{http.StatusInternalServerError, "Internal server error", "Terjadi kesalahan yang tidak diketahui"}
	}
}
