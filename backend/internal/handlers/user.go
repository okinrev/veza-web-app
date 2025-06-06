// internal/handlers/user.go
package handlers

import (
	"veza-web-app/internal/middleware"
	"veza-web-app/internal/common"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/models"
	"veza-web-app/internal/utils"
)

type UserHandler struct {
	db *database.DB
}

type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty"`
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type UserResponse struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Bio       *string `json:"bio"`
	Avatar    *string `json:"avatar"`
	Role      string  `json:"role"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{db: db}
}

// UpdateMe updates the current user's profile
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Build dynamic update query
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argCount := 1

	if req.Username != nil {
		setParts = append(setParts, "username = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Username))
		argCount++
	}
	if req.Email != nil {
		setParts = append(setParts, "email = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(strings.ToLower(*req.Email)))
		argCount++
	}
	if req.FirstName != nil {
		setParts = append(setParts, "first_name = $"+strconv.Itoa(argCount))
		args = append(args, req.FirstName)
		argCount++
	}
	if req.LastName != nil {
		setParts = append(setParts, "last_name = $"+strconv.Itoa(argCount))
		args = append(args, req.LastName)
		argCount++
	}
	if req.Bio != nil {
		setParts = append(setParts, "bio = $"+strconv.Itoa(argCount))
		args = append(args, req.Bio)
		argCount++
	}
	if req.Avatar != nil {
		setParts = append(setParts, "avatar = $"+strconv.Itoa(argCount))
		args = append(args, req.Avatar)
		argCount++
	}

	if len(setParts) == 1 { // Only updated_at
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No fields to update",
		})
		return
	}

	// Add user ID as the last argument
	args = append(args, userID)

	query := "UPDATE users SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err := h.db.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Username or email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update profile",
		})
		return
	}

	// Return updated user
	user, err := h.getUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Profile updated but failed to retrieve updated data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data":    user,
	})
}

// ChangePassword changes the user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Get current password hash
	var currentHash string
	err := h.db.QueryRow("SELECT password_hash FROM users WHERE id = $1 AND role != 'deleted'", userID).Scan(&currentHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Verify current password
	if err := utils.CheckPasswordHash(req.CurrentPassword, currentHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Current password is incorrect",
		})
		return
	}

	// Hash new password
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process new password",
		})
		return
	}

	// Update password
	_, err = h.db.Exec("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", newHash, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// GetUsers returns a paginated list of users
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := strings.TrimSpace(c.Query("search"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build search query
	baseQuery := `
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users
		WHERE role != 'deleted'
	`
	countQuery := "SELECT COUNT(*) FROM users WHERE role != 'deleted'"
	
	args := []interface{}{}
	argCount := 1
	
	if search != "" {
		searchClause := " AND (username ILIKE $" + strconv.Itoa(argCount) + " OR email ILIKE $" + strconv.Itoa(argCount) + 
			" OR first_name ILIKE $" + strconv.Itoa(argCount) + " OR last_name ILIKE $" + strconv.Itoa(argCount) + ")"
		baseQuery += searchClause
		countQuery += searchClause
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count users",
		})
		return
	}

	// Get users
	orderClause := " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+orderClause, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve users",
		})
		return
	}
	defer rows.Close()

	users := []UserResponse{}
	for rows.Next() {
		var user UserResponse
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
			&user.Bio, &user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}
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

// SearchUsers searches for users
func (h *UserHandler) SearchUsers(c *gin.Context) {
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

	rows, err := h.db.Query(`
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users 
		WHERE (username ILIKE $1 OR email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1)
		  AND role != 'deleted'
		ORDER BY username ASC
		LIMIT $2
	`, "%"+query+"%", limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search users",
		})
		return
	}
	defer rows.Close()

	users := []UserResponse{}
	for rows.Next() {
		var user UserResponse
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
			&user.Bio, &user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// GetUserByID returns a specific user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	user, err := h.getUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// Helper function to get user by ID
func (h *UserHandler) getUserByID(userID int) (*UserResponse, error) {
	var user UserResponse
	err := h.db.QueryRow(`
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users WHERE id = $1 AND role != 'deleted'
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
		&user.Bio, &user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Ajouts pour gestion avatar

