package user

import (
	"errors"
	"log/slog"
	"net/http"
	"petdoc/internal/infrastructure/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	const op = "user.handler.GetAllUsers"
	logger := slog.With("operation", op)

	// Bind query parameters
	var req PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Warn("Invalid request parameters", "error", err)
		response.Error(c, http.StatusBadRequest, ErrInvalidPaginationParams, gin.H{
			"details": err.Error(),
		})
		return
	}

	// Panggil service
	result, err := h.service.GetPaginatedUsers(c.Request.Context(), req)
	if err != nil {
		var validationErr *ValidationError

		if errors.As(err, &validationErr) {
			// Jika error adalah ValidationError
			logger.Warn("Validation error", "error", err)
			response.Error(c, http.StatusBadRequest, err, gin.H{
				"field":   validationErr.Field,
				"message": validationErr.Message,
			})
			return
		}

		if errors.Is(err, ErrPageOutOfRange) {
			// Jika error karena halaman melebihi batas
			logger.Warn("Page out of range", "requested_page", req.Page)
			response.Error(c, http.StatusBadRequest, err, gin.H{
				"message": "Page out of range",
			})
			return
		}

		if errors.Is(err, ErrDatabaseQuery) {
			// Jika error berasal dari database
			logger.Error("Database error", "error", err)
			response.Error(c, http.StatusInternalServerError, err, nil)
			return
		}

		// Error lainnya
		logger.Error("Unexpected error", "error", err)
		response.Error(c, http.StatusInternalServerError, errors.New("internal server error"), nil)
		return
	}

	// Jika tidak ada error, kirim response sukses
	response.Success(c, http.StatusOK, result)
}
