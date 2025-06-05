package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware creates a JWT authentication middleware
func JWTAuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"success": false,
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:] // Remove "Bearer " prefix
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header must start with 'Bearer '",
				"success": false,
			})
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Make sure that the token method conforms to "SigningMethodHMAC"
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid token signing method",
					"success": false,
				})
				c.Abort()
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token: " + err.Error(),
				"success": false,
			})
			c.Abort()
			return
		}

		// Check if token is valid and extract claims
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// Set user information in the context for use in handlers
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)
			c.Set("claims", claims)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token claims",
				"success": false,
			})
			c.Abort()
			return
		}

		// Continue to the next handler
		c.Next()
	}
}

// AdminMiddleware checks if the user has admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User role not found in context",
				"success": false,
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || (role != "admin" && role != "super_admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Admin access required",
				"success": false,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserEmailFromContext extracts user email from the Gin context
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	email, ok := userEmail.(string)
	return email, ok
}

// GetUserRoleFromContext extracts user role from the Gin context
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := userRole.(string)
	return role, ok
}