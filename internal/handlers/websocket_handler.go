package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

type WhatsAppClient interface {
	Logout(ctx context.Context) error
	Connect() error
	IsConnected() bool
	IsLoggedIn() bool
	Disconnect()
	GetStoreID() *types.JID
	GetClient() *whatsmeow.Client
	PairPhone(ctx context.Context, phone string) (string, error)
}

type QRChannelGetter func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error)

type ClientInfo struct {
	Conn       *websocket.Conn
	writeMu    sync.Mutex
	LastPing   time.Time
	ID         string
	UserAgent  string
	RemoteAddr string
}

type WebsocketHandler struct {
	common.BaseHandler
	upgrader             websocket.Upgrader
	clients              map[string]*ClientInfo
	clientsByConn        map[*websocket.Conn]string
	mu                   sync.RWMutex
	whatsappClient       WhatsAppClient
	getQRChannel         QRChannelGetter
	connectionGeneration  uint64
	connectionInProgress bool
	pairingPhone         string
	healthTicker         *time.Ticker
	shutdownChan         chan struct{}
}

type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func NewWebSocketHandler(db database.Service, client WhatsAppClient) *WebsocketHandler {
	handler := &WebsocketHandler{
		BaseHandler:   common.NewBaseHandler(db),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:        make(map[string]*ClientInfo),
		clientsByConn:  make(map[*websocket.Conn]string),
		whatsappClient: client,
		shutdownChan:   make(chan struct{}),
	}

	handler.startHealthCheck()

	return handler
}

func (h *WebsocketHandler) generateClientID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

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

func (h *WebsocketHandler) checkConnectionHealth() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	var deadClients []string

	for clientID, client := range h.clients {
		if now.Sub(client.LastPing) > 60*time.Second {
			log.Printf("Client %s appears to be dead, removing", clientID)
			client.Conn.Close()
			deadClients = append(deadClients, clientID)
		}
	}

	for _, clientID := range deadClients {
		h.removeClient(clientID)
	}

	log.Printf("Health check complete. Active connections: %d", len(h.clients))
}

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

func (h *WebsocketHandler) removeClient(clientID string) {
	if client, exists := h.clients[clientID]; exists {
		delete(h.clientsByConn, client.Conn)
		delete(h.clients, clientID)
		log.Printf("Client %s disconnected. Total connections: %d",
			clientID, len(h.clients))
	}
}

func (h *WebsocketHandler) getClientByConn(conn *websocket.Conn) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clientsByConn[conn]
}

func (h *WebsocketHandler) updateClientPing(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, exists := h.clients[clientID]; exists {
		client.LastPing = time.Now()
	}
}

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
		clientID := h.getClientByConn(conn)
		if clientID != "" {
			h.mu.Lock()
			h.removeClient(clientID)
			h.mu.Unlock()
		}
	}()

	clientID := h.addClient(conn, r)

	h.mu.RLock()
	clientInfo := h.clients[clientID]
	h.mu.RUnlock()

	msg := WSMessage{
		Type: "connected",
		Payload: map[string]interface{}{
			"message":   "WebSocket connection established",
			"client_id": clientID,
		},
	}
	clientInfo.writeMu.Lock()
	err = clientInfo.Conn.WriteJSON(msg)
	clientInfo.writeMu.Unlock()
	if err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	h.BroadcastConnectionStatus()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.updateClientPing(clientID)
			clientInfo.writeMu.Lock()
			writeErr := clientInfo.Conn.WriteJSON(WSMessage{Type: "ping"})
			clientInfo.writeMu.Unlock()
			if writeErr != nil {
				log.Printf("Ping error for client %s: %v", clientID, writeErr)
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

			var wsMsg WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("Message parse error from client %s: %v", clientID, err)
				continue
			}

			switch wsMsg.Type {
			case "ping", "pong":
				h.updateClientPing(clientID)
				log.Printf("Received %s from client %s", wsMsg.Type, clientID)
				continue
			case "refresh_whatsapp":
				log.Printf("Received WhatsApp refresh request from client %s", clientID)
				h.handleWhatsAppRefresh()
				continue
			case "pair_phone":
				log.Printf("Received phone pairing request from client %s", clientID)
				payload, ok := wsMsg.Payload.(map[string]any)
				if !ok {
					log.Println("Invalid pair_phone payload")
					continue
				}
				phone, _ := payload["phone"].(string)
				if phone == "" {
					log.Println("Missing phone number in pair_phone request")
					h.Broadcast("whatsapp_status", map[string]any{
						"status":  "disconnected",
						"message": "Phone number is required for pairing",
					})
					continue
				}
				h.handlePhonePairing(phone)
				continue
			default:
				log.Printf("Unknown message type '%s' from client %s", wsMsg.Type, clientID)
			}
		}
	}
}

