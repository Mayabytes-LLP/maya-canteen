package main

import (
	"context"
	"fmt"
	"log"
	"maya-canteen/internal/gozk"
	"maya-canteen/internal/server"
	"maya-canteen/internal/server/routes"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Setup database path to be configurable, default to current directory
func setupDBPath() string {
	dbPath := os.Getenv("BLUEPRINT_DB_URL")
	if dbPath == "" {
		// Default to current directory if not specified
		executablePath, err := os.Executable()
		if err != nil {
			log.Fatal("Failed to get executable path:", err)
		}
		return filepath.Join(filepath.Dir(executablePath), "db", "canteen.db")
	}
	return dbPath
}

func setupZKDevice(eventLogger *log.Logger) *gozk.ZK {
	port := os.Getenv("ZK_PORT")
	if port == "" {
		port = "4370"
	}
	ip := os.Getenv("ZK_IP")
	if ip == "" {
		ip = "192.168.1.153"
	}
	timezone := os.Getenv("ZK_TIMEZONE")
	if timezone == "" {
		timezone = "0"
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("Invalid port number: %v, using default port 4370", err)
		portInt = 4370
	}

	zkSocket := gozk.NewZK(ip, portInt, 0, gozk.DefaultTimezone)

	// Start a goroutine to handle connection and event capturing
	go func() {
		for {
			if err := zkSocket.Connect(); err != nil {
				log.Printf("Failed to connect to ZK device: %v. Retrying in 3 seconds...", err)
				time.Sleep(3 * time.Second)
				continue
			}

			c, err := zkSocket.LiveCapture(time.Duration(5) * time.Second)
			if err != nil {
				log.Printf("Failed to start live capture: %v. Retrying in 3 seconds...", err)
				time.Sleep(3 * time.Second)
				continue
			}
			log.Println("ZK Device Connected")
			routes.GlobalWebSocketHandler.Broadcast("device_status", map[string]interface{}{
				"status": "connected",
			})

			// Send device status every 3 seconds
			go func() {
				for {
					time.Sleep(3 * time.Second)
					routes.GlobalWebSocketHandler.Broadcast("device_status", map[string]interface{}{
						"status": "connected",
					})
				}
			}()
			// Handle events
			for event := range c {
				log.Printf("Event: %v", event)
				if event.UserID != "" {
					log.Printf("[WebSocket] Broadcasting attendance event - UserID: %v, Time: %v",
						event.UserID, event.AttendedAt)
					eventLogger.Printf("Event: %v", event)
					routes.GlobalWebSocketHandler.Broadcast("attendance_event", map[string]interface{}{
						"user_id":   event.UserID,
						"timestamp": event.AttendedAt.Format(time.RFC3339),
					})
				}
			}

			log.Println("Connection lost, attempting to reconnect...")
			routes.GlobalWebSocketHandler.Broadcast("device_status", map[string]interface{}{
				"status": "disconnected",
			})
			time.Sleep(3 * time.Second)
		}
	}()

	return zkSocket
}

func setupLogFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// minimal event handler that only processes connection-related events
func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		// Connected to WhatsApp
		fmt.Println("Connected to WhatsApp")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "connected",
			"message": "WhatsApp connected successfully",
		})
	case *events.LoggedOut:
		// Logged out from WhatsApp
		fmt.Println("Logged out from WhatsApp")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "disconnected",
			"message": "WhatsApp logged out",
		})
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_qr", map[string]interface{}{
			"qr_code_base64": "",
			"logged_in":      false,
		})
	case *events.StreamReplaced:
		fmt.Println("Stream replaced")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "disconnected",
			"message": "WhatsApp connected from another location",
		})
	default:
		fmt.Printf("Unhandled event: %v\n", v)
	}
}

func setupWhatsapp() *whatsmeow.Client {
	dbLog := waLog.Stdout("Database", "INFO", true)
	// Use a more reliable database path
	dbPath, err := filepath.Abs("./whatsapp-store.db")
	if err != nil {
		log.Printf("Error getting absolute path: %v, using default", err)
		dbPath = "file:whatsapp-store.db?_foreign_keys=on"
	} else {
		dbPath = fmt.Sprintf("file:%s?_foreign_keys=on", dbPath)
	}

	log.Printf("Using WhatsApp database at: %s", dbPath)
	container, err := sqlstore.New("sqlite3", dbPath, dbLog)
	if err != nil {
		log.Printf("Failed to connect to WhatsApp database: %v", err)
		panic(err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		log.Printf("Failed to get WhatsApp device: %v", err)
		panic(err)
	}

	// Create client with specific options to disable history sync
	client := whatsmeow.NewClient(deviceStore, nil)

	// Only add event handler for essential events (connection status)
	client.AddEventHandler(eventHandler)

	// Register the WhatsApp client with the WebSocket handler for refresh capability
	routes.GlobalWebSocketHandler.SetWhatsAppClient(client)
	routes.GlobalWebSocketHandler.RegisterQRChannelGetter(func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error) {
		return client.GetQRChannel(ctx)
	})

	// Initialize with disconnected status - we'll connect only on demand
	routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]interface{}{
		"status":  "disconnected",
		"message": "WhatsApp initialized but not connected",
	})
	routes.GlobalWebSocketHandler.Broadcast("whatsapp_qr", map[string]interface{}{
		"qr_code_base64": "",
		"logged_in":      false,
	})

	// Check if we already have credentials
	if client.Store.ID != nil {
		log.Println("WhatsApp credentials found, checking connection status...")
		// Inform the frontend that credentials exist
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "disconnected",
			"message": "WhatsApp credentials found, click refresh to connect",
		})
	} else {
		// No credentials, inform the frontend
		log.Println("No WhatsApp credentials found, will need to scan QR code")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]interface{}{
			"status":  "disconnected",
			"message": "No WhatsApp credentials found, click refresh to get QR code",
		})
	}

	return client
}

func gracefulShutdown(apiServer *http.Server, zkSocket *gozk.ZK, whatsapp *whatsmeow.Client, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	zkSocket.StopCapture()
	log.Println("ZK device capture stopped")

	zkSocket.Disconnect()
	log.Println("ZK device disconnected")

	whatsapp.Disconnect()
	log.Println("Whatsapp disconnected")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	close(done)
}

func main() {
	logFile, err := setupLogFile("zk_events.log")
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Set the database path
	os.Setenv("BLUEPRINT_DB_URL", setupDBPath())

	eventLogger := log.New(logFile, "", log.LstdFlags)
	apiServer := server.NewServer()
	zkSocket := setupZKDevice(eventLogger)
	whatsapp := setupWhatsapp()

	done := make(chan bool, 1)
	go gracefulShutdown(apiServer, zkSocket, whatsapp, done)

	// Start the API server
	log.Println("Starting API server")
	if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", apiServer.Addr, err)
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
