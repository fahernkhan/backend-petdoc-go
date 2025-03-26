package doctor

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrDoctorNotFound      = errors.New("doctor not found")
	ErrInvalidDoctorData   = errors.New("invalid doctor data")
	ErrDatabaseOperation   = errors.New("database operation failed")
	ErrUnauthorizedAccess  = errors.New("unauthorized access")
	ErrInvalidWorkingDays  = errors.New("invalid working days format")
	ErrInvalidWorkingHours = errors.New("invalid working hours format")
	ErrFailedUpdateRole    = errors.New("failed to update user role")
	ErrUserNotAllowed      = errors.New("user is not allowed to become doctor")
	ErrUserNotFound        = errors.New("user not found")
	ErrDuplicateEntry      = errors.New("duplicate entry")
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func MapError(err error) ErrorResponse {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "User not found",
			Details: "The specified user does not exist",
		}
	case errors.Is(err, ErrUserNotAllowed):
		return ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid role transition",
			Details: "User must have 'user' role to become doctor",
		}
	case errors.Is(err, ErrDuplicateEntry):
		return ErrorResponse{
			Code:    http.StatusConflict,
			Message: "Duplicate entry",
			Details: "Doctor already exists for this user",
		}
	case errors.Is(err, ErrDatabaseOperation):
		return ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Database error",
			Details: "Unexpected database operation failure",
		}
	default:
		return ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Details: "Unexpected error occurred",
		}
	}
}