func (h *WebsocketHandler) handlePhonePairing(phone string) {
	// Validate phone number: strip non-digits, must be >6 digits and not start with 0
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.TrimPrefix(strings.TrimSpace(phone), "+"))
	if len(cleaned) <= 6 {
		h.Broadcast("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "Phone number must be more than 6 digits (use international format, e.g., 923001234567)",
		})
		return
	}
	if strings.HasPrefix(cleaned, "0") {
		h.Broadcast("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "Phone number must not start with 0 (use international format, e.g., 923001234567)",
		})
		return
	}

	h.mu.Lock()
	if h.connectionInProgress {
		h.mu.Unlock()
		h.Broadcast("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "Connection attempt already in progress. Please wait.",
		})
		return
	}
	h.pairingPhone = phone
	h.mu.Unlock()

	h.handleWhatsAppRefresh()
}

func (h *WebsocketHandler) handleWhatsAppRefresh() {
	if h.whatsappClient == nil {
		log.Println("WhatsApp client not initialized")
		h.Broadcast("whatsapp_status", map[string]any{
			"status":      "disconnected",
			"message":     "WhatsApp client not initialized",
			"client_info": h.getClientInfo(),
		})
		return
	}

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
	h.connectionGeneration++
	gen := h.connectionGeneration
	h.mu.Unlock()

	clearProgress := func() {
		h.mu.Lock()
		if h.connectionGeneration == gen {
			h.connectionInProgress = false
			h.pairingPhone = ""
		}
		h.mu.Unlock()
	}

	h.mu.RLock()
	pairingPhone := h.pairingPhone
	h.mu.RUnlock()

	if h.whatsappClient.IsConnected() {
		log.Println("WhatsApp is already connected")
		clearProgress()
		h.Broadcast("whatsapp_status", map[string]any{
			"status":      "connected",
			"message":     "WhatsApp is already connected",
			"client_info": h.getClientInfo(),
		})
		h.Broadcast("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      true,
		})
		return
	}

	// If already connected but not paired, and in phone pairing mode, call PairPhone directly
	if pairingPhone != "" && h.whatsappClient.GetStoreID() == nil {
		// Not connected and not paired — will connect and call PairPhone via QR channel below
	} else if h.whatsappClient.GetStoreID() != nil {
		log.Println("WhatsApp credentials found, connecting directly...")
		h.Broadcast("whatsapp_status", map[string]any{
			"status":      "connecting",
			"message":     "Connecting to WhatsApp with stored credentials...",
			"client_info": h.getClientInfo(),
		})

		go func() {
			defer clearProgress()
			if err := h.whatsappClient.Connect(); err != nil {
				log.Printf("Failed to connect to WhatsApp: %v", err)
				h.Broadcast("whatsapp_status", map[string]any{
					"status":      "disconnected",
					"message":     "Connection failed: " + err.Error(),
					"client_info": h.getClientInfo(),
				})
			}
		}()
		return
	}

	log.Println("Attempting to connect to WhatsApp...")
	h.Broadcast("whatsapp_status", map[string]any{
		"status":      "disconnected",
		"message":     "Connecting to WhatsApp...",
		"client_info": h.getClientInfo(),
	})

	go func() {
		defer clearProgress()

		if h.getQRChannel == nil {
			log.Println("QR channel getter not registered")
			h.Broadcast("whatsapp_status", map[string]any{
				"status":  "disconnected",
				"message": "QR code generation not available",
			})
			return
		}

		qrCtx, qrCancel := context.WithCancel(context.Background())
		defer qrCancel()
		qrChan, err := h.getQRChannel(qrCtx)
		if err != nil {
			log.Printf("Failed to get QR channel: %v", err)
			h.Broadcast("whatsapp_status", map[string]any{
				"status":  "disconnected",
				"message": "Failed to initialize QR code process: " + err.Error(),
			})
			return
		}

		if err := h.whatsappClient.Connect(); err != nil {
			log.Printf("Failed to connect to WhatsApp: %v", err)
			h.Broadcast("whatsapp_status", map[string]any{
				"status":  "disconnected",
				"message": "Connection failed: " + err.Error(),
			})
			return
		}

		var qrCodeShown bool
		var pairingCodeSent bool
		for evt := range qrChan {
			switch {
			case evt.Event == "code" && evt.Code != "":
				if pairingPhone != "" && !pairingCodeSent {
					// Phone pairing mode: call PairPhone on first QR code
					code, err := h.whatsappClient.PairPhone(context.Background(), pairingPhone)
					if err != nil {
						log.Printf("Failed to pair phone %s: %v", pairingPhone, err)
						h.Broadcast("whatsapp_status", map[string]any{
							"status":  "disconnected",
							"message": "Failed to pair phone: " + err.Error(),
						})
						return
					}
					pairingCodeSent = true
					qrCodeShown = true
					log.Printf("Phone pairing code generated for %s: %s", pairingPhone, code)
					h.Broadcast("whatsapp_pairing_code", map[string]any{
						"code":  code,
						"phone": pairingPhone,
					})
					h.Broadcast("whatsapp_status", map[string]any{
						"status":      "connecting",
						"message":     "Enter the pairing code on your phone to complete linking",
						"client_info": h.getClientInfo(),
					})
				} else if pairingPhone == "" {
					// QR mode: show the QR code
					qrCodeShown = true
					log.Println("WhatsApp QR code received, broadcasting to UI")
					h.Broadcast("whatsapp_qr", map[string]any{
						"qr_code_base64": evt.Code,
						"logged_in":      false,
					})
				}
				// In phone pairing mode after code sent, ignore subsequent QR codes
			case evt == whatsmeow.QRChannelSuccess:
				log.Println("WhatsApp login successful via QR channel")
				h.Broadcast("whatsapp_status", map[string]any{
					"status":  "connected",
					"message": "WhatsApp login successful",
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      true,
				})
				break
			case evt == whatsmeow.QRChannelTimeout:
				log.Println("WhatsApp QR code scanning timed out")
				h.Broadcast("whatsapp_status", map[string]any{
					"status":  "disconnected",
					"message": "QR code scanning timed out. Please try again.",
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      false,
				})
				break
			case evt == whatsmeow.QRChannelClientOutdated:
				log.Println("WhatsApp client is outdated")
				h.Broadcast("whatsapp_status", map[string]any{
					"status":  "disconnected",
					"message": "WhatsApp client is outdated. Please update and try again.",
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      false,
				})
				break
			case evt == whatsmeow.QRChannelErrUnexpectedEvent:
				log.Println("WhatsApp unexpected state - pairing may have already occurred")
				h.Broadcast("whatsapp_status", map[string]any{
					"status":  "disconnected",
					"message": "Unexpected connection state. Please try refreshing the QR code.",
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      false,
				})
				break
			case evt == whatsmeow.QRChannelScannedWithoutMultidevice:
				log.Println("WhatsApp QR scanned without multidevice enabled")
				h.Broadcast("whatsapp_status", map[string]any{
					"status":  "disconnected",
					"message": "Please enable linked devices in WhatsApp settings and try again.",
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      false,
				})
				break
			case evt.Event == "error":
				log.Printf("WhatsApp pairing error: %v", evt.Error)
				h.Broadcast("whatsapp_status", map[string]any{
					"status":  "disconnected",
					"message": "WhatsApp pairing error: " + evt.Error.Error(),
				})
				h.Broadcast("whatsapp_qr", map[string]any{
					"qr_code_base64": "",
					"logged_in":      false,
				})
				break
			default:
				log.Printf("QR channel event: %s", evt.Event)
			}
		}

		if !qrCodeShown && !h.whatsappClient.IsConnected() {
			h.Broadcast("whatsapp_status", map[string]any{
				"status":  "disconnected",
				"message": "Could not generate QR code. Please try again later.",
			})
		}
	}()
}

