package routes

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers"

	"github.com/gorilla/mux"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router *mux.Router, db database.Service) {
	// Create user handler
	userHandler := handlers.NewUserHandler(db)

	// Register routes
	router.HandleFunc("/api/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	router.HandleFunc("/api/users/upload-csv", userHandler.UploadUserCSV).Methods("POST")
}
