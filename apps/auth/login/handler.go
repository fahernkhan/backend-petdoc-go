package login

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	logger := slog.With("module", "login-handler")

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := strings.Split(err.Error(), "\n") // Pisahkan error jika lebih dari satu
		logger.Warn("Validation error", "error", validationErrors)
		c.JSON(http.StatusBadRequest, ErrorResponse("Validation error, Invalid request format", errors.New(strings.Join(validationErrors, ", "))))
		return
	}
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	logger.Warn("Invalid request format", "error", err)
	// 	c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request format", err))
	// 	return
	// }

	// Validasi input tidak boleh kosong dipakai jika di bagian model tidak di binding
	// if strings.TrimSpace(req.Email) == "" {
	// 	logger.Warn("Email is required")
	// 	c.JSON(http.StatusBadRequest, ErrorResponse("Email is required", nil))
	// 	return
	// }
	// if strings.TrimSpace(req.Password) == "" {
	// 	logger.Warn("Password is required")
	// 	c.JSON(http.StatusBadRequest, ErrorResponse("Password is required", nil))
	// 	return
	// }

	response, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		statusCode := GetStatusCode(err)
		logger.Error("Login failed",
			"error", err,
			"status_code", statusCode,
		)
		c.JSON(statusCode, ErrorResponse("Login failed", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"details": "Login successfully",
	})
}
