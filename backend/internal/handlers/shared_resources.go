// internal/handlers/shared_resources.go
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
	"veza-web-app/internal/models"
)

type SharedResourcesHandler struct {
	db *database.DB
}

type SharedResourceResponse struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Description      *string  `json:"description"`
	Filename         string   `json:"filename"`
	URL              string   `json:"url"`
	Type             string   `json:"type"`
	Tags             []string `json:"tags"`
	UploaderID       int      `json:"uploader_id"`
	UploaderUsername string   `json:"uploader_username,omitempty"`
	IsPublic         bool     `json:"is_public"`
	DownloadCount    int      `json:"download_count"`
	UploadedAt       string   `json:"uploaded_at"`
	UpdatedAt        string   `json:"updated_at"`
	DownloadURL      string   `json:"download_url,omitempty"`
}

type CreateSharedResourceRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description *string  `json:"description"`
	Type        string   `json:"type" binding:"required"`
	Tags        []string `json:"tags"`
	IsPublic    bool     `json:"is_public"`
}

type UpdateSharedResourceRequest struct {
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Type        *string   `json:"type,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	IsPublic    *bool     `json:"is_public,omitempty"`
}

func NewSharedResourcesHandler(db *database.DB) *SharedResourcesHandler {
	return &SharedResourcesHandler{db: db}
}

// UploadSharedResource handles file upload and resource creation
func (h *SharedResourcesHandler) UploadSharedResource(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid multipart form",
		})
		return
	}

	// Get form values
	title := strings.TrimSpace(c.PostForm("title"))
	description := c.PostForm("description")
	resourceType := strings.TrimSpace(c.PostForm("type"))
	tagsStr := strings.TrimSpace(c.PostForm("tags"))
	isPublicStr := c.PostForm("is_public")

	if title == "" || resourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Title and type are required",
		})
		return
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// Parse is_public
	isPublic := true // default to public
	if isPublicStr == "false" {
		isPublic = false
	}

	// Get file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "File is required",
		})
		return
	}
	defer file.Close()

	// Create shared_resources directory if it doesn't exist
	resourcesDir := "shared_resources"
	if err := os.MkdirAll(resourcesDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create resources directory",
		})
		return
	}

	// Save file with safe name
	filename := filepath.Base(fileHeader.Filename)
	savePath := filepath.Join(resourcesDir, filename)

	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save file",
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write file",
		})
		return
	}

	// Insert resource into database
	url := "/shared_resources/" + filename
	var resourceID int
	err = h.db.QueryRow(`
		INSERT INTO shared_resources (title, description, filename, url, type, tags, uploader_id, is_public, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id
	`, title, description, filename, url, resourceType, pq.Array(tags), userID, isPublic).Scan(&resourceID)

	if err != nil {
		// Clean up file on database error
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save resource to database",
		})
		return
	}

	// Return resource data
	resource, err := h.getResourceByID(resourceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Resource uploaded but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Resource uploaded successfully",
		"data":    resource,
	})
}

// ListSharedResources returns a paginated list of shared resources
func (h *SharedResourcesHandler) ListSharedResources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	showPrivate := c.Query("show_private") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query based on permissions
	baseQuery := `
		SELECT sr.id, sr.title, sr.description, sr.filename, sr.url, sr.type, sr.tags,
		       sr.uploader_id, u.username, sr.is_public, sr.download_count, sr.uploaded_at, sr.updated_at
		FROM shared_resources sr
		JOIN users u ON sr.uploader_id = u.id
	`
	countQuery := `SELECT COUNT(*) FROM shared_resources sr`

	whereClause := ""
	args := []interface{}{}

	// Apply visibility filters
	if showPrivate {
		// Only show user's own resources if requesting private
		userID, exists := middleware.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required for private resources",
			})
			return
		}
		whereClause = " WHERE sr.uploader_id = $1"
		args = append(args, userID)
	} else {
		// Only public resources
		whereClause = " WHERE sr.is_public = true"
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count resources",
		})
		return
	}

	// Get resources
	orderClause := " ORDER BY sr.uploaded_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+whereClause+orderClause, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve resources",
		})
		return
	}
	defer rows.Close()

	resources := []SharedResourceResponse{}
	for rows.Next() {
		var resource SharedResourceResponse
		var tags pq.StringArray
		err := rows.Scan(
			&resource.ID, &resource.Title, &resource.Description, &resource.Filename,
			&resource.URL, &resource.Type, &tags, &resource.UploaderID,
			&resource.UploaderUsername, &resource.IsPublic, &resource.DownloadCount,
			&resource.UploadedAt, &resource.UpdatedAt,
		)
		if err != nil {
			continue
		}
		resource.Tags = []string(tags)
		resource.DownloadURL = fmt.Sprintf("/shared_resources/%s?download=true", resource.Filename)
		resources = append(resources, resource)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resources,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// SearchSharedResources searches resources with filters
func (h *SharedResourcesHandler) SearchSharedResources(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	resourceType := strings.TrimSpace(c.Query("type"))
	tag := strings.TrimSpace(c.Query("tag"))
	uploader := strings.TrimSpace(c.Query("uploader"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build search query
	baseQuery := `
		SELECT sr.id, sr.title, sr.description, sr.filename, sr.url, sr.type, sr.tags,
		       sr.uploader_id, u.username, sr.is_public, sr.download_count, sr.uploaded_at, sr.updated_at
		FROM shared_resources sr
		JOIN users u ON sr.uploader_id = u.id
		WHERE sr.is_public = true
	`

	conditions := []string{}
	args := []interface{}{}
	argCount := 1

	if query != "" {
		conditions = append(conditions, "(LOWER(sr.title) LIKE LOWER($"+strconv.Itoa(argCount)+") OR LOWER(sr.description) LIKE LOWER($"+strconv.Itoa(argCount)+"))")
		args = append(args, "%"+query+"%")
		argCount++
	}

	if resourceType != "" {
		conditions = append(conditions, "LOWER(sr.type) = LOWER($"+strconv.Itoa(argCount)+")")
		args = append(args, resourceType)
		argCount++
	}

	if tag != "" {
		conditions = append(conditions, "$"+strconv.Itoa(argCount)+" = ANY(sr.tags)")
		args = append(args, tag)
		argCount++
	}

	if uploader != "" {
		conditions = append(conditions, "LOWER(u.username) LIKE LOWER($"+strconv.Itoa(argCount)+")")
		args = append(args, "%"+uploader+"%")
		argCount++
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY sr.uploaded_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search resources",
		})
		return
	}
	defer rows.Close()

	resources := []SharedResourceResponse{}
	for rows.Next() {
		var resource SharedResourceResponse
		var tags pq.StringArray
		err := rows.Scan(
			&resource.ID, &resource.Title, &resource.Description, &resource.Filename,
			&resource.URL, &resource.Type, &tags, &resource.UploaderID,
			&resource.UploaderUsername, &resource.IsPublic, &resource.DownloadCount,
			&resource.UploadedAt, &resource.UpdatedAt,
		)
		if err != nil {
			continue
		}
		resource.Tags = []string(tags)
		resource.DownloadURL = fmt.Sprintf("/shared_resources/%s?download=true", resource.Filename)
		resources = append(resources, resource)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resources,
	})
}

// UpdateSharedResource updates a resource's metadata
func (h *SharedResourcesHandler) UpdateSharedResource(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	resourceID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid resource ID",
		})
		return
	}

	// Verify ownership
	var ownerID int
	err = h.db.QueryRow("SELECT uploader_id FROM shared_resources WHERE id = $1", resourceID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Resource not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to update this resource",
		})
		return
	}

	var req UpdateSharedResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Title != nil {
		setParts = append(setParts, "title = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Title))
		argCount++
	}
	if req.Description != nil {
		setParts = append(setParts, "description = $"+strconv.Itoa(argCount))
		args = append(args, req.Description)
		argCount++
	}
	if req.Type != nil {
		setParts = append(setParts, "type = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Type))
		argCount++
	}
	if req.Tags != nil {
		setParts = append(setParts, "tags = $"+strconv.Itoa(argCount))
		args = append(args, pq.Array(*req.Tags))
		argCount++
	}
	if req.IsPublic != nil {
		setParts = append(setParts, "is_public = $"+strconv.Itoa(argCount))
		args = append(args, *req.IsPublic)
		argCount++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No fields to update",
		})
		return
	}

	// Add updated_at and resource_id
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, resourceID)

	query := "UPDATE shared_resources SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update resource",
		})
		return
	}

	// Return updated resource
	resource, err := h.getResourceByID(resourceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Resource updated but failed to retrieve updated data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Resource updated successfully",
		"data":    resource,
	})
}

// DeleteSharedResource deletes a resource and its file
func (h *SharedResourcesHandler) DeleteSharedResource(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	resourceID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid resource ID",
		})
		return
	}

	// Get resource details for ownership verification and file deletion
	var ownerID int
	var filename string
	err = h.db.QueryRow("SELECT uploader_id, filename FROM shared_resources WHERE id = $1", resourceID).Scan(&ownerID, &filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Resource not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to delete this resource",
		})
		return
	}

	// Delete from database first
	_, err = h.db.Exec("DELETE FROM shared_resources WHERE id = $1", resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete resource from database",
		})
		return
	}

	// Delete file (don't fail if file doesn't exist)
	filePath := filepath.Join("shared_resources", filename)
	os.Remove(filePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Resource deleted successfully",
	})
}

// ServeSharedFile serves shared resource files
func (h *SharedResourcesHandler) ServeSharedFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Filename is required",
		})
		return
	}

	// Security: only allow files from shared_resources directory
	safePath := filepath.Join("shared_resources", filepath.Base(filename))
	
	if _, err := os.Stat(safePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "File not found",
		})
		return
	}

	// Check if download is requested
	if c.Query("download") == "true" {
		// Force download with proper headers
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		
		// Set appropriate MIME type
		ext := filepath.Ext(filename)
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			c.Header("Content-Type", mimeType)
		} else {
			c.Header("Content-Type", "application/octet-stream")
		}

		// Update download count
		h.db.Exec("UPDATE shared_resources SET download_count = download_count + 1 WHERE filename = $1", filename)
	}

	c.File(safePath)
}

// Helper function to get resource by ID with permission checking
func (h *SharedResourcesHandler) getResourceByID(resourceID, userID int) (*SharedResourceResponse, error) {
	query := `
		SELECT sr.id, sr.title, sr.description, sr.filename, sr.url, sr.type, sr.tags,
		       sr.uploader_id, u.username, sr.is_public, sr.download_count, sr.uploaded_at, sr.updated_at
		FROM shared_resources sr
		JOIN users u ON sr.uploader_id = u.id
		WHERE sr.id = $1 AND (sr.is_public = true OR sr.uploader_id = $2)
	`

	var resource SharedResourceResponse
	var tags pq.StringArray
	err := h.db.QueryRow(query, resourceID, userID).Scan(
		&resource.ID, &resource.Title, &resource.Description, &resource.Filename,
		&resource.URL, &resource.Type, &tags, &resource.UploaderID,
		&resource.UploaderUsername, &resource.IsPublic, &resource.DownloadCount,
		&resource.UploadedAt, &resource.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	resource.Tags = []string(tags)
	resource.DownloadURL = fmt.Sprintf("/shared_resources/%s?download=true", resource.Filename)
	return &resource, nil
}