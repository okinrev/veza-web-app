package response

import (
    "encoding/json"
    "net/http"
)

// WriteJSON écrit directement une réponse JSON
func WriteJSON(w http.ResponseWriter, data interface{}, status int) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(data)
}

// ValidationErrorJSON envoie une réponse d'erreur de validation
func ValidationErrorJSON(w http.ResponseWriter, errors map[string]string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    
    response := struct {
        Error   string            `json:"error"`
        Code    string            `json:"code"`
        Details map[string]string `json:"details"`
    }{
        Error:   "Validation failed",
        Code:    "VALIDATION_ERROR",
        Details: errors,
    }
    
    json.NewEncoder(w).Encode(response)
}