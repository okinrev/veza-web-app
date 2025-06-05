// internal/api/middleware/validation.go
package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ValidatePagination validates pagination parameters
func ValidatePagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid page parameter",
			})
			c.Abort()
			return
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid limit parameter (1-100)",
			})
			c.Abort()
			return
		}

		c.Set("page", page)
		c.Set("limit", limit)
		c.Next()
	}
}

// ValidateJSON ensures the request has valid JSON content type
func ValidateJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}