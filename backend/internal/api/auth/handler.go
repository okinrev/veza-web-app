package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/utils/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			response.ErrorJSON(c.Writer, err.Error(), http.StatusConflict)
			return
		}
		response.ErrorJSON(c.Writer, "Registration failed", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, map[string]interface{}{
		"user_id": user.ID,
		"username": user.Username,
		"email": user.Email,
	}, "User registered successfully")
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data: "+err.Error(), http.StatusBadRequest)
		return
	}

	loginResp, err := h.service.Login(req)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response.SuccessJSON(c.Writer, loginResp, "Login successful")
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	tokenResp, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	response.SuccessJSON(c.Writer, tokenResp, "Token refreshed")
}

func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	err := h.service.Logout(req.RefreshToken)
	if err != nil {
		response.ErrorJSON(c.Writer, "Logout failed", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, nil, "Logged out successfully")
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetMe(userID)
	if err != nil {
		response.ErrorJSON(c.Writer, "User not found", http.StatusNotFound)
		return
	}

	response.SuccessJSON(c.Writer, user, "User profile retrieved")
}
