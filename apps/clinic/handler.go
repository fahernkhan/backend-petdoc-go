package clinic

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateClinic godoc
// @Summary Create new clinic
// @Tags Clinics
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param name formData string true "Clinic name"
// @Param phone_number formData string true "Clinic phone number"
// @Param address formData string true "Clinic address"
// @Param map_link formData string false "Map link"
// @Param image formData file true "Clinic image"
// @Success 201 {object} ClinicResponse
// @Failure 400 {object} map[string]string
// @Router /clinics [post]
func (h *Handler) CreateClinic(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	clinic, err := h.service.CreateClinic(c.Request.Context(), req, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, clinic)
}

// GetClinic godoc
// @Summary Get clinic by ID
// @Tags Clinics
// @Produce json
// @Param id path int true "Clinic ID"
// @Success 200 {object} ClinicResponse
// @Failure 404 {object} map[string]string
// @Router /clinics/{id} [get]
func (h *Handler) GetClinic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clinic ID"})
		return
	}

	clinic, err := h.service.GetClinic(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, clinic)
}

// GetAllClinics godoc
// @Summary Get all clinics
// @Tags Clinics
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(10)
// @Success 200 {object} PaginationResponse
// @Failure 400 {object} map[string]string
// @Router /clinics [get]
func (h *Handler) GetAllClinics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.service.GetAllClinics(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateClinic godoc
// @Summary Update clinic
// @Tags Clinics
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Clinic ID"
// @Param name formData string false "Clinic name"
// @Param phone_number formData string false "Clinic phone number"
// @Param address formData string false "Clinic address"
// @Param map_link formData string false "Map link"
// @Param image formData file false "Clinic image"
// @Success 200 {object} ClinicResponse
// @Failure 400 {object} map[string]string
// @Router /clinics/{id} [put]
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clinic ID"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	clinic, err := h.service.UpdateClinic(c.Request.Context(), req, id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, clinic)
}

// DeleteClinic godoc
// @Summary Delete clinic
// @Tags Clinics
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Clinic ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /clinics/{id} [delete]
func (h *Handler) DeleteClinic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clinic ID"})
		return
	}

	userID := c.GetInt("user_id")
	if err := h.service.DeleteClinic(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clinic deleted successfully"})
}

// SearchClinics godoc
// @Summary Search clinics
// @Tags Clinics
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(10)
// @Success 200 {object} PaginationResponse
// @Failure 400 {object} map[string]string
// @Router /clinics/search [get]
func (h *Handler) SearchClinics(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.service.SearchClinics(c.Request.Context(), query, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func handleError(c *gin.Context, err error) {
	switch err {
	case ErrClinicNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrUnauthorizedAccess:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case ErrInvalidImageFormat:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrPaginationInvalid:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
