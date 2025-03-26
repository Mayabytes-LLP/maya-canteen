package handlers

import (
	"context"
	"encoding/json"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// WhatsAppClient interface allows us to interact with the WhatsApp client
type WhatsAppClient interface {
	Logout() error
	Connect() error
	IsConnected() bool
}

// QRChannelGetter is a function type that gets a QR channel from the WhatsApp client
type QRChannelGetter func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error)

type WebsocketHandler struct {
	common.BaseHandler
	upgrader             websocket.Upgrader
	clients              map[*websocket.Conn]bool
	mu                   sync.Mutex
	latestWhatsappQR     string          // Store the latest WhatsApp QR code
	whatsappClient       WhatsAppClient  // Store reference to WhatsApp client
	getQRChannel         QRChannelGetter // Function to get a QR channel
	connectionInProgress bool            // Flag to prevent multiple connection attempts
	qrTimeout            *time.Timer     // Timer to cancel QR refresh after timeout
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
		clients:              make(map[*websocket.Conn]bool),
		connectionInProgress: false,
	}
}

// SetWhatsAppClient allows the main app to set the WhatsApp client for interaction
func (h *WebsocketHandler) SetWhatsAppClient(client WhatsAppClient) {
	h.whatsappClient = client
}

// GetWhatsAppClient returns the WhatsApp client for use by other handlers
func (h *WebsocketHandler) GetWhatsAppClient() *whatsmeow.Client {
	if client, ok := h.whatsappClient.(*whatsmeow.Client); ok {
		return client
	}
	return nil
}

// RegisterQRChannelGetter sets the function to get a QR channel
func (h *WebsocketHandler) RegisterQRChannelGetter(getter QRChannelGetter) {
	h.getQRChannel = getter
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
			case "refresh_whatsapp":
				// Client requested a WhatsApp refresh
				log.Printf("Received WhatsApp refresh request")
				h.handleWhatsAppRefresh()
				continue
			default:
				log.Printf("Unknown message type: %s", wsMsg.Type)
			}
		}
	}
}

// handleWhatsAppRefresh handles the WhatsApp refresh request from the client
func (h *WebsocketHandler) handleWhatsAppRefresh() {
	if h.whatsappClient == nil {
		log.Println("WhatsApp client not initialized")
		h.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "disconnected",
			"message": "WhatsApp client not initialized",
		})
		return
	}

	// Check if a connection is already in progress
	h.mu.Lock()
	if h.connectionInProgress {
		log.Println("WhatsApp connection already in progress, ignoring request")
		h.mu.Unlock()
		h.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "disconnected",
			"message": "Connection attempt already in progress",
		})
		return
	}
	h.connectionInProgress = true

	// Cancel existing QR timeout if there is one
	if h.qrTimeout != nil {
		h.qrTimeout.Stop()
	}

	h.mu.Unlock()

	// Set a timeout to reset the connection flag
	defer func() {
		time.Sleep(5 * time.Second) // Allow some time for the connection process
		h.mu.Lock()
		h.connectionInProgress = false
		h.mu.Unlock()
	}()

	if h.whatsappClient.IsConnected() {
		log.Println("WhatsApp is already connected")
		h.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "connected",
			"message": "WhatsApp is already connected",
		})
		h.Broadcast("whatsapp_qr", map[string]interface{}{
			"qr_code_base64": "",
			"logged_in":      true,
		})
		return
	}

	log.Println("Attempting to connect to WhatsApp...")
	h.Broadcast("whatsapp_status", map[string]interface{}{
		"status":  "disconnected",
		"message": "Connecting to WhatsApp...",
	})

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start WhatsApp connection in a goroutine
	go func() {
		if h.getQRChannel == nil {
			log.Println("QR channel getter not registered")
			h.Broadcast("whatsapp_status", map[string]interface{}{
				"status":  "disconnected",
				"message": "QR code generation not available",
			})
			return
		}

		// Set a timeout of 5 minutes for QR code scanning
		h.mu.Lock()
		h.qrTimeout = time.AfterFunc(5*time.Minute, func() {
			log.Println("QR code scanning timed out after 5 minutes")
			cancel() // Cancel the QR context
			h.Broadcast("whatsapp_status", map[string]interface{}{
				"status":  "disconnected",
				"message": "QR code scanning timed out. Please try again.",
			})
			h.Broadcast("whatsapp_qr", map[string]interface{}{
				"qr_code_base64": "",
				"logged_in":      false,
			})
		})
		h.mu.Unlock()

		// Get QR channel first with our cancellable context
		qrChan, err := h.getQRChannel(ctx)
		if err != nil {
			log.Printf("Failed to get QR channel: %v", err)
			h.Broadcast("whatsapp_status", map[string]interface{}{
				"status":  "disconnected",
				"message": "Failed to initialize QR code process: " + err.Error(),
			})
			return
		}

		// Now connect - this will generate the QR code if needed
		if err := h.whatsappClient.Connect(); err != nil {
			log.Printf("Failed to connect to WhatsApp: %v", err)
			h.Broadcast("whatsapp_status", map[string]interface{}{
				"status":  "disconnected",
				"message": "Connection failed: " + err.Error(),
			})
			return
		}

		// Wait for QR code or successful connection
		var qrCodeShown bool
		for evt := range qrChan {
			if evt.Event == "code" {
				qrCodeShown = true
				log.Println("WhatsApp QR code received, broadcasting to UI")
				h.Broadcast("whatsapp_qr", map[string]interface{}{
					"qr_code_base64": evt.Code,
					"logged_in":      false,
				})
			} else if evt.Event == "success" {
				// Stop the timeout timer since we've successfully connected
				h.mu.Lock()
				if h.qrTimeout != nil {
					h.qrTimeout.Stop()
					h.qrTimeout = nil
				}
				h.mu.Unlock()

				log.Println("WhatsApp login successful")
				h.Broadcast("whatsapp_status", map[string]interface{}{
					"status":  "connected",
					"message": "WhatsApp login successful",
				})
				h.Broadcast("whatsapp_qr", map[string]interface{}{
					"qr_code_base64": "",
					"logged_in":      true,
				})
				break
			}
		}

		// If we didn't show a QR code and didn't connect successfully,
		// inform the user about the issue
		if !qrCodeShown && !h.whatsappClient.IsConnected() {
			h.Broadcast("whatsapp_status", map[string]interface{}{
				"status":  "disconnected",
				"message": "Could not generate QR code. Please try again later.",
			})
		}
	}()
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

// Clean up any resources when the handler is being destroyed
func (h *WebsocketHandler) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Cancel any active QR timeout
	if h.qrTimeout != nil {
		h.qrTimeout.Stop()
		h.qrTimeout = nil
	}
}
