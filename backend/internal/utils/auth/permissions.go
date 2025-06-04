//file: internal/utils/auth/permissions.go

package auth

import (
	"context"
	"net/http"
	"veza-web-app/internal/constants"
	"veza-web-app/internal/utils/response"
)

// HasPermission vérifie si un utilisateur a une permission spécifique
func HasPermission(userRole constants.Role, permission constants.Permission) bool {
	permissions, exists := constants.RolePermissions[userRole]
	if !exists {
		return false
	}
	
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// RequirePermission middleware pour vérifier les permissions
func RequirePermission(permission constants.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("user_role").(constants.Role)
			if !ok {
				response.ErrorJSON(w, "Rôle utilisateur non trouvé", http.StatusUnauthorized)
				return
			}
			
			if !HasPermission(userRole, permission) {
				response.ErrorJSON(w, "Permission insuffisante", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware pour les routes admin
func RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("user_role").(constants.Role)
			if !ok {
				response.ErrorJSON(w, "Rôle utilisateur non trouvé", http.StatusUnauthorized)
				return
			}
			
			if userRole != constants.RoleAdmin && userRole != constants.RoleSuperAdmin {
				response.ErrorJSON(w, "Accès admin requis", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}