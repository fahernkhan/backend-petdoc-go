package user

import "errors"

var (
	ErrInvalidPaginationParams = errors.New("invalid pagination parameters")
	ErrDatabaseQuery           = errors.New("failed to query users")
	ErrDatabaseCount           = errors.New("failed to count users")
	ErrPageOutOfRange          = errors.New("requested page is out of range")
)

type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return "database error: " + e.Err.Error()
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return "validation failed for field '" + e.Field + "'"
}
