package consultation

import (
	"database/sql"
	"log/slog"

	"petdoc/internal/infrastructure/cloudinary"
	"petdoc/internal/infrastructure/middleware"
	"petdoc/internal/infrastructure/utils/jwt"

	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine, db *sql.DB, cloudinaryService cloudinary.Service, tokenService jwt.JWT) {
	repo := NewRepository(db)
	service := NewService(repo, cloudinaryService, slog.Default())
	handler := NewHandler(service)

	consultationGroup := router.Group("api/v1/consultations")
	consultationGroup.Use(middleware.AuthMiddleware(tokenService)) // Middleware untuk semua endpoint konsultasi

	{
		consultationGroup.POST("", handler.CreateConsultation)
		consultationGroup.GET("", handler.GetConsultations)

		// Jika ada endpoint khusus admin
		adminRoutes := consultationGroup.Group("")
		adminRoutes.Use(middleware.AdminOnly())
		{
			// adminRoutes.GET("/all", handler.GetAllConsultations)
		}
	}
}
