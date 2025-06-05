// internal/handlers/admin.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"context"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/api/middleware"
	"veza-web-app/internal/models"
)

// Constants for roles, statuses, and pagination
const (
	RoleAdmin         = "admin"
	RoleSuperAdmin    = "super_admin"
	ListingStatusOpen = "open"
	OfferStatusPending  = "pending"
	DefaultPage       = 1
	DefaultLimit      = 20
	MaxLimit          = 100
)

type AdminHandler struct {
	db *database.DB
}

type DashboardStats struct {
	TotalUsers          int     `json:"total_users"`
	ActiveUsers         int     `json:"active_users"`
	TotalTracks         int     `json:"total_tracks"`
	PublicTracks        int     `json:"public_tracks"`
	TotalSharedResources int    `json:"total_shared_resources"`
	TotalListings       int     `json:"total_listings"`
	ActiveListings      int     `json:"active_listings"`
	TotalOffers         int     `json:"total_offers"`
	PendingOffers       int     `json:"pending_offers"`
	TotalMessages       int     `json:"total_messages"`
	TotalRooms          int     `json:"total_rooms"`
	LastUpdated         string  `json:"last_updated"`
}

type UserAnalytics struct {
	UserID           int     `json:"user_id"`
	Username         string  `json:"username"`
	Email            string  `json:"email"`
	Role             string  `json:"role"`
	TracksCount      int     `json:"tracks_count"`
	ResourcesCount   int     `json:"resources_count"`
	ListingsCount    int     `json:"listings_count"`
	MessagesCount    int     `json:"messages_count"`
	RegistrationDate string  `json:"registration_date"`
	LastActivity     *string `json:"last_activity"`
	IsActive         bool    `json:"is_active"`
}

type ContentAnalytics struct {
	TracksByMonth    []MonthlyCount `json:"tracks_by_month"`
	ResourcesByMonth []MonthlyCount `json:"resources_by_month"`
	UsersByMonth     []MonthlyCount `json:"users_by_month"`
	PopularTags      []TagCount     `json:"popular_tags"`
	TopUploaders     []UploaderStats `json:"top_uploaders"`
}

type MonthlyCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type UploaderStats struct {
	UserID        int    `json:"user_id"`
	Username      string `json:"username"`
	TracksCount   int    `json:"tracks_count"`
	ResourcesCount int   `json:"resources_count"`
	TotalUploads  int    `json:"total_uploads"`
}

type CategoryResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  *string `json:"description"`
	Icon         *string `json:"icon"`
	Color        *string `json:"color"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
	ProductCount int    `json:"product_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
	SortOrder   int     `json:"sort_order"`
	IsActive    bool    `json:"is_active"`
}

