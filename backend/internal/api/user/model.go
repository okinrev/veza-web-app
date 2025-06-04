package user

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          int        `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	Password    string     `json:"-" db:"password_hash"` // Never expose password in JSON
	FirstName   *string    `json:"first_name" db:"first_name"`
	LastName    *string    `json:"last_name" db:"last_name"`
	Username    *string    `json:"username" db:"username"`
	Avatar      *string    `json:"avatar" db:"avatar"`
	Bio         *string    `json:"bio" db:"bio"`
	Role        string     `json:"role" db:"role"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	IsVerified  bool       `json:"is_verified" db:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// UserResponse represents the user data returned to clients (without sensitive fields)
type UserResponse struct {
	ID          int        `json:"id"`
	Email       string     `json:"email"`
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Username    *string    `json:"username"`
	Avatar      *string    `json:"avatar"`
	Bio         *string    `json:"bio"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	IsVerified  bool       `json:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	Password  string  `json:"password" binding:"required,min=8"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Username  *string `json:"username"`
	Role      string  `json:"role,omitempty"` // Only admins can set role
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Username   *string `json:"username"`
	Avatar     *string `json:"avatar"`
	Bio        *string `json:"bio"`
	IsActive   *bool   `json:"is_active"`   // Only admins can change this
	IsVerified *bool   `json:"is_verified"` // Only admins can change this
	Role       *string `json:"role"`        // Only admins can change this
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	Password  string  `json:"password" binding:"required,min=8"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Username  *string `json:"username"`
}

// ChangePasswordRequest represents the password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Username:    u.Username,
		Avatar:      u.Avatar,
		Bio:         u.Bio,
		Role:        u.Role,
		IsActive:    u.IsActive,
		IsVerified:  u.IsVerified,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// GetDisplayName returns the display name for the user
func (u *User) GetDisplayName() string {
	if u.Username != nil && *u.Username != "" {
		return *u.Username
	}
	
	if u.FirstName != nil && u.LastName != nil {
		return *u.FirstName + " " + *u.LastName
	}
	
	if u.FirstName != nil {
		return *u.FirstName
	}
	
	return u.Email
}

// IsAdmin checks if the user has admin privileges
func (u *User) IsAdmin() bool {
	return u.Role == "admin" || u.Role == "super_admin"
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}