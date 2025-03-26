package response

import (
	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, err error, details gin.H) {
	if details == nil {
		details = gin.H{}
	}

	details["type"] = "general_error"
	if details["type"] == "" {
		details["type"] = "unknown_error"
	}

	c.JSON(status, gin.H{
		"success": false,
		"error":   err.Error(),
		"details": details,
	})
}
