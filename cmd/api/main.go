package main

import (
	"context"
	"fmt"
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

	logStd "log"

	log "github.com/sirupsen/logrus"
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
			log.Infof("Error getting executable path: %v, using default", err)
		}
		return filepath.Join(filepath.Dir(executablePath), "db", "canteen.db")
	}
	return dbPath
}

func setupZKDevice(eventLogger *logStd.Logger) *gozk.ZK {
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
		log.Infof("Error parsing ZK_PORT: %v,/n using default port 4370", err)
		portInt = 4370
	}

	zkSocket := gozk.NewZK(ip, portInt, 0, gozk.DefaultTimezone)

	// Start a goroutine to handle connection and event capturing
	go func() {
		for {
			if err := zkSocket.Connect(); err != nil {
				log.Infof("Failed to connect to ZK device: %v. Retrying in 3 seconds...", err)
				time.Sleep(3 * time.Second)
				continue
			}

			c, err := zkSocket.LiveCapture(time.Duration(5) * time.Second)
			if err != nil {
				log.Infof("Failed to start live capture: %v. Retrying in 3 seconds...", err)
				time.Sleep(3 * time.Second)
				continue
			}
			log.Info("ZK Device Connected")
			routes.GlobalWebSocketHandler.Broadcast("device_status", map[string]any{
				"status": "connected",
			})

			// Send device status every 3 seconds
			go func() {
				for {
					time.Sleep(3 * time.Second)
					routes.GlobalWebSocketHandler.Broadcast("device_status", map[string]any{
						"status": "connected",
					})
				}
			}()
			// Handle events
			for event := range c {
				log.Printf("Event: %v", event)
				if event.UserID != "" {
					log.Infof("[WebSocket] Broadcasting attendance event - UserID: %v, Time: %v", event.UserID, event.AttendedAt)
					eventLogger.Printf("Event: %v", event)
					routes.GlobalWebSocketHandler.Broadcast("attendance_event", map[string]any{
						"user_id":   event.UserID,
						"timestamp": event.AttendedAt.Format(time.RFC3339),
					})
				}
			}

			log.Info("ZK Device disconnected. Retrying in 3 seconds...")
			routes.GlobalWebSocketHandler.Broadcast("device_status", map[string]any{
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
		log.Infof("Failed to open log file: %v", err)
		return nil, err
	}
	return file, nil
}

// minimal event handler that only processes connection-related events
func eventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Connected:
		// Connected to WhatsApp
		log.Info("Connected to WhatsApp")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]any{
			"status":  "connected",
			"message": "WhatsApp connected successfully",
		})
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      true,
		})
	case *events.LoggedOut:
		// Logged out from WhatsApp
		log.Info("Logged out from WhatsApp")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "WhatsApp logged out",
		})
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      false,
		})
	case *events.StreamReplaced:
		log.Info("WhatsApp connected from another location")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "WhatsApp connected from another location",
		})
	default:
		log.Infof("Unhandled event: %v", v)
	}
}

func getWhatsappPath() string {
	// Check if the path is set in the environment variable
	dbPath, err := filepath.Abs("./whatsapp-store.db")
	if err != nil {
		log.Errorf("Error getting absolute path: %v, using default", err)
		dbPath = "file:whatsapp-store.db?_foreign_keys=on"
	} else {
		// Ensure cross-platform compatibility for file URI
		if os.PathSeparator == '\\' {
			// Windows requires a leading slash for file URIs
			dbPath = fmt.Sprintf("file:/%s?_foreign_keys=on", filepath.ToSlash(dbPath))
		} else {
			dbPath = fmt.Sprintf("file:%s?_foreign_keys=on", dbPath)
		}
	}

	return dbPath
}

func setupWhatsapp() *whatsmeow.Client {
	dbLog := waLog.Stdout("Database", "INFO", true)
	// Use a more reliable database path
	dbPath := getWhatsappPath()

	log.Infof("Using WhatsApp database at: %s", dbPath)
	container, err := sqlstore.New("sqlite3", dbPath, dbLog)
	if err != nil {
		log.Infof("Failed to connect to WhatsApp database: %v", err)
		panic(err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		log.Infof("Failed to get first device from WhatsApp database: %v", err)
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
	routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]any{
		"status":  "disconnected",
		"message": "WhatsApp initialized but not connected",
	})
	routes.GlobalWebSocketHandler.Broadcast("whatsapp_qr", map[string]any{
		"qr_code_base64": "",
		"logged_in":      false,
	})

	// Check if we already have credentials
	if client.Store.ID != nil {
		log.Info("WhatsApp credentials found")
		// Inform the frontend that credentials exist
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "WhatsApp credentials found, click refresh to connect",
		})
	} else {
		// No credentials, inform the frontend
		log.Info("No WhatsApp credentials found")
		routes.GlobalWebSocketHandler.Broadcast("whatsapp_status", map[string]any{
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

	log.Infoln("Shutting down gracefully, press Ctrl+C again to force")

	zkSocket.StopCapture()
	log.Infoln("ZK device capture stopped")

	zkSocket.Disconnect()
	log.Infoln("ZK device disconnected")

	whatsapp.Disconnect()
	log.Infoln("WhatsApp client disconnected")

	fileDeleted := getWhatsappPath()
	deleteErr := os.Remove(fileDeleted)
	if deleteErr != nil {
		log.Infof("Error deleting WhatsApp store file: %v", deleteErr)
	}
	log.Infof("WhatsApp store file deleted: %s", fileDeleted)

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown with error: %v", err)
	}

	log.Info("API server shutdown")
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

	eventLogger := logStd.New(logFile, "", logStd.LstdFlags)
	apiServer := server.NewServer()
	zkSocket := setupZKDevice(eventLogger)
	whatsapp := setupWhatsapp()

	done := make(chan bool, 1)
	go gracefulShutdown(apiServer, zkSocket, whatsapp, done)

	log.Infoln("Starting API server")
	if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start API server: %v", err)
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Infoln("Shutdown complete")
}
