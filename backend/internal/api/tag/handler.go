// internal/api/tag/handler.go
package tag

import (
	"net/http"
	"veza-web-app/internal/utils/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GetAllTags récupère tous les tags
func (h *Handler) GetAllTags(c *gin.Context) {
	// TODO: Récupérer depuis la BDD
	tags := []map[string]interface{}{
		{"id": 1, "name": "electronic"},
		{"id": 2, "name": "ambient"},
		{"id": 3, "name": "techno"},
	}

	response.SuccessJSON(c.Writer, tags, "Tags retrieved")
}

// SearchTags recherche des tags
func (h *Handler) SearchTags(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorJSON(c.Writer, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// TODO: Recherche de tags
	tags := []map[string]interface{}{
		{"id": 1, "name": query + "tronic"},
	}

	response.SuccessJSON(c.Writer, tags, "Tags search completed")
}