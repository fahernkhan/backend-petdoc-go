package admin

import (
	"net/http"
	"petdoc/internal/infrastructure/response"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

// CreateAdmin godoc
// @Summary Create new admin
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body CreateAdminRequest true "User ID"
// @Success 201 {object} AdminResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin [post]
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Pakai error dari package admin tanpa prefix
		response.Error(c, http.StatusBadRequest, ErrInvalidRoleTransition, gin.H{
			"validation": err.Error(),
		})
		return
	}

	res, err := h.service.CreateAdmin(c.Request.Context(), req)
	if err != nil {
		adminErr := MapAdminError(err)
		response.Error(c, adminErr.Code, adminErr, gin.H{
			"details": adminErr.Details,
		})
		return
	}

	response.Success(c, http.StatusCreated, res)
}
