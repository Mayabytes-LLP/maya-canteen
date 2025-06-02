package routes

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers"

	"github.com/gorilla/mux"
)

// RegisterWhatsAppRoutes registers all routes for WhatsApp functionality
func RegisterWhatsAppRoutes(router *mux.Router, db database.Service) {
	// Create a WhatsApp handler using the global WhatsApp client getter (function, not instance)
	whatsappHandler := handlers.NewWhatsAppHandler(db, GlobalWebSocketHandler.GetWhatsAppClient)

	// Create a subrouter for WhatsApp routes
	whatsappRouter := router.PathPrefix("/api/whatsapp").Subrouter()

	// Register WhatsApp notification routes
	whatsappRouter.HandleFunc("/notify/{id}", whatsappHandler.NotifyUserBalance).Methods("POST")
	whatsappRouter.HandleFunc("/notify-all", whatsappHandler.NotifyAllUsersBalances).Methods("POST")
}
