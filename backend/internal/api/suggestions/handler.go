// internal/api/suggestions/handler.go
package suggestions

import (
	"net/http"
	"veza-web-app/internal/utils/response"  // ADD THIS
    "veza-web-app/internal/common"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/middleware"
	"veza-web-app/internal/common"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GetSuggestions récupère des suggestions
func (h *Handler) GetSuggestions(c *gin.Context) {
	suggestionType := c.Query("type")
	query := c.Query("q")

	var suggestions []map[string]interface{}

	switch suggestionType {
	case "tag":
		suggestions = []map[string]interface{}{
			{"type": "tag", "value": "electronic"},
			{"type": "tag", "value": "ambient"},
		}
	case "user":
		suggestions = []map[string]interface{}{
			{"type": "user", "value": "john_doe"},
		}
	default:
		suggestions = []map[string]interface{}{
			{"type": "general", "value": "suggestion"},
		}
	}

	response.SuccessJSON(c.Writer, suggestions, "Suggestions retrieved")
}