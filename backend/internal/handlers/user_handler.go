//file: backend/handlers/user.go

package handlers

import (
	"encoding/json"
	"strings"
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
	"fmt"

	"veza-web-app/db"
	"veza-web-app/utils"
	"veza-web-app/models"
	"veza-web-app/middleware"
)

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "Champs manquants", http.StatusBadRequest)
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	var id int
	err = db.DB.QueryRowx(`
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3) RETURNING id
	`, req.Username, req.Email, hash).Scan(&id)

	if err != nil {
		http.Error(w, "Email ou nom déjà utilisé", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Compte créé avec succès"}`))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.DB.Get(&user, "SELECT * FROM users WHERE email = $1", req.Email)
	if err != nil {
		http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	if err := utils.CheckPasswordHash(req.Password, user.PasswordHash); err != nil {
		http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	accessToken, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Erreur JWT", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Erreur RT", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, NOW() + interval '7 days')
	`, user.ID, refreshToken)
	if err != nil {
		http.Error(w, "Erreur enregistrement RT", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var user models.User
	err := db.DB.Get(&user, "SELECT id, username, email, created_at FROM users WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🟡 UpdateUserHandler appelé")
	userIDFromToken := r.Context().Value(middleware.UserIDKey).(int)

	// On récupère l'id présent dans l'URL
	vars := mux.Vars(r)
	targetIDStr := vars["id"]

	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID != userIDFromToken {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	if req.Email == "" || req.Username == "" {
		http.Error(w, "Champs manquants", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(`
		UPDATE users SET email = $1, username = $2 WHERE id = $3
	`, req.Email, req.Username, userIDFromToken)

	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Profil mis à jour"}`))
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🛠️ DEBUG: handler ChangePasswordHandler atteint")
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.CurrentPassword) == "" || strings.TrimSpace(req.NewPassword) == "" {
		http.Error(w, "Champs requis manquants", http.StatusBadRequest)
		return
	}

	var storedHash string
	err := db.DB.Get(&storedHash, "SELECT password_hash FROM users WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Utilisateur introuvable", http.StatusInternalServerError)
		return
	}

	if err := utils.CheckPasswordHash(req.CurrentPassword, storedHash); err != nil {
		http.Error(w, "Mot de passe actuel incorrect", http.StatusUnauthorized)
		return
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", newHash, userID)
	if err != nil {
		http.Error(w, "Échec de la mise à jour", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Mot de passe mis à jour avec succès"}`))
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userIDFromToken := r.Context().Value(middleware.UserIDKey).(int)

	vars := mux.Vars(r)
	targetIDStr := vars["id"]

	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID != userIDFromToken {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec("DELETE FROM users WHERE id = $1", userIDFromToken)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Compte supprimé avec succès"}`))
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.DB.Get(&user, `
		SELECT users.* FROM refresh_tokens
		JOIN users ON users.id = refresh_tokens.user_id
		WHERE refresh_tokens.token = $1 AND refresh_tokens.expires_at > NOW()
	`, req.RefreshToken)

	if err != nil {
		http.Error(w, "Token invalide ou expiré", http.StatusUnauthorized)
		return
	}

	accessToken, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Erreur création JWT", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(RefreshResponse{AccessToken: accessToken})
}

// GetAllUsers retourne tous les utilisateurs (exclut les mots de passe)
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	err := db.DB.Select(&users, "SELECT id, username, email, created_at FROM users")
	if err != nil {
		http.Error(w, "Erreur récupération utilisateurs", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

// GetUserByID retourne les infos publiques d’un utilisateur par son ID
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var user models.User
	err = db.DB.Get(&user, "SELECT id, username, email, created_at FROM users WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	var users []models.User
	err := db.DB.Select(&users, `
		SELECT id, username, email, created_at
		FROM users
		WHERE username ILIKE $1 OR email ILIKE $1
		LIMIT 20
	`, "%"+query+"%")

	if err != nil {
		http.Error(w, "Erreur DB", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func GetUsersExceptMe(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(middleware.UserIDKey).(int)

	var users []models.User
	err := db.DB.Select(&users, `
		SELECT id, username, email, created_at
		FROM users
		WHERE id != $1
		ORDER BY created_at DESC
	`, currentUserID)

	if err != nil {
		http.Error(w, "Erreur DB", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func GetUserAvatar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	// 👇 Remplace cette logique par ton propre système d'avatar
	// Ex : lecture dans table avatars, ou fichier sur disque
	avatarURL := "/static/default-avatar.png" // fallback

	// Simule : si l'utilisateur existe on renvoie une image générique
	var exists bool
	err = db.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID)
	if err != nil || !exists {
		http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, avatarURL, http.StatusFound)
}
