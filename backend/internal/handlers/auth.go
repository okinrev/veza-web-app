// internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/models"
	"veza-web-app/internal/utils"
)

type AuthHandler struct {
	db        *database.DB
	jwtSecret string
}

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         models.User `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func NewAuthHandler(db *database.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user account
func (h *AuthHandler) Register(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Normalize input
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process password",
		})
		return
	}

	// Create user
	var userID int
	err = h.db.QueryRow(`
		INSERT INTO users (username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW()) 
		RETURNING id
	`, req.Username, req.Email, hashedPassword).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Email or username already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create account",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Account created successfully",
		"user_id": userID,
	})
}

// Login authenticates a user and returns tokens
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Get user from database
	var user models.User
	err := h.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, updated_at 
		FROM users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, 
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	// Verify password
	if err := utils.CheckPasswordHash(req.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	// Generate tokens
	accessToken, err := utils.GenerateJWT(user.ID, user.Username, user.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate access token",
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate refresh token",
		})
		return
	}

	// Store refresh token
	_, err = h.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days')
		ON CONFLICT (user_id) DO UPDATE SET 
			token = EXCLUDED.token, 
			expires_at = EXCLUDED.expires_at
	`, user.ID, refreshToken)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to store session",
		})
		return
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
		},
	})
}

// RefreshToken generates a new access token from a refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Verify refresh token and get user
	var user models.User
	err := h.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token = $1 AND rt.expires_at > NOW()
	`, req.RefreshToken).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, 
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid or expired refresh token",
		})
		return
	}

	// Generate new access token
	accessToken, err := utils.GenerateJWT(user.ID, user.Username, user.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate access token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": RefreshResponse{
			AccessToken: accessToken,
		},
	})
}

// Logout invalidates the refresh token
func (h *AuthHandler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Delete refresh token
	_, err := h.db.Exec(`DELETE FROM refresh_tokens WHERE token = $1`, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}