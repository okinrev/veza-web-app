// internal/handlers/tags_search.go
package handlers

import (
	"github.com/okinrev/veza-web-app/internal/common"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/middleware"
	"github.com/okinrev/veza-web-app/internal/models"
)

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
	userID, exists := common.GetUserIDFromContext(c)
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

// Ajouts pour cache et suggestions avancées

// CacheEntry represents a cached suggestion entry
type CacheEntry struct {
	Data      interface{} `json:"data"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// SuggestionCache handles caching of popular suggestions
type SuggestionCache struct {
	mu    sync.RWMutex
	cache map[string]CacheEntry
}

// NewSuggestionCache creates a new suggestion cache
func NewSuggestionCache() *SuggestionCache {
	cache := &SuggestionCache{
		cache: make(map[string]CacheEntry),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a cached entry
func (sc *SuggestionCache) Get(key string) (interface{}, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	entry, exists := sc.cache[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	
	return entry.Data, true
}

// Set stores a cache entry with expiration
func (sc *SuggestionCache) Set(key string, data interface{}, duration time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.cache[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(duration),
	}
}

// cleanup removes expired entries
func (sc *SuggestionCache) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sc.mu.Lock()
		now := time.Now()
		for key, entry := range sc.cache {
			if now.After(entry.ExpiresAt) {
				delete(sc.cache, key)
			}
		}
		sc.mu.Unlock()
	}
}

// Add cache to TagsSearchHandler
type TagsSearchHandler struct {
	db    *database.DB
	cache *SuggestionCache
}

func NewTagsSearchHandler(db *database.DB) *TagsSearchHandler {
	return &TagsSearchHandler{
		db:    db,
		cache: NewSuggestionCache(),
	}
}

// ContextualSuggestion represents context-aware suggestions
type ContextualSuggestion struct {
	Value       string            `json:"value"`
	Type        string            `json:"type"`
	Score       float64           `json:"score"`
	Context     map[string]string `json:"context,omitempty"`
	Usage       int               `json:"usage,omitempty"`
	Trending    bool              `json:"trending,omitempty"`
	Related     []string          `json:"related,omitempty"`
}

// GetContextualSuggestions provides advanced context-aware suggestions
func (h *TagsSearchHandler) GetContextualSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	context := c.Query("context") // e.g., "track", "resource", "user"
	userID, _ := common.GetUserIDFromContext(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Create cache key
	cacheKey := fmt.Sprintf("contextual_%s_%s_%d_%d", query, context, userID, limit)
	
	// Check cache first
	if cached, found := h.cache.Get(cacheKey); found {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    cached,
			"cached":  true,
		})
		return
	}

	suggestions := []ContextualSuggestion{}

	switch context {
	case "track":
		suggestions = h.getTrackContextSuggestions(query, userID, limit)
	case "resource":
		suggestions = h.getResourceContextSuggestions(query, userID, limit)
	case "user":
		suggestions = h.getUserContextSuggestions(query, userID, limit)
	default:
		suggestions = h.getGlobalContextSuggestions(query, userID, limit)
	}

	// Cache results for 5 minutes
	h.cache.Set(cacheKey, suggestions, 5*time.Minute)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    suggestions,
		"cached":  false,
	})
}

// getTrackContextSuggestions returns suggestions specific to track context
func (h *TagsSearchHandler) getTrackContextSuggestions(query string, userID, limit int) []ContextualSuggestion {
	suggestions := []ContextualSuggestion{}
	pattern := strings.ToLower(query) + "%"

	// Get popular track tags with usage stats
	rows, err := h.db.Query(`
		SELECT tag, COUNT(*) as usage, 
		       CASE WHEN COUNT(*) > (SELECT AVG(cnt) FROM (
		           SELECT COUNT(*) as cnt FROM (
		               SELECT unnest(tags) as tag FROM tracks WHERE is_public = true
		           ) t GROUP BY tag
		       ) avg_table) THEN true ELSE false END as trending
		FROM (
			SELECT unnest(tags) as tag FROM tracks WHERE is_public = true
		) track_tags
		WHERE LOWER(tag) LIKE $1
		GROUP BY tag
		ORDER BY usage DESC, tag ASC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var suggestion ContextualSuggestion
			var trending bool
			rows.Scan(&suggestion.Value, &suggestion.Usage, &trending)
			suggestion.Type = "track_tag"
			suggestion.Score = float64(suggestion.Usage) / 100.0 // Normalize score
			suggestion.Trending = trending
			suggestion.Context = map[string]string{"source": "tracks"}
			
			// Get related tags
			suggestion.Related = h.getRelatedTags(suggestion.Value, "tracks", 3)
			
			suggestions = append(suggestions, suggestion)
		}
	}

	// Add artist suggestions
	artistRows, err := h.db.Query(`
		SELECT artist, COUNT(*) as usage
		FROM tracks 
		WHERE is_public = true AND LOWER(artist) LIKE $1
		GROUP BY artist
		ORDER BY usage DESC, artist ASC
		LIMIT $2
	`, pattern, limit/2)

	if err == nil {
		defer artistRows.Close()
		for artistRows.Next() {
			var suggestion ContextualSuggestion
			artistRows.Scan(&suggestion.Value, &suggestion.Usage)
			suggestion.Type = "artist"
			suggestion.Score = float64(suggestion.Usage) / 50.0
			suggestion.Context = map[string]string{"source": "tracks"}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

// getResourceContextSuggestions returns suggestions specific to shared resources
func (h *TagsSearchHandler) getResourceContextSuggestions(query string, userID, limit int) []ContextualSuggestion {
	suggestions := []ContextualSuggestion{}
	pattern := strings.ToLower(query) + "%"

	// Get resource tags with download stats
	rows, err := h.db.Query(`
		SELECT tag, COUNT(*) as usage, AVG(sr.download_count) as avg_downloads
		FROM (
			SELECT unnest(tags) as tag, id FROM shared_resources WHERE is_public = true
		) resource_tags
		JOIN shared_resources sr ON sr.id = resource_tags.id
		WHERE LOWER(tag) LIKE $1
		GROUP BY tag
		ORDER BY avg_downloads DESC, usage DESC, tag ASC
		LIMIT $2
	`, pattern, limit)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var suggestion ContextualSuggestion
			var avgDownloads float64
			rows.Scan(&suggestion.Value, &suggestion.Usage, &avgDownloads)
			suggestion.Type = "resource_tag"
			suggestion.Score = avgDownloads / 10.0 // Score based on download popularity
			suggestion.Context = map[string]string{
				"source":        "shared_resources",
				"avg_downloads": fmt.Sprintf("%.1f", avgDownloads),
			}
			suggestion.Related = h.getRelatedTags(suggestion.Value, "shared_resources", 3)
			suggestions = append(suggestions, suggestion)
		}
	}

	// Add resource type suggestions
	typeRows, err := h.db.Query(`
		SELECT type, COUNT(*) as usage, AVG(download_count) as avg_downloads
		FROM shared_resources 
		WHERE is_public = true AND LOWER(type) LIKE $1
		GROUP BY type
		ORDER BY avg_downloads DESC, usage DESC
		LIMIT $2
	`, pattern, limit/2)

	if err == nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var suggestion ContextualSuggestion
			var avgDownloads float64
			typeRows.Scan(&suggestion.Value, &suggestion.Usage, &avgDownloads)
			suggestion.Type = "resource_type"
			suggestion.Score = avgDownloads / 5.0
			suggestion.Context = map[string]string{
				"source":        "shared_resources",
				"avg_downloads": fmt.Sprintf("%.1f", avgDownloads),
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

// getUserContextSuggestions returns user-specific suggestions
func (h *TagsSearchHandler) getUserContextSuggestions(query string, userID, limit int) []ContextualSuggestion {
	suggestions := []ContextualSuggestion{}
	pattern := strings.ToLower(query) + "%"

	// Get user suggestions with activity scores
	rows, err := h.db.Query(`
		SELECT u.username, u.id,
		       COALESCE(track_count, 0) as tracks,
		       COALESCE(resource_count, 0) as resources,
		       COALESCE(total_downloads, 0) as downloads
		FROM users u
		LEFT JOIN (
			SELECT uploader_id, COUNT(*) as track_count
			FROM tracks WHERE is_public = true
			GROUP BY uploader_id
		) t ON u.id = t.uploader_id
		LEFT JOIN (
			SELECT uploader_id, COUNT(*) as resource_count, SUM(download_count) as total_downloads
			FROM shared_resources WHERE is_public = true
			GROUP BY uploader_id
		) r ON u.id = r.uploader_id
		WHERE u.role != 'deleted' AND u.id != $1 AND LOWER(u.username) LIKE $2
		ORDER BY (COALESCE(track_count, 0) + COALESCE(resource_count, 0)) DESC, total_downloads DESC
		LIMIT $3
	`, userID, pattern, limit)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var suggestion ContextualSuggestion
			var userIDInt, tracks, resources, downloads int
			rows.Scan(&suggestion.Value, &userIDInt, &tracks, &resources, &downloads)
			suggestion.Type = "user"
			suggestion.Score = float64(tracks+resources) + float64(downloads)/100.0
			suggestion.Context = map[string]string{
				"tracks":    fmt.Sprintf("%d", tracks),
				"resources": fmt.Sprintf("%d", resources),
				"downloads": fmt.Sprintf("%d", downloads),
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

// getGlobalContextSuggestions returns general suggestions across all contexts
func (h *TagsSearchHandler) getGlobalContextSuggestions(query string, userID, limit int) []ContextualSuggestion {
	suggestions := []ContextualSuggestion{}
	
	// Combine suggestions from different contexts
	trackSuggestions := h.getTrackContextSuggestions(query, userID, limit/3)
	resourceSuggestions := h.getResourceContextSuggestions(query, userID, limit/3)
	userSuggestions := h.getUserContextSuggestions(query, userID, limit/3)
	
	// Merge and sort by score
	allSuggestions := append(append(trackSuggestions, resourceSuggestions...), userSuggestions...)
	
	// Sort by score descending
	for i := 0; i < len(allSuggestions)-1; i++ {
		for j := i + 1; j < len(allSuggestions); j++ {
			if allSuggestions[i].Score < allSuggestions[j].Score {
				allSuggestions[i], allSuggestions[j] = allSuggestions[j], allSuggestions[i]
			}
		}
	}
	
	// Return top results
	if len(allSuggestions) > limit {
		suggestions = allSuggestions[:limit]
	} else {
		suggestions = allSuggestions
	}
	
	return suggestions
}

// getRelatedTags finds tags that are commonly used together
func (h *TagsSearchHandler) getRelatedTags(tag, source string, limit int) []string {
	related := []string{}
	
	var query string
	switch source {
	case "tracks":
		query = `
			SELECT unnest(tags) as related_tag, COUNT(*) as co_occurrence
			FROM tracks 
			WHERE is_public = true AND $1 = ANY(tags) AND unnest(tags) != $1
			GROUP BY related_tag
			ORDER BY co_occurrence DESC
			LIMIT $2
		`
	case "shared_resources":
		query = `
			SELECT unnest(tags) as related_tag, COUNT(*) as co_occurrence
			FROM shared_resources 
			WHERE is_public = true AND $1 = ANY(tags) AND unnest(tags) != $1
			GROUP BY related_tag
			ORDER BY co_occurrence DESC
			LIMIT $2
		`
	default:
		return related
	}
	
	rows, err := h.db.Query(query, tag, limit)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var relatedTag string
			var count int
			rows.Scan(&relatedTag, &count)
			related = append(related, relatedTag)
		}
	}
	
	return related
}

// GetTrendingTags returns currently trending tags
func (h *TagsSearchHandler) GetTrendingTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	timeframe := c.DefaultQuery("timeframe", "7d") // 7d, 30d, 90d
	
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	// Create cache key
	cacheKey := fmt.Sprintf("trending_%s_%d", timeframe, limit)
	
	// Check cache
	if cached, found := h.cache.Get(cacheKey); found {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    cached,
			"cached":  true,
		})
		return
	}
	
	var interval string
	switch timeframe {
	case "30d":
		interval = "30 days"
	case "90d":
		interval = "90 days"
	default:
		interval = "7 days"
	}
	
	// Get trending tags from recent uploads
	rows, err := h.db.Query(`
		SELECT tag, COUNT(*) as recent_usage,
		       ROUND((COUNT(*) * 100.0 / NULLIF((
		           SELECT COUNT(*) FROM (
		               SELECT unnest(tags) FROM tracks WHERE created_at > NOW() - INTERVAL '%s'
		               UNION ALL
		               SELECT unnest(tags) FROM shared_resources WHERE uploaded_at > NOW() - INTERVAL '%s'
		           ) all_recent
		       ), 0)), 2) as trend_score
		FROM (
			SELECT unnest(tags) as tag FROM tracks 
			WHERE is_public = true AND created_at > NOW() - INTERVAL '%s'
			UNION ALL
			SELECT unnest(tags) as tag FROM shared_resources 
			WHERE is_public = true AND uploaded_at > NOW() - INTERVAL '%s'
		) recent_tags
		GROUP BY tag
		HAVING COUNT(*) > 1
		ORDER BY trend_score DESC, recent_usage DESC
		LIMIT $1
	`, interval, interval, interval, interval, limit)
	
	trending := []map[string]interface{}{}
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tag string
			var usage int
			var score float64
			rows.Scan(&tag, &usage, &score)
			
			trending = append(trending, map[string]interface{}{
				"tag":         tag,
				"usage":       usage,
				"trend_score": score,
				"timeframe":   timeframe,
			})
		}
	}
	
	// Cache for 30 minutes
	h.cache.Set(cacheKey, trending, 30*time.Minute)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    trending,
		"cached":  false,
	})
}

// ClearSuggestionCache clears the suggestion cache (admin only)
func (h *TagsSearchHandler) ClearSuggestionCache(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}
	
	// Check if user is admin
	var role string
	err := h.db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil || (role != "admin" && role != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}
	
	// Clear cache
	h.cache.mu.Lock()
	h.cache.cache = make(map[string]CacheEntry)
	h.cache.mu.Unlock()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Suggestion cache cleared successfully",
	})
}