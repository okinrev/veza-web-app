package common

import (
    "errors"

    "github.com/gin-gonic/gin"
)

// GetUserIDFromContext extracts the user ID from the gin.Context.
func GetUserIDFromContext(c *gin.Context) (int, bool) {
    userID, exists := c.Get("user_id")
    if !exists {
        return 0, false
    }
    id, ok := userID.(int)
    return id, ok
}

// GetUsernameFromContext extracts username from the Gin context
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

// RequireOwnership middleware checks if user owns the resource
func RequireOwnership(getOwnerIDFunc func(*gin.Context) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "User not authenticated",
			})
			c.Abort()
			return
		}

		ownerID, err := getOwnerIDFunc(c)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Resource not found",
			})
			c.Abort()
			return
		}

		if userID != ownerID {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