func (h *WebsocketHandler) getClientInfo() map[string]any {
	info := map[string]any{}
	client := h.whatsappClient
	if client == nil {
		info["status"] = "not_initialized"
		return info
	}
	if storeID := client.GetStoreID(); storeID != nil {
		info["platform"] = storeID.Device
		info["user"] = storeID.User
	}
	info["connected"] = client.IsConnected()
	return info
}

func (h *WebsocketHandler) Broadcast(msgType string, payload any) {
	message := WSMessage{
		Type:    msgType,
		Payload: payload,
	}

	h.mu.RLock()
	clients := make(map[string]*ClientInfo, len(h.clients))
	for id, c := range h.clients {
		clients[id] = c
	}
	log.Printf("Broadcasting message type '%s' to %d clients", msgType, len(clients))
	h.mu.RUnlock()

	var deadClients []string
	for clientID, client := range clients {
		client.writeMu.Lock()
		err := client.Conn.WriteJSON(message)
		client.writeMu.Unlock()
		if err != nil {
			log.Printf("WebSocket write error to client %s: %v", clientID, err)
			client.Conn.Close()
			deadClients = append(deadClients, clientID)
		}
	}

	if len(deadClients) > 0 {
		h.mu.Lock()
		for _, clientID := range deadClients {
			h.removeClient(clientID)
		}
		h.mu.Unlock()
	}
}

