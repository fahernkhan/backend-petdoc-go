package consultation

import (
	"errors"
	"net/http"
	"strconv"

	"petdoc/internal/infrastructure/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateConsultation godoc
// @Summary Buat konsultasi baru
// @Tags Konsultasi
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData integer true "User ID"
// @Param doctor_id formData integer true "Doctor ID"
// @Param pet_type formData string true "Jenis Hewan"
// @Param pet_name formData string true "Nama Hewan"
// @Param pet_age formData integer true "Umur Hewan"
// @Param disease_description formData string true "Deskripsi Penyakit"
// @Param consultation_date formData string true "Tanggal Konsultasi (YYYY-MM-DD)"
// @Param start_time formData string true "Waktu Mulai (HH:MM:SSZ)"
// @Param end_time formData string true "Waktu Selesai (HH:MM:SSZ)"
// @Param payment_proof formData string true "Bukti Pembayaran (Base64)"
// @Success 201 {object} ConsultationResponse
// @Failure 400 {object} map[string]string
// @Router /consultations [post]
func (h *Handler) CreateConsultation(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi user
	if !middleware.ValidateUser(c, req.UserID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "akses tidak diizinkan"})
		return
	}

	res, err := h.service.CreateConsultation(c.Request.Context(), req)
	if err != nil {
		c.JSON(getErrorCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetConsultations godoc
// @Summary Dapatkan riwayat konsultasi
// @Tags Konsultasi
// @Produce json
// @Param user_id query int true "User ID"
// @Param page query int false "Halaman" default(1)
// @Param page_size query int false "Item per halaman" default(10)
// @Success 200 {object} PaginationResponse
// @Failure 400 {object} map[string]string
// @Router /consultations [get]
func (h *Handler) GetConsultations(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Query("user_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if userID < 1 || page < 1 || pageSize < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter tidak valid"})
		return
	}

	// Validasi user
	if !middleware.ValidateUser(c, userID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "akses tidak diizinkan"})
		return
	}

	res, err := h.service.GetConsultations(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func getErrorCode(err error) int {
	switch {
	case errors.Is(err, ErrInvalidTimeFormat):
		return http.StatusBadRequest
	case errors.Is(err, ErrDoctorNotAvailable):
		return http.StatusConflict
	case errors.Is(err, ErrConsultationPastDate):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
