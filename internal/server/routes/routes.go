package routes

import (
	"log"
	"maya-canteen/internal/database"
	"maya-canteen/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all routes for the application
func RegisterRoutes(db database.Service) http.Handler {
	// Initialize database tables
	initDatabaseTables(db)

	// Create main router
	router := mux.NewRouter()

	RegisterWebSocketRoute(router, db)

	// Create HTTP router with middleware
	RegisterSystemRoutes(router, db)
	RegisterTransactionRoutes(router, db)
	RegisterUserRoutes(router, db)
	RegisterProductRoutes(router, db)

	// Apply middleware to HTTP routes
	httpHandlerWithMiddleware := middleware.Chain(router, middleware.CORS(), middleware.Logger(), middleware.Recover())

	return httpHandlerWithMiddleware
}

// initDatabaseTables initializes all database tables
func initDatabaseTables(db database.Service) {
	// Initialize user table
	if err := db.InitUserTable(); err != nil {
		log.Fatal(err)
	}

	// Initialize transaction table
	if err := db.InitTransactionTable(); err != nil {
		log.Fatal(err)
	}

	// Initialize product table
	if err := db.InitProductTable(); err != nil {
		log.Fatal(err)
	}
}
