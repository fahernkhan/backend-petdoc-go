package register

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

// Membuat instance handler baru
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Register menangani request pendaftaran pengguna baru
func (h *Handler) Register(c *gin.Context) {
	logger := slog.With("module", "register-handler")

	var req RegisterRequest

	// 1️⃣ Bind JSON request ke struct RegisterRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		logger.Warn("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid request format", err))
		return
	}

	// 2️⃣ Validasi input agar tidak ada field kosong
	if req.Email == "" || req.Username == "" || req.Password == "" {
		logger.Warn("Missing required fields", "email", req.Email, "username", req.Username)
		c.JSON(http.StatusBadRequest, ErrorResponse("Missing required fields", ErrRequiredFieldsMissing))
		return
	}

	// Validasi format tanggal wajib pakai 2006-01-02 YYYY/MM/DD == 2006 Tahun (YYYY), 01 Bulan (MM), 02 Hari (DD)
	if _, err := time.Parse("2006-01-02", req.DateOfBirth); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Invalid date format. Gunakan format YYYY-MM-DD", err))
		return
	}

	logger.Info("Processing registration request", "email", req.Email, "username", req.Username)

	// 3️⃣ Panggil service untuk registrasi (penting)
	user, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to register user", "error", err)
		c.JSON(GetStatusCode(err), ErrorResponse("Registration failed", err))
		return
	}
	// 4️⃣ Hapus password sebelum mengembalikan response
	user.Password = ""

	logger.Info("User registered successfully", "user_id", user.ID, "email", user.Email)

	// 5️⃣ Kirim response sukses
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id": user.ID,
		},
		"details": "User registered successfully",
	})

	// c.JSON(http.StatusCreated, gin.H{
	// 	"success": true,
	// 	"data":    user,
	// })
}

// ErrorResponse mengembalikan format error response yang konsisten
// func ErrorResponse(message string, err error) gin.H {
// 	resp := gin.H{"success": false, "error": message}
// 	if err != nil {
// 		resp["details"] = err.Error()
// 	}
// 	return resp
// }
