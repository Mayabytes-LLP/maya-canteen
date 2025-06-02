package server

import (
	"fmt"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers"
	"maya-canteen/internal/server/routes"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

// Server represents the HTTP server
type Server struct {
	port int
	db   database.Service
}

// NewServer creates a new server instance
func NewServer(whatsappClient handlers.WhatsAppClient) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080 // Default port
	}

	// Create server instance
	s := &Server{
		port: port,
		db:   database.New(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      routes.RegisterRoutes(s.db, whatsappClient),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
