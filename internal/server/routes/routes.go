package routes

import (
	"maya-canteen/frontend"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers"
	"maya-canteen/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes registers all routes for the application
func RegisterRoutes(db database.Service, whatsappClient handlers.WhatsAppClient) http.Handler {
	// Initialize database tables
	initDatabaseTables(db)

	// Create main router
	router := mux.NewRouter()

	RegisterWebSocketRoute(router, db, whatsappClient)

	// Create HTTP router with middleware
	RegisterSystemRoutes(router, db)
	RegisterTransactionRoutes(router, db)
	RegisterUserRoutes(router, db)
	RegisterProductRoutes(router, db)
	RegisterWhatsAppRoutes(router, db)

	// Apply middleware to HTTP routes
	httpHandlerWithMiddleware := middleware.Chain(router, middleware.CORS(), middleware.Logger(), middleware.Recover())

	return frontend.ServeStaticFiles(httpHandlerWithMiddleware)
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

	if err := db.InitTransactionProductTable(); err != nil {
		log.Fatal(err)
	}
}
