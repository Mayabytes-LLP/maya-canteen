package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	Logout(ctx context.Context) error
	Connect() error
	IsConnected() bool
}

// QRChannelGetter is a function type that gets a QR channel from the WhatsApp client
type QRChannelGetter func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error)

type ClientInfo struct {
	Conn       *websocket.Conn
	LastPing   time.Time
	ID         string
	UserAgent  string
	RemoteAddr string
}

type WebsocketHandler struct {
	common.BaseHandler
	upgrader             websocket.Upgrader
	clients              map[string]*ClientInfo // Changed to map with client ID
	clientsByConn        map[*websocket.Conn]string // Reverse lookup
	mu                   sync.RWMutex     // Use RWMutex for better performance
	latestWhatsappQR     string          // Store the latest WhatsApp QR code
	whatsappClient       WhatsAppClient  // Store reference to WhatsApp client
	getQRChannel         QRChannelGetter // Function to get a QR channel
	connectionInProgress bool            // Flag to prevent multiple connection attempts
	qrTimeout            *time.Timer     // Timer to cancel QR refresh after timeout
	healthTicker         *time.Ticker    // Health check ticker
	shutdownChan         chan struct{}   // Shutdown signal
}

type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func NewWebSocketHandler(db database.Service, client WhatsAppClient) *WebsocketHandler {
	handler := &WebsocketHandler{
		BaseHandler: common.NewBaseHandler(db),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients:              make(map[string]*ClientInfo),
		clientsByConn:        make(map[*websocket.Conn]string),
		connectionInProgress: false,
		whatsappClient:       client,
		shutdownChan:         make(chan struct{}),
	}

	// Start health check routine
	handler.startHealthCheck()

	return handler
}

// generateClientID generates a unique client ID
func (h *WebsocketHandler) generateClientID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// startHealthCheck starts the health monitoring routine
func (h *WebsocketHandler) startHealthCheck() {
	h.healthTicker = time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-h.healthTicker.C:
				h.checkConnectionHealth()
			case <-h.shutdownChan:
				h.healthTicker.Stop()
				return
			}
		}
	}()
}

// checkConnectionHealth checks and cleans up stale connections
func (h *WebsocketHandler) checkConnectionHealth() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	var deadClients []string

	for clientID, client := range h.clients {
		// Check if client hasn't responded to ping in over 60 seconds
		if now.Sub(client.LastPing) > 60*time.Second {
			log.Printf("Client %s appears to be dead, removing", clientID)
			client.Conn.Close()
			deadClients = append(deadClients, clientID)
		}
	}

	// Clean up dead clients
	for _, clientID := range deadClients {
		h.removeClient(clientID)
	}

	log.Printf("Health check complete. Active connections: %d", len(h.clients))
}

// addClient adds a new client connection
func (h *WebsocketHandler) addClient(conn *websocket.Conn, r *http.Request) string {
	clientID := h.generateClientID()
	client := &ClientInfo{
		Conn:       conn,
		LastPing:   time.Now(),
		ID:         clientID,
		UserAgent:  r.Header.Get("User-Agent"),
		RemoteAddr: r.RemoteAddr,
	}

	h.mu.Lock()
	h.clients[clientID] = client
	h.clientsByConn[conn] = clientID
	h.mu.Unlock()

	log.Printf("Client %s connected from %s. Total connections: %d",
		clientID, r.RemoteAddr, len(h.clients))

	return clientID
}

// removeClient removes a client connection
func (h *WebsocketHandler) removeClient(clientID string) {
	if client, exists := h.clients[clientID]; exists {
		delete(h.clientsByConn, client.Conn)
		delete(h.clients, clientID)
		log.Printf("Client %s disconnected. Total connections: %d",
			clientID, len(h.clients))
	}
}

// getClientByConn gets client ID by connection
func (h *WebsocketHandler) getClientByConn(conn *websocket.Conn) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clientsByConn[conn]
}

// updateClientPing updates the last ping time for a client
func (h *WebsocketHandler) updateClientPing(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, exists := h.clients[clientID]; exists {
		client.LastPing = time.Now()
	}
}

