package routes

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"net/http"

	"github.com/gorilla/mux"
)

// SystemHandlers contains handlers for system routes
type SystemHandlers struct {
	DB database.Service
}

// NewSystemHandlers creates a new system handlers instance
func NewSystemHandlers(db database.Service) *SystemHandlers {
	return &SystemHandlers{
		DB: db,
	}
}

// RegisterSystemRoutes registers all system-related routes
func RegisterSystemRoutes(router *mux.Router, db database.Service) {
	// Create handlers
	handlers := NewSystemHandlers(db)

	// Register routes
	router.HandleFunc("/", handlers.HelloWorldHandler).Methods("GET")
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
}

// HelloWorldHandler handles the root endpoint
func (h *SystemHandlers) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	common.RespondWithSuccess(w, http.StatusOK, map[string]string{
		"message": "This api works!!!",
	})
}

// HealthHandler handles the health endpoint
func (h *SystemHandlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	common.RespondWithSuccess(w, http.StatusOK, h.DB.Health())
}
