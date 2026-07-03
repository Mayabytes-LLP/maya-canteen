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

type Server struct {
	port           int
	db             database.Service
	whatsappClient handlers.WhatsAppClient
}

func DefaultPort() int {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		return 8080
	}
	return port
}

func NewServer(whatsappClient handlers.WhatsAppClient) *http.Server {
	port := DefaultPort()

	s := &Server{
		port:           port,
		db:             database.New(),
		whatsappClient: whatsappClient,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      routes.RegisterRoutes(s.db, s.whatsappClient),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func (s *Server) UpdateWhatsAppClient(client handlers.WhatsAppClient) {
	s.whatsappClient = client
	routes.GlobalWebSocketHandler.UpdateWhatsAppClient(client)
}