// getWhatsAppClientInfo returns a summary of the WhatsApp client (platform, user, version, etc.)
func (h *WebsocketHandler) getWhatsAppClientInfo() map[string]any {
	info := map[string]any{}
	client, ok := h.whatsappClient.(*whatsmeow.Client)
	if !ok || client == nil {
		info["status"] = "not_initialized"
		return info
	}
	// Platform and user info
	if client.Store != nil && client.Store.ID != nil {
		info["platform"] = client.Store.ID.Device
		info["user"] = client.Store.ID.User
	}
	// Version info (if available)
	if client.Store != nil && client.Store.PushName != "" {
		info["push_name"] = client.Store.PushName
	}
	// Add more fields as needed (e.g., connected, etc.)
	info["connected"] = client.IsConnected()
	return info
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
	defer func() {
		conn.Close()
		// Clean up client when connection closes
		clientID := h.getClientByConn(conn)
		if clientID != "" {
			h.mu.Lock()
			h.removeClient(clientID)
			h.mu.Unlock()
		}
	}()

	// Add client with tracking info
	clientID := h.addClient(conn, r)

	// Send initial connection message
	msg := WSMessage{
		Type:    "connected",
		Payload: map[string]interface{}{
			"message":   "WebSocket connection established",
			"client_id": clientID,
		},
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	// Send connection status broadcast
	h.BroadcastConnectionStatus()

	// Keep connection alive with ping/pong
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.updateClientPing(clientID)
			if err := conn.WriteJSON(WSMessage{Type: "ping"}); err != nil {
				log.Printf("Ping error for client %s: %v", clientID, err)
				return
			}
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error for client %s: %v", clientID, err)
				}
				return
			}

			// Handle incoming messages
			var wsMsg WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("Message parse error from client %s: %v", clientID, err)
				continue
			}

			// Handle different message types
			switch wsMsg.Type {
			case "ping", "pong":
				// Client is alive, update ping time
				h.updateClientPing(clientID)
				log.Printf("Received %s from client %s", wsMsg.Type, clientID)
				continue
			case "refresh_whatsapp":
				// Client requested a WhatsApp refresh
				log.Printf("Received WhatsApp refresh request from client %s", clientID)
				h.handleWhatsAppRefresh()
				continue
			default:
				log.Printf("Unknown message type '%s' from client %s", wsMsg.Type, clientID)
			}
		}
	}
}