const (
	MaxAvatarSize = 2 << 20 // 2MB
	AvatarWidth = 256
	AvatarHeight = 256
)

// UploadAvatar uploads and processes user avatar
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(MaxAvatarSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Avatar file too large",
		})
		return
	}

	// Get avatar file
	file, fileHeader, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Avatar file is required",
		})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "File must be an image",
		})
		return
	}

	// Validate file size
	if fileHeader.Size > MaxAvatarSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Avatar file too large (max 2MB)",
		})
		return
	}

	// Create avatars directory
	avatarsDir := "avatars"
	if err := os.MkdirAll(avatarsDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create avatars directory",
		})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("user_%d_%d%s", userID, time.Now().Unix(), ext)
	savePath := filepath.Join(avatarsDir, filename)

	// Save file
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save avatar",
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write avatar file",
		})
		return
	}

	// Update user avatar in database
	avatarURL := "/avatars/" + filename
	_, err = h.db.Exec("UPDATE users SET avatar = $1, updated_at = NOW() WHERE id = $2", avatarURL, userID)
	if err != nil {
		os.Remove(savePath) // Clean up file
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update avatar in database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Avatar uploaded successfully",
		"data": gin.H{
			"avatar_url": avatarURL,
		},
	})
}

// DeleteAvatar removes user avatar
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Get current avatar
	var currentAvatar *string
	err := h.db.QueryRow("SELECT avatar FROM users WHERE id = $1", userID).Scan(&currentAvatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get current avatar",
		})
		return
	}

	// Remove avatar from database
	_, err = h.db.Exec("UPDATE users SET avatar = NULL, updated_at = NOW() WHERE id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to remove avatar from database",
		})
		return
	}

	// Delete file if exists
	if currentAvatar != nil && *currentAvatar != "" {
		filename := filepath.Base(*currentAvatar)
		filePath := filepath.Join("avatars", filename)
		os.Remove(filePath) // Don't fail if file doesn't exist
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Avatar removed successfully",
	})
}

// ServeAvatar serves user avatar files
func (h *UserHandler) ServeAvatar(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Filename is required",
		})
		return
	}

	// Security: only allow files from avatars directory
	safePath := filepath.Join("avatars", filepath.Base(filename))
	
	if _, err := os.Stat(safePath); os.IsNotExist(err) {
		// Return default avatar if file doesn't exist
		defaultPath := filepath.Join("avatars", "default.png")
		if _, err := os.Stat(defaultPath); err == nil {
			c.File(defaultPath)
			return
		}
		
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Avatar not found",
		})
		return
	}

	// Set cache headers for avatars
	c.Header("Cache-Control", "public, max-age=86400") // 24 hours
	c.File(safePath)
}

// GetUsersExceptMe returns all users except current user (for chat)
func (h *UserHandler) GetUsersExceptMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	search := strings.TrimSpace(c.Query("search"))

	baseQuery := `
		SELECT id, username, email, first_name, last_name, avatar, role, created_at
		FROM users 
		WHERE id != $1 AND role != 'deleted'
	`
	
	args := []interface{}{userID}
	argCount := 2

	if search != "" {
		baseQuery += " AND (username ILIKE $" + strconv.Itoa(argCount) + " OR first_name ILIKE $" + strconv.Itoa(argCount) + " OR last_name ILIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+search+"%")
		argCount++
	}

	baseQuery += " ORDER BY username ASC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve users",
		})
		return
	}
	defer rows.Close()

	users := []UserResponse{}
	for rows.Next() {
		var user UserResponse
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, 
			&user.LastName, &user.Avatar, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// GetUserAvatar returns avatar URL for a user
func (h *UserHandler) GetUserAvatar(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	var avatar *string
	err = h.db.QueryRow("SELECT avatar FROM users WHERE id = $1 AND role != 'deleted'", userID).Scan(&avatar)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	avatarURL := "/avatars/default.png" // Default avatar
	if avatar != nil && *avatar != "" {
		avatarURL = *avatar
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"avatar_url": avatarURL,
		},
	})
}