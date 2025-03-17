package handlers

import (
	"encoding/json"
	"log"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketHandler struct {
	common.BaseHandler
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.Mutex
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewWebSocketHandler(db database.Service) *WebsocketHandler {
	return &WebsocketHandler{
		BaseHandler: common.NewBaseHandler(db),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *WebsocketHandler) Socket(w http.ResponseWriter, r *http.Request) {
	log.Printf("New WebSocket connection request from %s", r.RemoteAddr)

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
	// Send initial connection message
	msg := WSMessage{
		Type:    "connected",
		Payload: "WebSocket connection established",
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	// Keep connection alive with ping/pong
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteJSON(WSMessage{Type: "ping"}); err != nil {
				log.Printf("Ping error: %v", err)
				return
			}
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				h.mu.Lock()
				delete(h.clients, conn)
				h.mu.Unlock()
				return
			}

			// Handle incoming messages
			var wsMsg WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("Message parse error: %v", err)
				continue
			}

			// Handle different message types
			switch wsMsg.Type {
			case "ping":
				// Client is alive
				log.Printf("Received ping message")
				continue
			default:
				log.Printf("Unknown message type: %s", wsMsg.Type)
			}
		}
	}
}

func (h *WebsocketHandler) Broadcast(msgType string, payload interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	message := WSMessage{
		Type:    msgType,
		Payload: payload,
	}

	log.Printf("Broadcasting message: %+v", message)
	log.Printf("Number of connected clients: %d", len(h.clients))

	deadClients := []*websocket.Conn{}

	log.Printf("Broadcasting message: %v", message)
	// list all the clients and send the message to each of them
	log.Printf("Number of clients: %v", len(h.clients))
	for client := range h.clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			deadClients = append(deadClients, client)
		}
	}

	// Clean up dead clients
	for _, client := range deadClients {
		delete(h.clients, client)
	}
}
