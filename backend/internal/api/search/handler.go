//file: backend/handlers/search.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"log"
	"strconv"
	"veza-backend/models"
	"veza-backend/middleware"
	"github.com/jmoiron/sqlx"
)

type SearchResult struct {
	Users           []models.User            `json:"users"`
	Products        []models.Product         `json:"products"`
	Tracks          []models.Track           `json:"tracks"`
	SharedResources []models.SharedResource  `json:"shared_resources"`
	Files           []models.File            `json:"files"`
	InternalDocs    []models.InternalRessource `json:"internal_documents"`
	Messages        []models.Message         `json:"messages"`
	Query           string                   `json:"query"`
	TotalResults    int                      `json:"total_results"`
}

type AutocompleteResult struct {
	Tags      []string `json:"tags"`
	Artists   []string `json:"artists"`
	Products  []string `json:"products"`
	Users     []string `json:"users"`
}

func GlobalSearchHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			http.Error(w, "Missing search query", http.StatusBadRequest)
			return
		}

		// Limite de sécurité pour éviter les requêtes trop larges
		if len(query) < 2 {
			http.Error(w, "Query too short (minimum 2 characters)", http.StatusBadRequest)
			return
		}

		// Pagination
		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		pattern := "%" + strings.ToLower(query) + "%"
		
		var result SearchResult
		result.Query = query

		// Recherche utilisateurs (publique)
		var users []models.User
		err := db.Select(&users, `
			SELECT id, username, email, created_at 
			FROM users 
			WHERE LOWER(username) LIKE $1 OR LOWER(email) LIKE $1 
			LIMIT $2`, pattern, limit)
		if err != nil {
			log.Printf("Error searching users: %v", err)
		} else {
			result.Users = users
		}

		// Recherche produits (seulement si l'utilisateur est connecté pour la confidentialité)
		userID := getUserIDFromContext(r)
		if userID > 0 {
			var products []models.Product
			err = db.Select(&products, `
				SELECT id, user_id, name, version, purchase_date, warranty_expires 
				FROM products 
				WHERE user_id = $1 AND LOWER(name) LIKE $2 
				LIMIT $3`, userID, pattern, limit)
			if err != nil {
				log.Printf("Error searching products: %v", err)
			} else {
				result.Products = products
			}
		}

		// Recherche tracks publiques
		var tracks []models.Track
		err = db.Select(&tracks, `
			SELECT id, title, artist, filename, created_at, duration_seconds, tags, is_public, uploader_id 
			FROM tracks 
			WHERE (LOWER(title) LIKE $1 OR LOWER(artist) LIKE $1) AND is_public = true 
			LIMIT $2`, pattern, limit)
		if err != nil {
			log.Printf("Error searching tracks: %v", err)
		} else {
			result.Tracks = tracks
		}

		// Recherche ressources partagées publiques
		var resources []models.SharedResource
		err = db.Select(&resources, `
			SELECT id, title, filename, url, type, tags, uploader_id, is_public, uploaded_at 
			FROM shared_resources 
			WHERE (LOWER(title) LIKE $1 OR LOWER(filename) LIKE $1) AND is_public = true 
			LIMIT $2`, pattern, limit)
		if err != nil {
			log.Printf("Error searching shared resources: %v", err)
		} else {
			result.SharedResources = resources
		}

		// Recherche fichiers (seulement pour l'utilisateur connecté)
		if userID > 0 {
			var files []models.File
			err = db.Select(&files, `
				SELECT f.id, f.product_id, f.filename, f.url, f.type, f.uploaded_at 
				FROM files f
				JOIN products p ON f.product_id = p.id
				WHERE p.user_id = $1 AND (LOWER(f.filename) LIKE $2 OR LOWER(f.type) LIKE $2)
				LIMIT $3`, userID, pattern, limit)
			if err != nil {
				log.Printf("Error searching files: %v", err)
			} else {
				result.Files = files
			}

			// Recherche documents internes
			var internalDocs []models.InternalRessource
			err = db.Select(&internalDocs, `
				SELECT d.id, d.product_id, d.title, d.filename, d.url, d.type, d.uploaded_at 
				FROM internal_documents d
				JOIN products p ON d.product_id = p.id
				WHERE p.user_id = $1 AND (LOWER(d.title) LIKE $2 OR LOWER(d.filename) LIKE $2 OR LOWER(d.type) LIKE $2)
				LIMIT $3`, userID, pattern, limit)
			if err != nil {
				log.Printf("Error searching internal docs: %v", err)
			} else {
				result.InternalDocs = internalDocs
			}
		}

		// Calcul du total des résultats
		result.TotalResults = len(result.Users) + len(result.Products) + len(result.Tracks) + 
			len(result.SharedResources) + len(result.Files) + len(result.InternalDocs)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func AdvancedSearchHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		searchType := r.URL.Query().Get("type")
		tag := r.URL.Query().Get("tag")
		//author := r.URL.Query().Get("author")
		
		userID := getUserIDFromContext(r)
		if userID == 0 {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		var result SearchResult
		result.Query = query

		// Construction dynamique de la requête selon les filtres
		switch searchType {
		case "tracks":
			var tracks []models.Track
			queryStr := `SELECT id, title, artist, filename, created_at, duration_seconds, tags, is_public, uploader_id FROM tracks WHERE is_public = true`
			params := []interface{}{}
			paramCount := 0

			if query != "" {
				paramCount++
				queryStr += ` AND (LOWER(title) LIKE $` + strconv.Itoa(paramCount) + ` OR LOWER(artist) LIKE $` + strconv.Itoa(paramCount) + `)`
				params = append(params, "%"+strings.ToLower(query)+"%")
			}

			if tag != "" {
				paramCount++
				queryStr += ` AND $` + strconv.Itoa(paramCount) + ` = ANY(tags)`
				params = append(params, tag)
			}

			queryStr += ` LIMIT 50`
			
			err := db.Select(&tracks, queryStr, params...)
			if err != nil {
				log.Printf("Error in advanced track search: %v", err)
			} else {
				result.Tracks = tracks
			}

		case "shared_resources":
			var resources []models.SharedResource
			queryStr := `SELECT id, title, filename, url, type, tags, uploader_id, is_public, uploaded_at FROM shared_resources WHERE is_public = true`
			params := []interface{}{}
			paramCount := 0

			if query != "" {
				paramCount++
				queryStr += ` AND (LOWER(title) LIKE $` + strconv.Itoa(paramCount) + ` OR LOWER(filename) LIKE $` + strconv.Itoa(paramCount) + `)`
				params = append(params, "%"+strings.ToLower(query)+"%")
			}

			if tag != "" {
				paramCount++
				queryStr += ` AND $` + strconv.Itoa(paramCount) + ` = ANY(tags)`
				params = append(params, tag)
			}

			queryStr += ` LIMIT 50`
			
			err := db.Select(&resources, queryStr, params...)
			if err != nil {
				log.Printf("Error in advanced resource search: %v", err)
			} else {
				result.SharedResources = resources
			}

		default:
			// Si pas de type spécifique, recherche globale normale
			GlobalSearchHandler(db)(w, r)
            return
		}

		result.TotalResults = len(result.Tracks) + len(result.SharedResources)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func AutocompleteHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if len(query) < 2 {
			http.Error(w, "Query too short", http.StatusBadRequest)
			return
		}

		pattern := strings.ToLower(query) + "%"
		var result AutocompleteResult

		// Tags depuis les tracks
		var trackTags []string
		err := db.Select(&trackTags, `
			SELECT DISTINCT unnest(tags) as tag 
			FROM tracks 
			WHERE is_public = true AND unnest(tags) ILIKE $1 
			LIMIT 10`, pattern)
		if err == nil {
			result.Tags = append(result.Tags, trackTags...)
		}

		// Tags depuis les ressources partagées
		var resourceTags []string
		err = db.Select(&resourceTags, `
			SELECT DISTINCT unnest(tags) as tag 
			FROM shared_resources 
			WHERE is_public = true AND unnest(tags) ILIKE $1 
			LIMIT 10`, pattern)
		if err == nil {
			result.Tags = append(result.Tags, resourceTags...)
		}

		// Artistes
		err = db.Select(&result.Artists, `
			SELECT DISTINCT artist 
			FROM tracks 
			WHERE is_public = true AND LOWER(artist) LIKE $1 
			LIMIT 10`, pattern)
		if err != nil {
			result.Artists = []string{}
		}

		// Noms d'utilisateurs
		err = db.Select(&result.Users, `
			SELECT DISTINCT username 
			FROM users 
			WHERE LOWER(username) LIKE $1 
			LIMIT 10`, pattern)
		if err != nil {
			result.Users = []string{}
		}

		// Produits (pour utilisateur connecté)
		userID := getUserIDFromContext(r)
		if userID > 0 {
			err = db.Select(&result.Products, `
				SELECT DISTINCT name 
				FROM products 
				WHERE user_id = $1 AND LOWER(name) LIKE $2 
				LIMIT 10`, userID, pattern)
			if err != nil {
				result.Products = []string{}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// Fonction utilitaire pour extraire l'ID utilisateur du contexte
func getUserIDFromContext(r *http.Request) int {
	if userID, ok := r.Context().Value(middleware.UserIDKey).(int); ok {
		return userID
	}
	return 0
}