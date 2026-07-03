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
	"go.mau.fi/whatsmeow/types"
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
	default:
		log.Debugf("Unhandled WhatsApp event: %T", v)
	}
}

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
	client.EnableAutoReconnect = true
	client.InitialAutoReconnect = true
	client.AutoReconnectHook = func(err error) bool {
		log.Warnf("WhatsApp auto-reconnect failed (%d attempts so far): %v", client.AutoReconnectErrors, err)
		return client.AutoReconnectErrors < 5
	}
	client.SynchronousAck = true
	client.UseRetryMessageStore = true
	client.PrePairCallback = func(jid types.JID, platform, businessName string) bool {
		log.Infof("WhatsApp pairing with %s (platform: %s, business: %s)", jid, platform, businessName)
		return true
	}
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

func NewWhatsAppClientInterface(client *whatsmeow.Client) WhatsAppClient {
	return &whatsmeowClientWrapper{client: client}
}

type whatsmeowClientWrapper struct {
	client *whatsmeow.Client
}

func (w *whatsmeowClientWrapper) Logout(ctx context.Context) error {
	return w.client.Logout(ctx)
}

func (w *whatsmeowClientWrapper) Connect() error {
	return w.client.Connect()
}

func (w *whatsmeowClientWrapper) IsConnected() bool {
	return w.client.IsConnected()
}

func (w *whatsmeowClientWrapper) IsLoggedIn() bool {
	return w.client.IsLoggedIn()
}

func (w *whatsmeowClientWrapper) Disconnect() {
	w.client.Disconnect()
}

func (w *whatsmeowClientWrapper) GetStoreID() *types.JID {
	if w.client.Store == nil {
		return nil
	}
	return w.client.Store.ID
}

func (w *whatsmeowClientWrapper) GetClient() *whatsmeow.Client {
	return w.client
}

func (w *whatsmeowClientWrapper) PairPhone(ctx context.Context, phone string) (string, error) {
	return w.client.PairPhone(ctx, phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
}