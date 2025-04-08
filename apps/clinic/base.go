package clinic

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

	clinicGroup := router.Group("api/v1/clinics")
	{
		clinicGroup.GET("", handler.GetAllClinics)
		clinicGroup.GET("/search", handler.SearchClinics)
		clinicGroup.GET("/:id", handler.GetClinic)

		// Authenticated routes
		authGroup := clinicGroup.Group("")
		authGroup.Use(middleware.AuthMiddleware(tokenService))
		{
			authGroup.POST("", handler.CreateClinic)
			authGroup.PUT("/:id", handler.UpdateClinic)
			authGroup.DELETE("/:id", handler.DeleteClinic)
		}
	}
}
