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
	case errors.Is(err, ErrDoctorNotFound):
		return ErrorResponse{Code: http.StatusNotFound, Message: "Doctor not found"}
	case errors.Is(err, ErrInvalidDoctorData):
		return ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid doctor data"}
	case errors.Is(err, ErrInvalidWorkingHours):
		return ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid working hours"}
	case errors.Is(err, ErrInvalidWorkingDays):
		return ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid working days"}
	case errors.Is(err, ErrDatabaseOperation):
		return ErrorResponse{Code: http.StatusInternalServerError, Message: "Database error"}
	default:
		return ErrorResponse{Code: http.StatusInternalServerError, Message: "Internal server error"}
	}
}
