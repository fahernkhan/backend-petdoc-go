package doctor

import (
	"net/http"
	"petdoc/internal/infrastructure/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// @Summary Create new doctor
// @Tags Doctors
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body DoctorRequest true "Doctor data"
// @Success 201 {object} DoctorResponse
// @Router /doctors [post]
func (h *Handler) CreateDoctor(c *gin.Context) {
	var req DoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Handle error binding
		response.Error(c, http.StatusBadRequest, ErrInvalidDoctorData, gin.H{
			"details": err.Error(),
		})
		return
	}

	doctor, err := h.service.CreateDoctor(c.Request.Context(), req)
	if err != nil {
		errResp := MapError(err)
		response.Error(c, errResp.Code, err, gin.H{
			"details": errResp.Details,
		})
		return
	}

	response.Success(c, http.StatusCreated, doctor)
}

// @Summary Get all doctors with pagination
// @Tags Doctors
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} []DoctorResponse
// @Router /doctors [get]
func (h *Handler) ListDoctors(c *gin.Context) {
	var pagination Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctors, total, err := h.service.ListDoctors(c.Request.Context(), pagination)
	if err != nil {
		response := MapError(err)
		c.JSON(response.Code, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  doctors,
		"total": total,
		"page":  pagination.Page,
		"limit": pagination.Limit,
	})
}

// Implementasi handler lainnya untuk Get, Update, Delete
// @Summary Get doctor by ID
// @Tags Doctors
// @Produce json
// @Param id path int true "Doctor ID"
// @Success 200 {object} DoctorResponse
// @Router /doctors/{id} [get]
func (h *Handler) GetDoctor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, MapError(ErrInvalidDoctorData))
		return
	}

	doctor, err := h.service.GetDoctor(c.Request.Context(), id)
	if err != nil {
		response := MapError(err)
		c.JSON(response.Code, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    doctor,
	})
}

// @Summary Update doctor
// @Tags Doctors
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Doctor ID"
// @Param input body DoctorRequest true "Doctor data"
// @Success 200 {object} DoctorResponse
// @Router /doctors/{id} [put]
func (h *Handler) UpdateDoctor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, MapError(ErrInvalidDoctorData))
		return
	}

	var req DoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MapError(ErrInvalidDoctorData))
		return
	}

	doctor, err := h.service.UpdateDoctor(c.Request.Context(), id, req)
	if err != nil {
		response := MapError(err)
		c.JSON(response.Code, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    doctor,
	})
}

// @Summary Delete doctor
// @Tags Doctors
// @Security BearerAuth
// @Produce json
// @Param id path int true "Doctor ID"
// @Success 204
// @Router /doctors/{id} [delete]
func (h *Handler) DeleteDoctor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, MapError(ErrInvalidDoctorData))
		return
	}

	if err := h.service.DeleteDoctor(c.Request.Context(), id); err != nil {
		response := MapError(err)
		c.JSON(response.Code, response)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
