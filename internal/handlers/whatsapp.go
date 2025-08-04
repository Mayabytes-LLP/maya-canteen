package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// EventHandler processes WhatsApp connection-related events.
func EventHandler(evt any, broadcastFunc func(event string, data map[string]any)) {
	switch v := evt.(type) {
	case *events.Connected:
		log.Info("Connected to WhatsApp")
		broadcastFunc("whatsapp_status", map[string]any{
			"status":  "connected",
			"message": "WhatsApp connected successfully",
		})
		broadcastFunc("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      true,
		})
	case *events.LoggedOut:
		log.Info("Logged out from WhatsApp")
		broadcastFunc("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "WhatsApp logged out",
		})
		broadcastFunc("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      false,
		})
	case *events.StreamReplaced:
		log.Info("WhatsApp connected from another location")
		broadcastFunc("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "WhatsApp connected from another location",
		})
	default:
		log.Infof("Unhandled event: %v", v)
	}
}

// GetWhatsappPath returns both the database URI for SQLite and the actual file path.
func GetWhatsappPath() (dbUri string, filePath string) {
	absPath, err := filepath.Abs("./whatsapp-store.db")
	if err != nil {
		log.Errorf("Error getting absolute path: %v, using default", err)
		return "file:whatsapp-store.db?_foreign_keys=on", "whatsapp-store.db"
	}
	filePath = absPath
	if os.PathSeparator == '\\' {
		dbUri = fmt.Sprintf("file:/%s?_foreign_keys=on", filepath.ToSlash(absPath))
	} else {
		dbUri = fmt.Sprintf("file:%s?_foreign_keys=on", absPath)
	}
	return dbUri, filePath
}

// SetupWhatsapp initializes the WhatsApp client and registers event handlers.
func SetupWhatsapp(broadcastFunc func(event string, data map[string]any), registerQRChannelGetter func(func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error))) (*whatsmeow.Client, string) {
	dbLog := waLog.Stdout("Database", "INFO", true)
	dbUri, filePath := GetWhatsappPath()
	log.Infof("Using WhatsApp database at: %s", filePath)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", dbUri, dbLog)
	if err != nil {
		log.Infof("Failed to connect to WhatsApp database: %v", err)
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		log.Infof("Failed to get first device from WhatsApp database: %v", err)
		panic(err)
	}
	clientLog := waLog.Stdout("whatapp client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(func(evt any) { EventHandler(evt, broadcastFunc) })
	registerQRChannelGetter(func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error) {
		return client.GetQRChannel(ctx)
	})
	broadcastFunc("whatsapp_status", map[string]any{
		"status":  "disconnected",
		"message": "WhatsApp initialized but not connected",
	})
	broadcastFunc("whatsapp_qr", map[string]any{
		"qr_code_base64": "",
		"logged_in":      false,
	})
	return client, filePath
}
