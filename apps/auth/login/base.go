package login

import (
	"database/sql"
	"log/slog"
	"net/http"
	"petdoc/internal/infrastructure/utils/jwt"
	"time"

	"github.com/gin-gonic/gin"
)

func InitModule(router *gin.Engine, db *sql.DB, jwt jwt.JWT, tokenExp time.Duration) {
	slog.Info("Initializing login module...")

	repo := NewRepository(db)
	service := NewService(repo, jwt, tokenExp)
	handler := NewHandler(service)

	authRouter := router.Group("api/v1/auth")
	{
		authRouter.POST("/login", handler.Login)
	}

	slog.Debug("Login route initialized",
		slog.String("path", "/v1/auth/login"),
		slog.String("method", http.MethodPost),
	)

}