// handleWhatsAppRefresh handles the WhatsApp refresh request from the client
func (h *WebsocketHandler) handleWhatsAppRefresh() {
	if h.whatsappClient == nil {
		log.Println("WhatsApp client not initialized")
		h.Broadcast("whatsapp_status", map[string]any{
			"status":      "disconnected",
			"message":     "WhatsApp client not initialized",
			"client_info": h.getWhatsAppClientInfo(),
		})
		return
	}

	// Check if a connection is already in progress
	h.mu.Lock()
	if h.connectionInProgress {
		log.Println("WhatsApp connection already in progress, ignoring request")
		h.mu.Unlock()
		h.Broadcast("whatsapp_status", map[string]any{
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

	// --- NEW LOGIC: If WhatsApp credentials are stored, connect directly ---
	client := h.GetWhatsAppClient()
	if client != nil && client.Store.ID != nil {
		// Credentials are stored, try to connect directly
		log.Println("WhatsApp credentials found, connecting directly...")
		h.Broadcast("whatsapp_status", map[string]any{
			"status":      "connecting",
			"message":     "Connecting to WhatsApp with stored credentials...",
			"client_info": h.getWhatsAppClientInfo(),
		})

		if h.whatsappClient.IsConnected() {
			log.Println("WhatsApp login successful (with stored credentials)")
			h.Broadcast("whatsapp_status", map[string]any{
				"status":      "connected",
				"message":     "WhatsApp login successful",
				"client_info": h.getWhatsAppClientInfo(),
			})
			h.Broadcast("whatsapp_qr", map[string]any{
				"qr_code_base64": "",
				"logged_in":      true,
			})
		}

		go func() {
			if err := h.whatsappClient.Connect(); err != nil {
				log.Printf("Failed to connect to WhatsApp: %v", err)
				h.Broadcast("whatsapp_status", map[string]any{
					"status":      "disconnected",
					"message":     "Connection failed: " + err.Error(),
					"client_info": h.getWhatsAppClientInfo(),
				})
				return
			}
			// Wait a moment for connection to establish
			time.Sleep(2 * time.Second)
			if h.whatsappClient.IsConnected() {
				log.Println("WhatsApp login successful (with stored credentials)")
				h.Broadcast("whatsapp_status", map[string]any{
					"status":      "connected",
					"message":     "WhatsApp login successful",
					"client_info": h.getWhatsAppClientInfo(),
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      true,
				})
			} else {
				log.Println("WhatsApp connection failed (with stored credentials)")
				h.Broadcast("whatsapp_status", map[string]any{
					"status":      "disconnected",
					"message":     "Failed to connect with stored credentials. Please try again.",
					"client_info": h.getWhatsAppClientInfo(),
				})
			}
		}()
		return
	}
	// --- END NEW LOGIC ---

	if h.whatsappClient.IsConnected() {
		log.Println("WhatsApp is already connected")
		h.Broadcast("whatsapp_status", map[string]any{
			"status":      "connected",
			"message":     "WhatsApp is already connected",
			"client_info": h.getWhatsAppClientInfo(),
		})
		h.Broadcast("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      true,
		})
		return
	}

	log.Println("Attempting to connect to WhatsApp...")
	h.Broadcast("whatsapp_status", map[string]any{
		"status":      "disconnected",
		"message":     "Connecting to WhatsApp...",
		"client_info": h.getWhatsAppClientInfo(),
	})

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start WhatsApp connection in a goroutine
	go func() {
		if h.getQRChannel == nil {
			log.Println("QR channel getter not registered")
			h.Broadcast("whatsapp_status", map[string]any{
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
			h.Broadcast("whatsapp_status", map[string]any{
				"status":  "disconnected",
				"message": "QR code scanning timed out. Please try again.",
			})
			h.Broadcast("whatsapp_qr", map[string]any{
				"qr_code_base64": "",
				"logged_in":      false,
			})
		})
		h.mu.Unlock()

		// Get QR channel first with our cancellable context
		qrChan, err := h.getQRChannel(ctx)
		if err != nil {
			log.Printf("Failed to get QR channel: %v", err)
			h.Broadcast("whatsapp_status", map[string]any{
				"status":  "disconnected",
				"message": "Failed to initialize QR code process: " + err.Error(),
			})
			return
		}

		// Now connect - this will generate the QR code if needed
		if err := h.whatsappClient.Connect(); err != nil {
			log.Printf("Failed to connect to WhatsApp: %v", err)
			h.Broadcast("whatsapp_status", map[string]any{
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
				h.Broadcast("whatsapp_qr", map[string]any{
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
				h.Broadcast("whatsapp_status", map[string]any{
					"status":  "connected",
					"message": "WhatsApp login successful",
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      true,
				})
				break
			}
		}

		// If we didn't show a QR code and didn't connect successfully,
		// inform the user about the issue
		if !qrCodeShown && !h.whatsappClient.IsConnected() {
			h.Broadcast("whatsapp_status", map[string]any{
				"status":  "disconnected",
				"message": "Could not generate QR code. Please try again later.",
			})
		}
	}()
}

func (h *WebsocketHandler) Broadcast(msgType string, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If this is a WhatsApp QR code broadcast, store the latest code
	if msgType == "whatsapp_qr" {
		if payloadMap, ok := payload.(map[string]any); ok {
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

	var deadClients []string

	log.Printf("Broadcasting message type '%s' to %d clients", msgType, len(h.clients))

	// Send message to all connected clients
	for clientID, client := range h.clients {
		if err := client.Conn.WriteJSON(message); err != nil {
			log.Printf("WebSocket write error to client %s: %v", clientID, err)
			client.Conn.Close()
			deadClients = append(deadClients, clientID)
		}
	}

	// Clean up dead clients (need to re-acquire write lock)
	if len(deadClients) > 0 {
		h.mu.RUnlock() // Release read lock
		h.mu.Lock()    // Acquire write lock
		for _, clientID := range deadClients {
			h.removeClient(clientID)
		}
		h.mu.Unlock()
		h.mu.RLock() // Re-acquire read lock for defer
	}
}

// BroadcastConnectionStatus broadcasts the current connection status
func (h *WebsocketHandler) BroadcastConnectionStatus() {
	h.mu.RLock()
	connectionCount := len(h.clients)
	h.mu.RUnlock()

	h.Broadcast("connection_status", map[string]interface{}{
		"total_connections": connectionCount,
		"timestamp":        time.Now().Unix(),
	})
}

// GetConnectionStats returns statistics about current connections
func (h *WebsocketHandler) GetConnectionStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(h.clients),
		"timestamp":        time.Now().Unix(),
		"clients":          make([]map[string]interface{}, 0, len(h.clients)),
	}

	for clientID, client := range h.clients {
		clientStats := map[string]interface{}{
			"id":         clientID,
			"remote_addr": client.RemoteAddr,
			"user_agent":  client.UserAgent,
			"last_ping":   client.LastPing.Unix(),
			"connected_for": time.Since(client.LastPing).Seconds(),
		}
		stats["clients"] = append(stats["clients"].([]map[string]interface{}), clientStats)
	}

	return stats
}

// Cleanup properly shuts down the WebSocket handler
func (h *WebsocketHandler) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Signal shutdown to health checker
	close(h.shutdownChan)

	// Cancel any active QR timeout
	if h.qrTimeout != nil {
		h.qrTimeout.Stop()
		h.qrTimeout = nil
	}

	// Close all client connections
	for clientID, client := range h.clients {
		client.Conn.Close()
		log.Printf("Closed connection for client %s during cleanup", clientID)
	}

	// Clear client maps
	h.clients = make(map[string]*ClientInfo)
	h.clientsByConn = make(map[*websocket.Conn]string)

	log.Printf("WebSocket handler cleanup completed")
}

func (h *WebsocketHandler) GetWhatsAppClient() *whatsmeow.Client {
	if client, ok := h.whatsappClient.(*whatsmeow.Client); ok {
		return client
	}
	return nil
}

// UpdateWhatsAppClient updates the WhatsApp client
func (h *WebsocketHandler) UpdateWhatsAppClient(client WhatsAppClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.whatsappClient = client
}
