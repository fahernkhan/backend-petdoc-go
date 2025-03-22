package login

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrAccountNotActive   = errors.New("account not active")
)

func GetStatusCode(err error) int {
	switch err {
	case ErrInvalidCredentials, ErrUserNotFound:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func ErrorResponse(message string, err error) gin.H {
	resp := gin.H{
		"success": false,
		"error":   message,
	}
	// Hanya tambahkan "details" jika err tidak nil
	if err != nil {
		resp["details"] = err.Error()
	}
	return resp
}
