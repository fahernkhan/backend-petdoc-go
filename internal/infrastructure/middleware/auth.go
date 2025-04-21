package middleware

import (
	"net/http"
	"strconv"
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

// ValidateUser memeriksa apakah user memiliki akses berdasarkan user_id dari token JWT
func ValidateUser(c *gin.Context, userID int) bool {
	// Ambil user_id dari JWT yang sudah diset di context
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		return false
	}

	// Konversi tokenUserID ke tipe int
	tokenUserIDInt, ok := tokenUserID.(int)
	if !ok {
		return false
	}

	// Validasi apakah user_id dari request sama dengan user_id dari token JWT
	return tokenUserIDInt == userID
}

// ValidateDoctor memeriksa apakah user memiliki akses berdasarkan user_id dari token JWT
func ValidateDoctor(c *gin.Context, doctorID int) bool {
	// Ambil user_id dari JWT yang sudah diset di context
	tokenUserID, exists := c.Get("user_id")
	if !exists {
		return false
	}

	// Konversi ke int
	tokenUserIDInt, ok := tokenUserID.(int)
	if !ok {
		return false
	}

	// Ambil role dari JWT
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}

	// Validasi role dan ID
	return role == "doctor" && tokenUserIDInt == doctorID
}

// chatgpt rule:
// IsAdmin untuk cek apakah user adalah admin
func IsAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}
	roleStr, ok := role.(string)
	if !ok {
		return false
	}
	return strings.ToLower(roleStr) == "admin"
}

// CanAccessUser hanya memperbolehkan akses data user tertentu
func CanAccessUser(c *gin.Context, targetUserID int) bool {
	if IsAdmin(c) {
		return true
	}

	userID, exists := c.Get("user_id")
	if !exists {
		return false
	}

	userIDInt, ok := userID.(int)
	if !ok {
		return false
	}

	return userIDInt == targetUserID
}

// UserAccess middleware untuk endpoint get by user_id
func UserAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil role dari context
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses tidak diizinkan"})
			return
		}

		// Jika admin, langsung izinkan
		if roleStr, ok := role.(string); ok && strings.ToLower(roleStr) == "admin" {
			c.Next()
			return
		}

		// Ambil user_id dari query parameter
		userIDParam := c.Query("user_id")
		if userIDParam == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id diperlukan"})
			return
		}

		// Konversi ke integer
		requestedUserID, err := strconv.Atoi(userIDParam)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id tidak valid"})
			return
		}

		// Ambil user_id dari token
		loggedInUserID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses tidak diizinkan"})
			return
		}

		// Bandingkan user_id
		if loggedInUserID != requestedUserID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
			return
		}

		c.Next()
	}
}

// DoctorAccess middleware untuk endpoint get by doctor_id
func DoctorAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses tidak diizinkan"})
			return
		}

		// Jika admin, langsung izinkan
		if roleStr, ok := role.(string); ok && strings.ToLower(roleStr) == "admin" {
			c.Next()
			return
		}

		// Cek jika role dokter
		roleStr, ok := role.(string)
		if !ok || strings.ToLower(roleStr) != "doctor" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses dokter diperlukan"})
			return
		}

		// Ambil doctor_id dari query
		doctorIDParam := c.Query("doctor_id")
		if doctorIDParam == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "doctor_id diperlukan"})
			return
		}

		requestedDoctorID, err := strconv.Atoi(doctorIDParam)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "doctor_id tidak valid"})
			return
		}

		// Ambil user_id dari token (doctor_id == user_id)
		loggedInUserID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses tidak diizinkan"})
			return
		}

		if loggedInUserID != requestedDoctorID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
			return
		}

		c.Next()
	}
}
