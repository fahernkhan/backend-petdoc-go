package middleware

import (
	"net/http"
	"strings"

	"petdoc/internal/infrastructure/utils/jwt"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware untuk validasi JWT
func AuthMiddleware(tokenService jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token format",
			})
			return
		}

		claims, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token",
				"details": err.Error(),
			})
			return
		}

		// Set user context untuk digunakan di handler
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// AdminOnly middleware untuk membatasi akses ke admin saja
// AdminOnly middleware untuk membatasi akses ke admin saja
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Unauthorized access",
			})
			return
		}

		// Pastikan role bertipe string
		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Invalid role format",
			})
			return
		}

		// Handle case-insensitive role check
		if strings.ToLower(roleStr) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Admin access required",
			})
			return
		}

		c.Next()
	}
}
