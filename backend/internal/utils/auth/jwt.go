//file: backend/utils/jwt.go

package auth

import (
	"time"
	"os"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtKey = []byte(strings.Trim(os.Getenv("JWT_SECRET"), `"`))

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

func GenerateRefreshToken() (string, error) {
	return uuid.NewString(), nil
}