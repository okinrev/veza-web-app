// internal/utils/auth.go
package utils

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "strconv"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// Password hashing
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// JWT Claims
type Claims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// JWT functions
func GenerateTokenPair(userID int, username, role, secret string) (access, refresh string, expiresIn int64, error) {
    // Access token (1 hour)
    accessToken, expiresIn, err := GenerateAccessToken(userID, username, role, secret)
    if err != nil {
        return "", "", 0, err
    }

    // Refresh token (7 days)
    refreshClaims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString([]byte(secret))
    if err != nil {
        return "", "", 0, fmt.Errorf("failed to sign refresh token: %w", err)
    }

    return accessToken, refreshTokenString, expiresIn, nil
}

func GenerateAccessToken(userID int, username, role, secret string) (string, int64, error) {
    expiresIn := int64(3600) // 1 hour
    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Second)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(secret))
    if err != nil {
        return "", 0, fmt.Errorf("failed to sign token: %w", err)
    }

    return tokenString, expiresIn, nil
}

func ValidateJWT(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
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

// Signed URL functions
func GenerateSignedURL(filename string, userID int, secret string) (string, error) {
    expires := time.Now().Add(time.Hour).Unix()
    message := fmt.Sprintf("%s:%d:%d", filename, userID, expires)
    
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(message))
    signature := hex.EncodeToString(h.Sum(nil))
    
    return fmt.Sprintf("/stream/signed/%s?expires=%d&signature=%s&user=%d", 
        filename, expires, signature, userID), nil
}

func ValidateSignedURL(filename string, userID int, expires int64, signature, secret string) bool {
    if time.Now().Unix() > expires {
        return false
    }
    
    message := fmt.Sprintf("%s:%d:%d", filename, userID, expires)
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(message))
    expectedSignature := hex.EncodeToString(h.Sum(nil))
    
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}