func (h *WebsocketHandler) BroadcastConnectionStatus() {
	h.mu.RLock()
	connectionCount := len(h.clients)
	h.mu.RUnlock()

	h.Broadcast("connection_status", map[string]interface{}{
		"total_connections": connectionCount,
		"timestamp":         time.Now().Unix(),
	})
}

func (h *WebsocketHandler) GetConnectionStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(h.clients),
		"timestamp":         time.Now().Unix(),
		"clients":           make([]map[string]interface{}, 0, len(h.clients)),
	}

	for clientID, client := range h.clients {
		clientStats := map[string]interface{}{
			"id":            clientID,
			"remote_addr":   client.RemoteAddr,
			"user_agent":    client.UserAgent,
			"last_ping":     client.LastPing.Unix(),
			"connected_for": time.Since(client.LastPing).Seconds(),
		}
		stats["clients"] = append(stats["clients"].([]map[string]interface{}), clientStats)
	}

	return stats
}

func (h *WebsocketHandler) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	close(h.shutdownChan)

	for clientID, client := range h.clients {
		client.Conn.Close()
		log.Printf("Closed connection for client %s during cleanup", clientID)
	}

	h.clients = make(map[string]*ClientInfo)
	h.clientsByConn = make(map[*websocket.Conn]string)

	log.Printf("WebSocket handler cleanup completed")
}

func (h *WebsocketHandler) GetWhatsAppClient() WhatsAppClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.whatsappClient
}

func (h *WebsocketHandler) UpdateWhatsAppClient(client WhatsAppClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.whatsappClient = client
}