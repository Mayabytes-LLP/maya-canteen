package common

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RespondWithJSON sends a JSON response with the given status code and payload
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// RespondWithError sends an error response with the given status code and error message
func RespondWithError(w http.ResponseWriter, statusCode int, errorMsg string) {
	RespondWithJSON(w, statusCode, Response{
		Success: false,
		Error:   errorMsg,
	})
}

// RespondWithSuccess sends a success response with the given status code and data
func RespondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	RespondWithJSON(w, statusCode, Response{
		Success: true,
		Data:    data,
	})
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
