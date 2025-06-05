// internal/api/message/handler.go
package message

import (
	"net/http"
	"strconv"
	"veza-web-app/internal/api/middleware"
	"veza-web-app/internal/utils/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GetDmHandler récupère les messages directs avec un utilisateur
func (h *Handler) GetDmHandler(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid user ID", http.StatusBadRequest)
		return
	}

	currentUserID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	// TODO: Récupérer les messages depuis la BDD
	messages := []map[string]interface{}{
		{
			"id":        1,
			"from_user": currentUserID,
			"to_user":   targetUserID,
			"content":   "Hello!",
			"timestamp": "2025-01-01T00:00:00Z",
		},
	}

	response.SuccessJSON(c.Writer, messages, "Direct messages retrieved")
}