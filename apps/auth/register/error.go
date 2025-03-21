package register

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 1️⃣ Definisi error spesifik
var (
	ErrEmailExists           = errors.New("email already registered")
	ErrUsernameExists        = errors.New("username already taken")
	ErrRequiredFieldsMissing = errors.New("email, username, and password are required")
)

// 2️⃣ Fungsi untuk mendapatkan status code berdasarkan error
func GetStatusCode(err error) int {
	switch err {
	case ErrEmailExists, ErrUsernameExists:
		return http.StatusConflict // 409
	default:
		return http.StatusInternalServerError // 500
	}
}

// 3️⃣ Fungsi untuk membuat format response error yang konsisten
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
