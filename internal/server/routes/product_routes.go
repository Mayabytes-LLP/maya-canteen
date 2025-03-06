package routes

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers"

	"github.com/gorilla/mux"
)

// RegisterProductRoutes registers all product-related routes
func RegisterProductRoutes(router *mux.Router, db database.Service) {
	// Create product handler
	productHandler := handlers.NewProductHandler(db)

	// Register routes
	router.HandleFunc("/api/products", productHandler.GetAllProducts).Methods("GET")
	router.HandleFunc("/api/products", productHandler.CreateProduct).Methods("POST")
	router.HandleFunc("/api/products/{id}", productHandler.GetProduct).Methods("GET")
	router.HandleFunc("/api/products/{id}", productHandler.UpdateProduct).Methods("PUT")
	router.HandleFunc("/api/products/{id}", productHandler.DeleteProduct).Methods("DELETE")
}
