// internal/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "veza-web-app/internal/utils"
)

// JWTAuthMiddleware validates JWT tokens and sets user context
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Authorization header required",
            })
            c.Abort()
            return
        }

        // Check if the header starts with "Bearer "
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Authorization header must start with 'Bearer '",
            })
            c.Abort()
            return
        }

        tokenString := authHeader[7:] // Remove "Bearer " prefix

        // Verify token
        claims, err := utils.ValidateJWT(tokenString, jwtSecret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Invalid token: " + err.Error(),
            })
            c.Abort()
            return
        }

        // Set user information in the context
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("user_role", claims.Role)

        c.Next()
    }
}

// OptionalJWTAuthMiddleware validates JWT tokens but doesn't require them
func OptionalJWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            tokenString := authHeader[7:]
            if claims, err := utils.ValidateJWT(tokenString, jwtSecret); err == nil {
                c.Set("user_id", claims.UserID)
                c.Set("username", claims.Username)
                c.Set("user_role", claims.Role)
            }
        }
        c.Next()
    }
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("user_role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "User role not found in context",
            })
            c.Abort()
            return
        }

        role, ok := userRole.(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Invalid user role",
            })
            c.Abort()
            return
        }

        // Check if user has one of the required roles
        hasRole := false
        for _, requiredRole := range roles {
            if role == requiredRole {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error":   "Insufficient permissions",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// AdminMiddleware checks if the user has admin role
func AdminMiddleware() gin.HandlerFunc {
    return RequireRole("admin", "super_admin")
}