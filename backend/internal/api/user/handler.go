package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/api/middleware"
)

// Handler handles user-related HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetUsers handles GET /api/users
func (h *Handler) GetUsers(c *gin.Context) {
	// Optional query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.Query("search")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	limitNum, err := strconv.Atoi(limit)
	if err != nil || limitNum < 1 || limitNum > 100 {
		limitNum = 10
	}

	users, total, err := h.service.GetUsers(pageNum, limitNum, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get users: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": users,
			"pagination": gin.H{
				"page":  pageNum,
				"limit": limitNum,
				"total": total,
			},
		},
	})
}

// GetUser handles GET /api/users/:id
func (h *Handler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"success": false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// UpdateUser handles PUT /api/users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}

	// Check if user can update this profile (own profile or admin)
	currentUserID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"success": false,
		})
		return
	}

	currentUserRole, _ := middleware.GetUserRoleFromContext(c)
	if currentUserID != userID && currentUserRole != "admin" && currentUserRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Not authorized to update this user",
			"success": false,
		})
		return
	}

	var updateData UpdateUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data: " + err.Error(),
			"success": false,
		})
		return
	}

	updatedUser, err := h.service.UpdateUser(userID, updateData)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"success": false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User updated successfully",
		"data":    updatedUser,
	})
}

// DeleteUser handles DELETE /api/users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}

	// Only admins can delete users
	currentUserRole, _ := middleware.GetUserRoleFromContext(c)
	if currentUserRole != "admin" && currentUserRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Admin access required to delete users",
			"success": false,
		})
		return
	}

	// Prevent deleting the current user
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot delete your own account",
			"success": false,
		})
		return
	}

	err = h.service.DeleteUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"success": false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}

// GetProfile handles GET /api/profile (current user's profile)
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"success": false,
		})
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get profile: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// UpdateProfile handles PUT /api/profile (current user's profile)
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"success": false,
		})
		return
	}

	var updateData UpdateUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data: " + err.Error(),
			"success": false,
		})
		return
	}

	updatedUser, err := h.service.UpdateUser(userID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update profile: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data":    updatedUser,
	})
}