// internal/api/auth/service.go
package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	
	"veza-web-app/internal/api/user"
	"veza-web-app/internal/utils/auth"
)

// Service handles authentication business logic
type Service struct {
	db        *sql.DB
	jwtSecret string
	userService *user.Service
}

// NewService creates a new auth service
func NewService(db *sql.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
		userService: user.NewService(db),
	}
}

// Login handles user login
func (s *Service) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data: " + err.Error(),
			"success": false,
		})
		return
	}

	// Get user by email
	userRecord, err := s.userService.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid email or password",
			"success": false,
		})
		return
	}

	// Check if user is active
	if !userRecord.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Account is deactivated",
			"success": false,
		})
		return
	}

	// Verify password
	if !auth.CheckPasswordHash(req.Password, userRecord.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid email or password",
			"success": false,
		})
		return
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := s.generateTokens(userRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate tokens",
			"success": false,
		})
		return
	}

	// Update last login
	if err := s.userService.UpdateLastLogin(userRecord.ID); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login for user %d: %v\n", userRecord.ID, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": user.AuthResponse{
			User:         userRecord.ToResponse(),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    expiresIn,
		},
	})
}

// Register handles user registration
func (s *Service) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data: " + err.Error(),
			"success": false,
		})
		return
	}

	// Create user
	createReq := user.CreateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Role:      "user", // Default role
	}

	newUser, err := s.userService.CreateUser(createReq)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email already exists",
				"success": false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registration successful",
		"data":    newUser,
	})
}

// RefreshToken handles token refresh
func (s *Service) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data: " + err.Error(),
			"success": false,
		})
		return
	}

	// Validate refresh token
	userID, err := s.validateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid refresh token",
			"success": false,
		})
		return
	}

	// Get user
	userResponse, err := s.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not found",
			"success": false,
		})
		return
	}

	// Convert UserResponse back to User for token generation
	userRecord := &user.User{
		ID:       userResponse.ID,
		Email:    userResponse.Email,
		Role:     userResponse.Role,
		IsActive: userResponse.IsActive,
	}

	// Generate new tokens
	accessToken, newRefreshToken, expiresIn, err := s.generateTokens(userRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate tokens",
			"success": false,
		})
		return
	}

	// Invalidate old refresh token
	if err := s.invalidateRefreshToken(req.RefreshToken); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to invalidate old refresh token: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
			"expires_in":    expiresIn,
		},
	})
}

// generateTokens generates both access and refresh tokens
func (s *Service) generateTokens(user *user.User) (string, string, int64, error) {
	// Access token (expires in 1 hour)
	accessTokenExpiry := time.Now().Add(time.Hour)
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     accessTokenExpiry.Unix(),
		"iat":     time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", 0, err
	}

	// Refresh token (expires in 30 days)
	refreshTokenExpiry := time.Now().Add(30 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     refreshTokenExpiry.Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", 0, err
	}

	// Store refresh token in database
	if err := s.storeRefreshToken(user.ID, refreshTokenString, refreshTokenExpiry); err != nil {
		return "", "", 0, err
	}

	return accessTokenString, refreshTokenString, accessTokenExpiry.Unix(), nil
}

// storeRefreshToken stores a refresh token in the database
func (s *Service) storeRefreshToken(userID int, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	
	_, err := s.db.Exec(query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	
	return nil
}

// validateRefreshToken validates a refresh token and returns the user ID
func (s *Service) validateRefreshToken(tokenString string) (int, error) {
	// Parse and validate JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	// Check if it's a refresh token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return 0, fmt.Errorf("not a refresh token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user ID in token")
	}

	// Check if token exists in database and is not expired
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM refresh_tokens 
			WHERE token = $1 AND user_id = $2 AND expires_at > CURRENT_TIMESTAMP
		)
	`
	
	err = s.db.QueryRow(query, tokenString, int(userID)).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	
	if !exists {
		return 0, fmt.Errorf("refresh token not found or expired")
	}

	return int(userID), nil
}

// invalidateRefreshToken removes a refresh token from the database
func (s *Service) invalidateRefreshToken(token string) error {
	query := "DELETE FROM refresh_tokens WHERE token = $1"
	_, err := s.db.Exec(query, token)
	return err
}

// Logout handles user logout (invalidates refresh token)
func (s *Service) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	// This is optional - user can logout without providing refresh token
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		if err := s.invalidateRefreshToken(req.RefreshToken); err != nil {
			// Log error but don't fail the logout
			fmt.Printf("Failed to invalidate refresh token during logout: %v\n", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}