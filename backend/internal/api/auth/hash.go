// File: internal/utils/auth/hash.go
package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// File: internal/utils/auth/jwt.go
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token
func GenerateJWT(userID int, email, role, secretKey string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// File: internal/utils/auth/permissions.go
package auth

// Role constants
const (
	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)

// HasRole checks if a user has a specific role
func HasRole(userRole, requiredRole string) bool {
	switch requiredRole {
	case RoleUser:
		return userRole == RoleUser || userRole == RoleAdmin || userRole == RoleSuperAdmin
	case RoleAdmin:
		return userRole == RoleAdmin || userRole == RoleSuperAdmin
	case RoleSuperAdmin:
		return userRole == RoleSuperAdmin
	default:
		return false
	}
}

// IsAdmin checks if a user has admin privileges
func IsAdmin(role string) bool {
	return role == RoleAdmin || role == RoleSuperAdmin
}

// IsSuperAdmin checks if a user has super admin privileges
func IsSuperAdmin(role string) bool {
	return role == RoleSuperAdmin
}