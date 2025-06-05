// internal/handlers/auth.go
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
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
	ExpiresIn    int64       `json:"expires_in"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
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
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, 'user', NOW(), NOW()) 
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
		"data": gin.H{
			"user_id": userID,
		},
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
		FROM users WHERE email = $1 AND role != 'deleted'
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
	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(user.ID, user.Username, user.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate tokens",
		})
		return
	}

	// Store refresh token
	_, err = h.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days', NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			token = EXCLUDED.token, 
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
	`, user.ID, refreshToken)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to store session",
		})
		return
	}

	// Update last login
	h.db.Exec("UPDATE users SET updated_at = NOW() WHERE id = $1", user.ID)

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
			ExpiresIn:    expiresIn,
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
		WHERE rt.token = $1 AND rt.expires_at > NOW() AND u.role != 'deleted'
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
	accessToken, expiresIn, err := utils.GenerateAccessToken(user.ID, user.Username, user.Role, h.jwtSecret)
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
			ExpiresIn:   expiresIn,
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
	_, err := h.db.Exec("DELETE FROM refresh_tokens WHERE token = $1", req.RefreshToken)
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

// GetMe returns current user profile
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var user models.User
	err := h.db.QueryRow(`
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users WHERE id = $1 AND role != 'deleted'
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, 
		&user.LastName, &user.Bio, &user.Avatar, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}