package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"veza/internal/api/handlers"
	"veza/internal/config"

	"github.com/gorilla/mux"
)

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/user/me", handlers.GetMe).Methods("GET")
	return r
}

func TestGetMe_Unauthorized(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/user/me", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}
}
