package unit

import (
	"testing"
	"veza/internal/models"
)

func TestUserModelDefaults(t *testing.T) {
	u := models.User{}
	if u.IsAdmin {
		t.Errorf("Expected new user to not be admin by default")
	}
}
