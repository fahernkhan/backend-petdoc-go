package user

import (
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// InitUserModule menginisialisasi modul User dengan middleware dan routing
func InitUserModule(router *gin.Engine, db *sql.DB) {
	slog.Info("Initializing User module...")

	// Dependency Injection
	repo := NewUserRepository(db)
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	// Konfigurasi routing dengan middleware
	adminRoutes := router.Group("api/v1/users")
	// adminRoutes.Use(
	// 	middleware.JWTAuth(),        // JWT Authentication
	// 	middleware.AdminOnly(),      // Hanya admin yang bisa akses
	// 	middleware.RequestLogger(),  // Log semua request
	// 	middleware.RateLimiter(100), // Limit 100 request/min
	// )

	// // Definisi routes
	adminRoutes.GET("", handler.GetAllUsers) // Ambil semua user
	// adminRoutes.POST("", handler.CreateUser)       // Tambah user baru
	// adminRoutes.GET("/:id", handler.GetUser)       // Ambil user berdasarkan ID
	// adminRoutes.PUT("/:id", handler.UpdateUser)    // Update data user
	// adminRoutes.DELETE("/:id", handler.DeleteUser) // Hapus user

	slog.Debug("User module initialized",
		slog.String("base_path", "/api/v1/users"),
		slog.String("methods", "GET, POST, PUT, DELETE"),
	)
}
