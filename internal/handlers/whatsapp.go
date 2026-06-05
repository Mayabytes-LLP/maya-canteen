package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsAppClientManager interface {
	Disconnect()
	IsConnected() bool
}

type whatsAppClientManager struct {
	client *whatsmeow.Client
}

func (m *whatsAppClientManager) Disconnect() {
	m.client.Disconnect()
}

func (m *whatsAppClientManager) IsConnected() bool {
	return m.client.IsConnected()
}

// EventHandler processes WhatsApp connection-related events.
func EventHandler(evt any, broadcastFunc func(event string, data map[string]any), clientMgr WhatsAppClientManager) {
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
		log.Info("Logged out from WhatsApp - disconnecting and clearing session")
		if clientMgr.IsConnected() {
			clientMgr.Disconnect()
		}
		broadcastFunc("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "WhatsApp logged out. Please reconnect by refreshing the QR code.",
		})
		broadcastFunc("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      false,
		})
	case *events.StreamReplaced:
		log.Info("WhatsApp connected from another location - disconnecting")
		if clientMgr.IsConnected() {
			clientMgr.Disconnect()
		}
		broadcastFunc("whatsapp_status", map[string]any{
			"status":  "disconnected",
			"message": "WhatsApp connected from another location. Session replaced.",
		})
		broadcastFunc("whatsapp_qr", map[string]any{
			"qr_code_base64": "",
			"logged_in":      false,
		})
	case *events.PairSuccess:
		log.Infof("Pair success for device: %s", v.ID)
	default:
		log.Debugf("Unhandled WhatsApp event: %T", v)
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

	store.SetOSInfo("Maya Canteen", [3]uint32{2, 3000, 1040847988})
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()

	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.QRClientType = whatsmeow.PairClientChrome
	clientMgr := &whatsAppClientManager{client: client}
	client.AddEventHandler(func(evt any) { EventHandler(evt, broadcastFunc, clientMgr) })
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
