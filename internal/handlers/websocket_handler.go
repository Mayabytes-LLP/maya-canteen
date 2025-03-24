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

// WhatsAppClient interface allows us to interact with the WhatsApp client
type WhatsAppClient interface {
	Logout() error
	Connect() error
	IsConnected() bool
}

type WebsocketHandler struct {
	common.BaseHandler
	upgrader         websocket.Upgrader
	clients          map[*websocket.Conn]bool
	mu               sync.Mutex
	latestWhatsappQR string         // Store the latest WhatsApp QR code
	whatsappClient   WhatsAppClient // Store reference to WhatsApp client
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

// SetWhatsAppClient allows the main app to set the WhatsApp client for interaction
func (h *WebsocketHandler) SetWhatsAppClient(client WhatsAppClient) {
	h.whatsappClient = client
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

	// If we have a stored WhatsApp QR code, send it to the new client
	if h.latestWhatsappQR != "" {
		qrMsg := WSMessage{
			Type: "whatsapp_qr",
			Payload: map[string]interface{}{
				"qr_code_base64": h.latestWhatsappQR,
			},
		}
		if err := conn.WriteJSON(qrMsg); err != nil {
			log.Printf("QR code send error: %v", err)
		}
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
			case "refresh_whatsapp":
				// Client requested a WhatsApp refresh
				log.Printf("Received WhatsApp refresh request")
				if h.whatsappClient != nil {
					if h.whatsappClient.IsConnected() {
						log.Println("WhatsApp is already connected, sending status update")
						h.Broadcast("whatsapp_status", map[string]interface{}{
							"status":  "connected",
							"message": "WhatsApp is already connected",
						})
					} else {
						log.Println("Attempting to reconnect WhatsApp...")
						// If we're not connected, try to reconnect
						// This won't generate a new QR code if credentials exist
						if err := h.whatsappClient.Connect(); err != nil {
							log.Printf("Failed to reconnect WhatsApp: %v", err)
							h.Broadcast("whatsapp_status", map[string]interface{}{
								"status":  "disconnected",
								"message": "Failed to reconnect WhatsApp: " + err.Error(),
							})
						}
					}
				} else {
					log.Println("WhatsApp client not available")
					h.Broadcast("whatsapp_status", map[string]interface{}{
						"status":  "disconnected",
						"message": "WhatsApp client not initialized",
					})
				}
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

	// If this is a WhatsApp QR code broadcast, store the latest code
	if msgType == "whatsapp_qr" {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			if qrCode, ok := payloadMap["qr_code_base64"].(string); ok && qrCode != "" {
				h.latestWhatsappQR = qrCode
			} else if qrCode == "" {
				// Reset the QR code if empty
				h.latestWhatsappQR = ""
			}
		}
	}

	message := WSMessage{
		Type:    msgType,
		Payload: payload,
	}

	deadClients := []*websocket.Conn{}

	log.Printf("Broadcasting message: %v", message)
	log.Printf("Number of connected clients: %d", len(h.clients))
	// list all the clients and send the message to each of them
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
