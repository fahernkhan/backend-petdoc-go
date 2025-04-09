package doctor

import (
	"database/sql"
	"log/slog"
	"petdoc/internal/infrastructure/cloudinary"
	"petdoc/internal/infrastructure/middleware"
	"petdoc/internal/infrastructure/utils/jwt"

	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine, db *sql.DB, cloudinary cloudinary.Service, tokenService jwt.JWT) {
	repo := NewRepository(db)
	service := NewService(repo, cloudinary, slog.Default())
	handler := NewHandler(service)

	doctorGroup := router.Group("api/v1/doctors")
	doctorGroup.Use(middleware.AuthMiddleware(tokenService)) // Middleware untuk semua endpoint dokter

	{
		// Public routes (tanpa admin)
		doctorGroup.GET("", handler.ListDoctors)
		doctorGroup.GET("/:id", handler.GetDoctor)

		// Admin-only routes
		adminRoutes := doctorGroup.Group("")
		adminRoutes.Use(middleware.AdminOnly())
		{
			adminRoutes.POST("", handler.CreateDoctor)
			adminRoutes.PUT("/:id", handler.UpdateDoctor)
			adminRoutes.DELETE("/:id", handler.DeleteDoctor)
		}
	}
}
