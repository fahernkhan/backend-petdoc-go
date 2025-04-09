package doctor

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrDoctorNotFound        = errors.New("doctor not found")
	ErrInvalidDoctorData     = errors.New("invalid doctor data")
	ErrDatabaseOperation     = errors.New("database operation failed")
	ErrUnauthorizedAccess    = errors.New("unauthorized access")
	ErrInvalidWorkingDays    = errors.New("invalid working days format")
	ErrInvalidWorkingHours   = errors.New("invalid working hours format")
	ErrFailedUpdateRole      = errors.New("failed to update user role")
	ErrUserNotAllowed        = errors.New("user is not allowed to become doctor")
	ErrUserNotFound          = errors.New("user not found")
	ErrDuplicateEntry        = errors.New("duplicate entry")
	ErrInvalidRoleTransition = errors.New("invalid role transition")
	ErrImageUploadFailed     = errors.New("failed to upload image")
	ErrInvalidImageFormat    = errors.New("invalid image format")
	ErrFileTooLarge          = errors.New("file size exceeded 5MB")
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
		return ErrorResponse{http.StatusNotFound, "User not found", "The specified user does not exist"}
	case errors.Is(err, ErrUserNotAllowed):
		return ErrorResponse{http.StatusBadRequest, "Invalid role transition", "User must have 'user' role to become doctor"}
	case errors.Is(err, ErrDuplicateEntry):
		return ErrorResponse{http.StatusConflict, "Duplicate entry", "Doctor already exists for this user"}
	case errors.Is(err, ErrDatabaseOperation):
		return ErrorResponse{http.StatusInternalServerError, "Database error", "Unexpected database operation failure"}
	case errors.Is(err, ErrDoctorNotFound):
		return ErrorResponse{http.StatusNotFound, "Doctor not found", "The specified doctor does not exist"}
	case errors.Is(err, ErrInvalidRoleTransition):
		return ErrorResponse{http.StatusBadRequest, "Invalid role transition", "Cannot revert role for non-doctor user"}
	case errors.Is(err, ErrImageUploadFailed):
		return ErrorResponse{http.StatusInternalServerError, "Image upload failed", "Failed to upload profile image"}
	case errors.Is(err, ErrInvalidImageFormat):
		return ErrorResponse{http.StatusBadRequest, "Invalid image format", "Allowed formats: JPEG, PNG, WEBP, GIF"}
	default:
		return ErrorResponse{http.StatusInternalServerError, "Internal server error", "Unexpected error occurred"}
	}
}
