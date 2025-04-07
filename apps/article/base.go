package article

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

	articleGroup := router.Group("api/v1/articles")
	{
		articleGroup.GET("", handler.GetAllArticles)
		articleGroup.GET("/search", handler.SearchArticles)
		articleGroup.GET("/:id", handler.GetArticle)

		// Authenticated routes
		authGroup := articleGroup.Group("")
		authGroup.Use(middleware.AuthMiddleware(tokenService))
		{
			authGroup.POST("", handler.CreateArticle)
			authGroup.PUT("/:id", handler.UpdateArticle)
			authGroup.DELETE("/:id", handler.DeleteArticle)
		}
	}
}