func NewAdminHandler(db *database.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// Dashboard returns admin dashboard statistics
func (h *AdminHandler) Dashboard(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Verify admin access
	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	stats := DashboardStats{}

	// Get user statistics
	h.db.QueryRowContext(c.Request.Context(), "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	h.db.QueryRowContext(c.Request.Context(), "SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '30 days'").Scan(&stats.ActiveUsers)
	// ... apply to all other database calls

	// Get user statistics
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '30 days'").Scan(&stats.ActiveUsers)

	// Get track statistics
	h.db.QueryRow("SELECT COUNT(*) FROM tracks").Scan(&stats.TotalTracks)
	h.db.QueryRow("SELECT COUNT(*) FROM tracks WHERE is_public = true").Scan(&stats.PublicTracks)

	// Get shared resources statistics
	h.db.QueryRow("SELECT COUNT(*) FROM shared_resources").Scan(&stats.TotalSharedResources)

	// Get listing statistics
	h.db.QueryRow("SELECT COUNT(*) FROM listings").Scan(&stats.TotalListings)
	h.db.QueryRow("SELECT COUNT(*) FROM listings WHERE status = 'open'").Scan(&stats.ActiveListings)

	// Get offer statistics
	h.db.QueryRow("SELECT COUNT(*) FROM offers").Scan(&stats.TotalOffers)
	h.db.QueryRow("SELECT COUNT(*) FROM offers WHERE status = 'pending'").Scan(&stats.PendingOffers)

	// Get message statistics
	h.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&stats.TotalMessages)

	// Get room statistics
	h.db.QueryRow("SELECT COUNT(*) FROM rooms").Scan(&stats.TotalRooms)

	stats.LastUpdated = time.Now().Format("2006-01-02 15:04:05") // Or time.RFC3339 for ISO 8601

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUsers returns paginated list of users for admin
func (h *AdminHandler) GetUsers(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))
	search := strings.TrimSpace(c.Query("search"))
	role := c.Query("role")

	if page < DefaultPage {
		page = DefaultPage
	}
	if limit < DefaultLimit || limit > MaxLimit { // Use MaxLimit constant
		limit = DefaultLimit
	}

	offset := (page - 1) * limit

	// Build query
	baseQuery := `
		SELECT u.id, u.username, u.email, u.role, u.created_at, u.updated_at,
		       COALESCE(t.tracks_count, 0) as tracks_count,
		       COALESCE(sr.resources_count, 0) as resources_count,
		       COALESCE(l.listings_count, 0) as listings_count,
		       COALESCE(m.messages_count, 0) as messages_count
		FROM users u
		LEFT JOIN (SELECT uploader_id, COUNT(*) as tracks_count FROM tracks GROUP BY uploader_id) t ON u.id = t.uploader_id
		LEFT JOIN (SELECT uploader_id, COUNT(*) as resources_count FROM shared_resources GROUP BY uploader_id) sr ON u.id = sr.uploader_id
		LEFT JOIN (SELECT user_id, COUNT(*) as listings_count FROM listings GROUP BY user_id) l ON u.id = l.user_id
		LEFT JOIN (SELECT from_user, COUNT(*) as messages_count FROM messages GROUP BY from_user) m ON u.id = m.from_user
	`

	countQuery := "SELECT COUNT(*) FROM users u"
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if search != "" {
		whereClause += " AND (LOWER(u.username) LIKE $" + strconv.Itoa(argCount) + " OR LOWER(u.email) LIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+strings.ToLower(search)+"%")
		argCount++
	}

	if role != "" {
		whereClause += " AND u.role = $" + strconv.Itoa(argCount)
		args = append(args, role)
		argCount++
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery+" "+whereClause, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count users",
		})
		return
	}

	// Get users
	orderClause := " ORDER BY u.created_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+" "+whereClause+orderClause, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve users",
		})
		return
	}
	defer rows.Close()

	users := []UserAnalytics{}
	for rows.Next() {
		var user UserAnalytics
		err := rows.Scan(
			&user.UserID, &user.Username, &user.Email, &user.Role,
			&user.RegistrationDate, &user.LastActivity, &user.TracksCount,
			&user.ResourcesCount, &user.ListingsCount, &user.MessagesCount,
		)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}
		user.IsActive = user.TracksCount > 0 || user.ResourcesCount > 0 || user.MessagesCount > 0
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetAnalytics returns content analytics
func (h *AdminHandler) GetAnalytics(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	analytics := ContentAnalytics{}

	// Get tracks by month (last 12 months)
	trackRows, err := h.db.Query(`
		SELECT TO_CHAR(created_at, 'YYYY-MM') as month, COUNT(*) as count
		FROM tracks 
		WHERE created_at >= NOW() - INTERVAL '12 months'
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month ASC
	`)
	if err == nil {
		defer trackRows.Close()
		for trackRows.Next() {
			var mc MonthlyCount
			trackRows.Scan(&mc.Month, &mc.Count)
			analytics.TracksByMonth = append(analytics.TracksByMonth, mc)
		}
	}

	// Get resources by month (last 12 months)
	resourceRows, err := h.db.Query(`
		SELECT TO_CHAR(uploaded_at, 'YYYY-MM') as month, COUNT(*) as count
		FROM shared_resources 
		WHERE uploaded_at >= NOW() - INTERVAL '12 months'
		GROUP BY TO_CHAR(uploaded_at, 'YYYY-MM')
		ORDER BY month ASC
	`)
	if err == nil {
		defer resourceRows.Close()
		for resourceRows.Next() {
			var mc MonthlyCount
			resourceRows.Scan(&mc.Month, &mc.Count)
			analytics.ResourcesByMonth = append(analytics.ResourcesByMonth, mc)
		}
	}

	// Get users by month (last 12 months)
	userRows, err := h.db.Query(`
		SELECT TO_CHAR(created_at, 'YYYY-MM') as month, COUNT(*) as count
		FROM users 
		WHERE created_at >= NOW() - INTERVAL '12 months'
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month ASC
	`)
	if err == nil {
		defer userRows.Close()
		for userRows.Next() {
			var mc MonthlyCount
			userRows.Scan(&mc.Month, &mc.Count)
			analytics.UsersByMonth = append(analytics.UsersByMonth, mc)
		}
	}

	// Get popular tags
	tagRows, err := h.db.Query(`
		SELECT tag, COUNT(*) as count
		FROM (
			SELECT unnest(tags) as tag FROM tracks WHERE is_public = true
			UNION ALL
			SELECT unnest(tags) as tag FROM shared_resources WHERE is_public = true
		) all_tags
		GROUP BY tag
		ORDER BY count DESC
		LIMIT 20
	`)
	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var tc TagCount
			tagRows.Scan(&tc.Tag, &tc.Count)
			analytics.PopularTags = append(analytics.PopularTags, tc)
		}
	}

	// Get top uploaders
	uploaderRows, err := h.db.Query(`
		SELECT u.id, u.username, 
		       COALESCE(t.tracks_count, 0) as tracks_count,
		       COALESCE(sr.resources_count, 0) as resources_count,
		       (COALESCE(t.tracks_count, 0) + COALESCE(sr.resources_count, 0)) as total_uploads
		FROM users u
		LEFT JOIN (SELECT uploader_id, COUNT(*) as tracks_count FROM tracks GROUP BY uploader_id) t ON u.id = t.uploader_id
		LEFT JOIN (SELECT uploader_id, COUNT(*) as resources_count FROM shared_resources GROUP BY uploader_id) sr ON u.id = sr.uploader_id
		WHERE (COALESCE(t.tracks_count, 0) + COALESCE(sr.resources_count, 0)) > 0
		ORDER BY total_uploads DESC
		LIMIT 10
	`)
	if err == nil {
		defer uploaderRows.Close()
		for uploaderRows.Next() {
			var us UploaderStats
			uploaderRows.Scan(&us.UserID, &us.Username, &us.TracksCount, &us.ResourcesCount, &us.TotalUploads)
			analytics.TopUploaders = append(analytics.TopUploaders, us)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetCategories returns all categories
func (h *AdminHandler) GetCategories(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	rows, err := h.db.Query(`
		SELECT c.id, c.name, c.description, c.icon, c.color, c.sort_order, c.is_active, c.created_at, c.updated_at,
		       COALESCE(p.count, 0) as product_count
		FROM categories c
		LEFT JOIN (
			SELECT category_id, COUNT(*) as count 
			FROM products 
			WHERE category_id IS NOT NULL
			GROUP BY category_id
		) p ON c.id = p.category_id
		ORDER BY c.sort_order ASC, c.name ASC
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve categories",
		})
		return
	}
	defer rows.Close()

	categories := []CategoryResponse{}
	for rows.Next() {
		var category CategoryResponse
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.Color, &category.SortOrder, &category.IsActive, &category.CreatedAt,
			&category.UpdatedAt, &category.ProductCount,
		)
		if err != nil {
			log.Printf("Error scanning categories row: %v", err)
			continue
		}
		categories = append(categories, category)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

// CreateCategory creates a new category
func (h *AdminHandler) CreateCategory(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Check if category name already exists
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM categories WHERE LOWER(name) = LOWER($1)", req.Name).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Category name already exists",
		})
		return
	}

	// Create category
	var categoryID int
	err = h.db.QueryRow(`
		INSERT INTO categories (name, description, icon, color, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, req.Name, req.Description, req.Icon, req.Color, req.SortOrder, req.IsActive).Scan(&categoryID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create category",
		})
		return
	}

	// Return created category
	var category CategoryResponse
	err = h.db.QueryRow(`
		SELECT id, name, description, icon, color, sort_order, is_active, created_at, updated_at
		FROM categories WHERE id = $1
	`, categoryID).Scan(
		&category.ID, &category.Name, &category.Description, &category.Icon,
		&category.Color, &category.SortOrder, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Category created but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Category created successfully",
		"data":    category,
	})
}

// UpdateCategory updates a category
func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	idStr := c.Param("id")
	categoryID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Update category
	_, err = h.db.Exec(`
		UPDATE categories 
		SET name = $1, description = $2, icon = $3, color = $4, sort_order = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
	`, req.Name, req.Description, req.Icon, req.Color, req.SortOrder, req.IsActive, categoryID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category updated successfully",
	})
}

// DeleteCategory deletes a category
func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	idStr := c.Param("id")
	categoryID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	// Check if category is being used
	var count int
	err = h.db.QueryRow("SELECT COUNT(*) FROM products WHERE category_id = $1", categoryID).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Cannot delete category: it is being used by products",
		})
		return
	}

	// Delete category
	_, err = h.db.Exec("DELETE FROM categories WHERE id = $1", categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category deleted successfully",
	})
}

// Helper function to check if user is admin
func (h *AdminHandler) isAdmin(userID int) bool {
	var role string
	err := h.db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		return false
	}
	return role == RoleAdmin || role == RoleSuperAdmin
}