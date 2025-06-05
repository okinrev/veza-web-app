// internal/handlers/tags_search.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
	"veza-web-app/internal/models"
)

type TagsSearchHandler struct {
	db *database.DB
}

type GlobalSearchResult struct {
	Users          []UserSearchResult          `json:"users"`
	Tracks         []TrackSearchResult         `json:"tracks"`
	SharedResources []SharedResourceSearchResult `json:"shared_resources"`
	TotalResults   int                         `json:"total_results"`
	Query          string                      `json:"query"`
}

type UserSearchResult struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Avatar    *string `json:"avatar"`
}

type TrackSearchResult struct {
	ID              int      `json:"id"`
	Title           string   `json:"title"`
	Artist          string   `json:"artist"`
	Filename        string   `json:"filename"`
	Tags            []string `json:"tags"`
	UploaderID      int      `json:"uploader_id"`
	UploaderName    string   `json:"uploader_name"`
	CreatedAt       string   `json:"created_at"`
	StreamURL       string   `json:"stream_url"`
}

type SharedResourceSearchResult struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Description      *string  `json:"description"`
	Filename         string   `json:"filename"`
	Type             string   `json:"type"`
	Tags             []string `json:"tags"`
	UploaderID       int      `json:"uploader_id"`
	UploaderUsername string   `json:"uploader_username"`
	UploadedAt       string   `json:"uploaded_at"`
	DownloadURL      string   `json:"download_url"`
}

type AutocompleteResult struct {
	Tags     []string `json:"tags"`
	Artists  []string `json:"artists"`
	Users    []string `json:"users"`
	Types    []string `json:"types"`
}

type SuggestionResponse struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Count int    `json:"count,omitempty"`
}

func NewTagsSearchHandler(db *database.DB) *TagsSearchHandler {
	return &TagsSearchHandler{db: db}
}

