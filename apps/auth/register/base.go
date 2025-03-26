package register

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Run menginisialisasi modul register dan menambahkan rute endpoint untuk registrasi pengguna.
func InitModule(router *gin.Engine, db *sql.DB) {
	// Logging untuk menandakan bahwa modul register sedang diinisialisasi
	slog.Info("Initializing register module...")

	// Inisialisasi repository yang akan berkomunikasi dengan database
	repo := NewRepository(db)

	// Inisialisasi service yang menangani logika bisnis registrasi
	svc := NewService(repo)

	// Inisialisasi handler yang bertanggung jawab untuk menangani request dari client
	handler := NewHandler(svc)

	// Membuat route group "/v1/auth" untuk mengelompokkan endpoint yang berkaitan dengan autentikasi
	authRouter := router.Group("api/v1/auth")
	{
		// Menambahkan endpoint POST /register untuk proses registrasi
		authRouter.POST("/register", handler.Register)
	}

	// Logging tambahan untuk memastikan endpoint register telah diinisialisasi dengan benar
	slog.Debug("Register route initialized",
		slog.String("path", "/v1/auth/register"), // Menampilkan path endpoint yang dibuat
		slog.String("method", http.MethodPost),   // Menampilkan metode HTTP yang digunakan (POST)
	)
}
