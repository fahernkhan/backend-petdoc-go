package doctor

import (
	"database/sql"
	"petdoc/internal/infrastructure/middleware"
	"petdoc/internal/infrastructure/utils/jwt"

	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine, db *sql.DB, tokenService jwt.JWT) {
	// func InitRoutes(router *gin.RouterGroup, db *sql.DB, tokenService jwt.JWT) {
	repo := NewRepository(db)
	service := NewService(repo)
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
