// internal/api/user/models.go
package user

import (
    "time"
)

// User represents the user model for database operations
type User struct {
    ID           int        `db:"id"`
    Email        string     `db:"email"`
    Password     string     `db:"password_hash"`
    FirstName    string     `db:"first_name"`
    LastName     string     `db:"last_name"`
    Username     string     `db:"username"`
    Avatar       *string    `db:"avatar"`
    Bio          *string    `db:"bio"`
    Role         string     `db:"role"`
    IsActive     bool       `db:"is_active"`
    IsVerified   bool       `db:"is_verified"`
    LastLoginAt  *time.Time `db:"last_login_at"`
    CreatedAt    time.Time  `db:"created_at"`
    UpdatedAt    time.Time  `db:"updated_at"`
}

// UserResponse represents the user data returned to clients (without password)
type UserResponse struct {
    ID          int        `json:"id"`
    Email       string     `json:"email"`
    FirstName   string     `json:"first_name"`
    LastName    string     `json:"last_name"`
    Username    string     `json:"username"`
    Avatar      *string    `json:"avatar"`
    Bio         *string    `json:"bio"`
    Role        string     `json:"role"`
    IsActive    bool       `json:"is_active"`
    IsVerified  bool       `json:"is_verified"`
    LastLoginAt *time.Time `json:"last_login_at"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
    return &UserResponse{
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

// CreateUserRequest for user creation
type CreateUserRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=6"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
    Username  string `json:"username" validate:"required,min=3"`
    Role      string `json:"role,omitempty"`
}

// UpdateUserRequest for user updates
type UpdateUserRequest struct {
    FirstName  *string `json:"first_name,omitempty"`
    LastName   *string `json:"last_name,omitempty"`
    Username   *string `json:"username,omitempty"`
    Avatar     *string `json:"avatar,omitempty"`
    Bio        *string `json:"bio,omitempty"`
    IsActive   *bool   `json:"is_active,omitempty"`
    IsVerified *bool   `json:"is_verified,omitempty"`
    Role       *string `json:"role,omitempty"`
}

// LoginRequest for authentication
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// RegisterRequest for user registration
type RegisterRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=6"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
    Username  string `json:"username" validate:"required,min=3"`
}

// AuthResponse returned after successful authentication
type AuthResponse struct {
    User         *UserResponse `json:"user"`
    AccessToken  string        `json:"access_token"`
    RefreshToken string        `json:"refresh_token"`
    ExpiresIn    int64         `json:"expires_in"`
}