// GlobalSearch performs a global search across all entities
func (h *TagsSearchHandler) GlobalSearch(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Query parameter 'q' is required",
		})
		return
	}

	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Query must be at least 2 characters long",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result := GlobalSearchResult{
		Query:           query,
		Users:           []UserSearchResult{},
		Tracks:          []TrackSearchResult{},
		SharedResources: []SharedResourceSearchResult{},
	}

	pattern := "%" + strings.ToLower(query) + "%"

	// Search users (public)
	userRows, err := h.db.Query(`
		SELECT id, username, email, first_name, last_name, avatar
		FROM users 
		WHERE LOWER(username) LIKE $1 OR LOWER(email) LIKE $1 
		   OR LOWER(first_name) LIKE $1 OR LOWER(last_name) LIKE $1
		ORDER BY username ASC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer userRows.Close()
		for userRows.Next() {
			var user UserSearchResult
			userRows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Avatar)
			result.Users = append(result.Users, user)
		}
	}

	// Search public tracks
	trackRows, err := h.db.Query(`
		SELECT t.id, t.title, t.artist, t.filename, t.tags, t.uploader_id, u.username, t.created_at
		FROM tracks t
		JOIN users u ON t.uploader_id = u.id
		WHERE t.is_public = true AND (LOWER(t.title) LIKE $1 OR LOWER(t.artist) LIKE $1)
		ORDER BY t.created_at DESC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer trackRows.Close()
		for trackRows.Next() {
			var track TrackSearchResult
			var tags pq.StringArray
			trackRows.Scan(&track.ID, &track.Title, &track.Artist, &track.Filename, &tags, 
				&track.UploaderID, &track.UploaderName, &track.CreatedAt)
			track.Tags = []string(tags)
			track.StreamURL = "/stream/" + track.Filename
			result.Tracks = append(result.Tracks, track)
		}
	}

	// Search public shared resources
	resourceRows, err := h.db.Query(`
		SELECT sr.id, sr.title, sr.description, sr.filename, sr.type, sr.tags, 
		       sr.uploader_id, u.username, sr.uploaded_at
		FROM shared_resources sr
		JOIN users u ON sr.uploader_id = u.id
		WHERE sr.is_public = true AND (LOWER(sr.title) LIKE $1 OR LOWER(sr.description) LIKE $1 OR LOWER(sr.filename) LIKE $1)
		ORDER BY sr.uploaded_at DESC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer resourceRows.Close()
		for resourceRows.Next() {
			var resource SharedResourceSearchResult
			var tags pq.StringArray
			resourceRows.Scan(&resource.ID, &resource.Title, &resource.Description, &resource.Filename, 
				&resource.Type, &tags, &resource.UploaderID, &resource.UploaderUsername, &resource.UploadedAt)
			resource.Tags = []string(tags)
			resource.DownloadURL = "/shared_resources/" + resource.Filename + "?download=true"
			result.SharedResources = append(result.SharedResources, resource)
		}
	}

	result.TotalResults = len(result.Users) + len(result.Tracks) + len(result.SharedResources)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// AdvancedSearch performs filtered search within specific categories
func (h *TagsSearchHandler) AdvancedSearch(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required for advanced search",
		})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	searchType := c.Query("type")
	tag := strings.TrimSpace(c.Query("tag"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if limit < 1 || limit > 100 {
		limit = 50
	}

	result := GlobalSearchResult{
		Query:           query,
		Users:           []UserSearchResult{},
		Tracks:          []TrackSearchResult{},
		SharedResources: []SharedResourceSearchResult{},
	}

	switch searchType {
	case "tracks":
		h.searchTracks(&result, query, tag, userID, limit)
	case "shared_resources":
		h.searchSharedResources(&result, query, tag, userID, limit)
	case "users":
		h.searchUsers(&result, query, userID, limit)
	default:
		// Perform global search if no specific type
		h.GlobalSearch(c)
		return
	}

	result.TotalResults = len(result.Users) + len(result.Tracks) + len(result.SharedResources)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetAllTags returns all available tags
func (h *TagsSearchHandler) GetAllTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 || limit > 500 {
		limit = 100
	}

	// Get tags from both tracks and shared resources
	rows, err := h.db.Query(`
		SELECT tag, COUNT(*) as count
		FROM (
			SELECT unnest(tags) as tag FROM tracks WHERE is_public = true
			UNION ALL
			SELECT unnest(tags) as tag FROM shared_resources WHERE is_public = true
		) all_tags
		GROUP BY tag
		ORDER BY count DESC, tag ASC
		LIMIT $1
	`, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve tags",
		})
		return
	}
	defer rows.Close()

	tags := []SuggestionResponse{}
	for rows.Next() {
		var tag SuggestionResponse
		rows.Scan(&tag.Value, &tag.Count)
		tag.Type = "tag"
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}

// SearchTags searches for tags with autocomplete
func (h *TagsSearchHandler) SearchTags(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Query parameter 'q' is required",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	pattern := strings.ToLower(query) + "%"

	rows, err := h.db.Query(`
		SELECT tag, COUNT(*) as count
		FROM (
			SELECT unnest(tags) as tag FROM tracks WHERE is_public = true
			UNION ALL
			SELECT unnest(tags) as tag FROM shared_resources WHERE is_public = true
		) all_tags
		WHERE LOWER(tag) LIKE $1
		GROUP BY tag
		ORDER BY count DESC, tag ASC
		LIMIT $2
	`, pattern, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search tags",
		})
		return
	}
	defer rows.Close()

	tags := []SuggestionResponse{}
	for rows.Next() {
		var tag SuggestionResponse
		rows.Scan(&tag.Value, &tag.Count)
		tag.Type = "tag"
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}

// GetAutocomplete provides autocomplete suggestions
func (h *TagsSearchHandler) GetAutocomplete(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Query parameter 'q' is required",
		})
		return
	}

	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Query must be at least 2 characters long",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	pattern := strings.ToLower(query) + "%"
	result := AutocompleteResult{
		Tags:    []string{},
		Artists: []string{},
		Users:   []string{},
		Types:   []string{},
	}

	// Get tag suggestions
	tagRows, err := h.db.Query(`
		SELECT DISTINCT unnest(tags) as tag 
		FROM (
			SELECT tags FROM tracks WHERE is_public = true
			UNION ALL
			SELECT tags FROM shared_resources WHERE is_public = true
		) all_entities
		WHERE unnest(tags) ILIKE $1
		ORDER BY tag ASC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var tag string
			tagRows.Scan(&tag)
			result.Tags = append(result.Tags, tag)
		}
	}

	// Get artist suggestions
	artistRows, err := h.db.Query(`
		SELECT DISTINCT artist 
		FROM tracks 
		WHERE is_public = true AND LOWER(artist) LIKE $1
		ORDER BY artist ASC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer artistRows.Close()
		for artistRows.Next() {
			var artist string
			artistRows.Scan(&artist)
			result.Artists = append(result.Artists, artist)
		}
	}

	// Get user suggestions
	userRows, err := h.db.Query(`
		SELECT DISTINCT username 
		FROM users 
		WHERE LOWER(username) LIKE $1
		ORDER BY username ASC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer userRows.Close()
		for userRows.Next() {
			var username string
			userRows.Scan(&username)
			result.Users = append(result.Users, username)
		}
	}

	// Get resource type suggestions
	typeRows, err := h.db.Query(`
		SELECT DISTINCT type 
		FROM shared_resources 
		WHERE is_public = true AND LOWER(type) LIKE $1
		ORDER BY type ASC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var resourceType string
			typeRows.Scan(&resourceType)
			result.Types = append(result.Types, resourceType)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetSuggestions provides context-aware suggestions
func (h *TagsSearchHandler) GetSuggestions(c *gin.Context) {
	suggestionType := c.Query("type")
	query := strings.TrimSpace(c.Query("q"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))

	if limit < 1 || limit > 50 {
		limit = 15
	}

	suggestions := []SuggestionResponse{}

	switch suggestionType {
	case "tag":
		if query != "" {
			pattern := query + "%"
			rows, err := h.db.Query(`
				SELECT tag, COUNT(*) as count
				FROM (
					SELECT unnest(tags) as tag FROM tracks WHERE is_public = true
					UNION ALL
					SELECT unnest(tags) as tag FROM shared_resources WHERE is_public = true
				) all_tags
				WHERE tag ILIKE $1
				GROUP BY tag
				ORDER BY count DESC, tag ASC
				LIMIT $2
			`, pattern, limit)

			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var suggestion SuggestionResponse
					rows.Scan(&suggestion.Value, &suggestion.Count)
					suggestion.Type = "tag"
					suggestions = append(suggestions, suggestion)
				}
			}
		}

	case "artist":
		if query != "" {
			pattern := query + "%"
			rows, err := h.db.Query(`
				SELECT artist, COUNT(*) as count
				FROM tracks 
				WHERE is_public = true AND artist ILIKE $1
				GROUP BY artist
				ORDER BY count DESC, artist ASC
				LIMIT $2
			`, pattern, limit)

			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var suggestion SuggestionResponse
					rows.Scan(&suggestion.Value, &suggestion.Count)
					suggestion.Type = "artist"
					suggestions = append(suggestions, suggestion)
				}
			}
		}

	case "user":
		if query != "" {
			pattern := query + "%"
			rows, err := h.db.Query(`
				SELECT username
				FROM users 
				WHERE username ILIKE $1
				ORDER BY username ASC
				LIMIT $2
			`, pattern, limit)

			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var suggestion SuggestionResponse
					rows.Scan(&suggestion.Value)
					suggestion.Type = "user"
					suggestions = append(suggestions, suggestion)
				}
			}
		}

	default:
		// Return popular tags by default
		rows, err := h.db.Query(`
			SELECT tag, COUNT(*) as count
			FROM (
				SELECT unnest(tags) as tag FROM tracks WHERE is_public = true
				UNION ALL
				SELECT unnest(tags) as tag FROM shared_resources WHERE is_public = true
			) all_tags
			GROUP BY tag
			ORDER BY count DESC, tag ASC
			LIMIT $1
		`, limit)

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var suggestion SuggestionResponse
				rows.Scan(&suggestion.Value, &suggestion.Count)
				suggestion.Type = "tag"
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    suggestions,
	})
}

// Helper functions for advanced search
func (h *TagsSearchHandler) searchTracks(result *GlobalSearchResult, query, tag string, userID, limit int) {
	baseQuery := `
		SELECT t.id, t.title, t.artist, t.filename, t.tags, t.uploader_id, u.username, t.created_at
		FROM tracks t
		JOIN users u ON t.uploader_id = u.id
		WHERE t.is_public = true
	`

	conditions := []string{}
	args := []interface{}{}
	argCount := 1

	if query != "" {
		conditions = append(conditions, "(LOWER(t.title) LIKE $"+strconv.Itoa(argCount)+" OR LOWER(t.artist) LIKE $"+strconv.Itoa(argCount)+")")
		args = append(args, "%"+strings.ToLower(query)+"%")
		argCount++
	}

	if tag != "" {
		conditions = append(conditions, "$"+strconv.Itoa(argCount)+" = ANY(t.tags)")
		args = append(args, tag)
		argCount++
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY t.created_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := h.db.Query(baseQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var track TrackSearchResult
			var tags pq.StringArray
			rows.Scan(&track.ID, &track.Title, &track.Artist, &track.Filename, &tags,
				&track.UploaderID, &track.UploaderName, &track.CreatedAt)
			track.Tags = []string(tags)
			track.StreamURL = "/stream/" + track.Filename
			result.Tracks = append(result.Tracks, track)
		}
	}
}

func (h *TagsSearchHandler) searchSharedResources(result *GlobalSearchResult, query, tag string, userID, limit int) {
	baseQuery := `
		SELECT sr.id, sr.title, sr.description, sr.filename, sr.type, sr.tags,
		       sr.uploader_id, u.username, sr.uploaded_at
		FROM shared_resources sr
		JOIN users u ON sr.uploader_id = u.id
		WHERE sr.is_public = true
	`

	conditions := []string{}
	args := []interface{}{}
	argCount := 1

	if query != "" {
		conditions = append(conditions, "(LOWER(sr.title) LIKE $"+strconv.Itoa(argCount)+" OR LOWER(sr.description) LIKE $"+strconv.Itoa(argCount)+")")
		args = append(args, "%"+strings.ToLower(query)+"%")
		argCount++
	}

	if tag != "" {
		conditions = append(conditions, "$"+strconv.Itoa(argCount)+" = ANY(sr.tags)")
		args = append(args, tag)
		argCount++
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY sr.uploaded_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := h.db.Query(baseQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var resource SharedResourceSearchResult
			var tags pq.StringArray
			rows.Scan(&resource.ID, &resource.Title, &resource.Description, &resource.Filename,
				&resource.Type, &tags, &resource.UploaderID, &resource.UploaderUsername, &resource.UploadedAt)
			resource.Tags = []string(tags)
			resource.DownloadURL = "/shared_resources/" + resource.Filename + "?download=true"
			result.SharedResources = append(result.SharedResources, resource)
		}
	}
}

func (h *TagsSearchHandler) searchUsers(result *GlobalSearchResult, query string, userID, limit int) {
	baseQuery := `
		SELECT id, username, email, first_name, last_name, avatar
		FROM users
		WHERE id != $1
	`

	args := []interface{}{userID}
	argCount := 2

	if query != "" {
		baseQuery += " AND (LOWER(username) LIKE $" + strconv.Itoa(argCount) + " OR LOWER(email) LIKE $" + strconv.Itoa(argCount) + " OR LOWER(first_name) LIKE $" + strconv.Itoa(argCount) + " OR LOWER(last_name) LIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+strings.ToLower(query)+"%")
		argCount++
	}

	baseQuery += " ORDER BY username ASC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := h.db.Query(baseQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var user UserSearchResult
			rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Avatar)
			result.Users = append(result.Users, user)
		}
	}
}