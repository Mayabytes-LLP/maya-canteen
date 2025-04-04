package common

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response represents the standard API response structure
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    any `json:"data,omitempty"`
}

// RespondWithJSON writes a JSON response with the given status code and payload
func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	response := Response{
		Status: "success",
		Data:   payload,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithError writes a JSON error response
func RespondWithError(w http.ResponseWriter, code int, message string) {
	response := Response{
		Status:  "error",
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithSuccess writes a JSON success response with a message
func RespondWithSuccess(w http.ResponseWriter, code int, data any) {
	RespondWithJSON(w, code, data)
}

// RespondWithNotFound sends a 404 Not Found response
func RespondWithNotFound(w http.ResponseWriter) {
	RespondWithError(w, http.StatusNotFound, "Resource not found")
}

// RespondWithBadRequest sends a 400 Bad Request response with the given error message
func RespondWithBadRequest(w http.ResponseWriter, errorMsg string) {
	RespondWithError(w, http.StatusBadRequest, errorMsg)
}

// RespondWithInternalError sends a 500 Internal Server Error response
func RespondWithInternalError(w http.ResponseWriter, err error) {
	log.Printf("Internal server error: %v", err)
	RespondWithError(w, http.StatusInternalServerError, "Internal server error")
}
