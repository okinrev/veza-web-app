// internal/api/user/models.go
package user

import (
	"database/sql"
	"time"
)

// User représente un utilisateur avec mot de passe (pour auth)
type User struct {
	ID          int            `json:"id" db:"id"`
	Email       string         `json:"email" db:"email"`
	Password    string         `json:"-" db:"password_hash"` // Ne pas sérialiser le mot de passe
	FirstName   *string        `json:"first_name" db:"first_name"`
	LastName    *string        `json:"last_name" db:"last_name"`
	Username    *string        `json:"username" db:"username"`
	Avatar      *string        `json:"avatar" db:"avatar"`
	Bio         *string        `json:"bio" db:"bio"`
	Role        string         `json:"role" db:"role"`
	IsActive    bool           `json:"is_active" db:"is_active"`
	IsVerified  bool           `json:"is_verified" db:"is_verified"`
	LastLoginAt *time.Time     `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// UserResponse représente un utilisateur sans données sensibles
type UserResponse struct {
	ID          int            `json:"id" db:"id"`
	Email       string         `json:"email" db:"email"`
	FirstName   *string        `json:"first_name" db:"first_name"`
	LastName    *string        `json:"last_name" db:"last_name"`
	Username    *string        `json:"username" db:"username"`
	Avatar      *string        `json:"avatar" db:"avatar"`
	Bio         *string        `json:"bio" db:"bio"`
	Role        string         `json:"role" db:"role"`
	IsActive    bool           `json:"is_active" db:"is_active"`
	IsVerified  bool           `json:"is_verified" db:"is_verified"`
	LastLoginAt *time.Time     `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// ToResponse convertit User en UserResponse
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

// Requests pour l'API
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Username  string `json:"username" binding:"required"`
}

type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Role      string `json:"role"`
}

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

type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
}