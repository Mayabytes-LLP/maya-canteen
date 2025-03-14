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

	// Create a separate router for WebSocket without middleware
	wsRouter := mux.NewRouter()
	RegisterWebSocketRoute(wsRouter, db)

	// Create HTTP router with middleware
	httpRouter := mux.NewRouter()
	RegisterSystemRoutes(httpRouter, db)
	RegisterTransactionRoutes(httpRouter, db)
	RegisterUserRoutes(httpRouter, db)
	RegisterProductRoutes(httpRouter, db)

	// Apply middleware to HTTP routes
	httpHandlerWithMiddleware := middleware.Chain(httpRouter, middleware.CORS(), middleware.Logger(), middleware.Recover())

	// Combine both routers
	router.PathPrefix("/ws").Handler(wsRouter)
	router.PathPrefix("/").Handler(httpHandlerWithMiddleware)

	return router
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
