// internal/utils/response/response.go
package response

import (
    "encoding/json"
    "net/http"
)

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Page       int `json:"page,omitempty"`
    PerPage    int `json:"per_page,omitempty"`
    Total      int `json:"total,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}

func SuccessJSON(w http.ResponseWriter, data interface{}, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    response := APIResponse{
        Success: true,
        Data:    data,
        Message: message,
    }
    
    json.NewEncoder(w).Encode(response)
}

func ErrorJSON(w http.ResponseWriter, error string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    response := APIResponse{
        Success: false,
        Error:   error,
    }
    
    json.NewEncoder(w).Encode(response)
}

func PaginatedJSON(w http.ResponseWriter, data interface{}, meta *Meta, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    response := APIResponse{
        Success: true,
        Data:    data,
        Message: message,
        Meta:    meta,
    }
    
    json.NewEncoder(w).Encode(response)
}