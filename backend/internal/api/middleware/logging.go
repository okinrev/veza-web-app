// internal/api/middleware/logging.go
package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a Gin middleware for logging HTTP requests
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CustomLogger returns a custom logger middleware with detailed information
func CustomLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Get user info if available
		userID, hasUserID := GetUserIDFromContext(c)
		username, _ := GetUsernameFromContext(c)

		userInfo := "anonymous"
		if hasUserID {
			userInfo = fmt.Sprintf("user_id=%d", userID)
			if username != "" {
				userInfo += fmt.Sprintf(" username=%s", username)
			}
		}

		// Log the request
		fmt.Printf("[%s] %s %s %d %v %d bytes - %s - %s - %s\n",
			time.Now().Format("2006/01/02 15:04:05"),
			method,
			path,
			statusCode,
			latency,
			bodySize,
			clientIP,
			userAgent,
			userInfo,
		)
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Internal server error",
				"message": "Something went wrong",
			})
		} else {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Internal server error",
				"message": "Something went wrong",
			})
		}
		c.Abort()
	})
}

