package routes

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers"

	"github.com/gorilla/mux"
)

// RegisterWebSocketRoute registers the WebSocket route
func RegisterWebSocketRoute(router *mux.Router, db database.Service) {
	// Create WebSocket handler
	websocketHandler := handlers.NewWebSocketHandler(db)

	// Register WebSocket route directly without subrouter
	router.HandleFunc("/ws", websocketHandler.Socket)
}
