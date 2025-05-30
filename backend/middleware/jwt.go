//file: backend/middleware/jwt.go

package middleware

import (
	"context"
	"net/http"
	"strings"
	"fmt"

	"veza-backend/utils"
)

type contextKey string

const UserIDKey = contextKey("user_id")

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Token manquant", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, "Token invalide", http.StatusUnauthorized)
			return
		}

		// Injecte l'ID utilisateur dans le contexte
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		fmt.Printf("✅ Utilisateur ID %d authentifié\n", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

