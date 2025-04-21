package admin

import (
	"database/sql"
	"petdoc/internal/infrastructure/middleware"
	"petdoc/internal/infrastructure/utils/jwt"

	"github.com/gin-gonic/gin"
)

func InitAdminRoutes(router *gin.Engine, db *sql.DB, jwtService jwt.JWT) {
	repo := NewAdminRepository(db)
	service := NewAdminService(repo)
	handler := NewAdminHandler(service)

	adminRoutes := router.Group("api/v1/admin")
	adminRoutes.Use(
		middleware.AuthMiddleware(jwtService),
		middleware.AdminOnly(), // Ganti RoleMiddleware dengan AdminOnly
	)
	{
		adminRoutes.POST("", handler.CreateAdmin)
	}
}
