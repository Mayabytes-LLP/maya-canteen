package main

import (
	"context"
	logStd "log"
	"maya-canteen/internal/handlers"
	"maya-canteen/internal/server"
	"maya-canteen/internal/server/routes"
	"os"

	"go.mau.fi/whatsmeow"
)


func main() {
	logFile, err := server.SetupLogFile("zk_events.log")
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}
	defer logFile.Close()

	// Set the database path
	os.Setenv("BLUEPRINT_DB_URL", server.SetupDBPath())

	eventLogger := logStd.New(logFile, "", logStd.LstdFlags)

	// Use the broadcast function from the global WebSocket handler
	broadcastFunc := func(event string, data map[string]any) {
		routes.GlobalWebSocketHandler.Broadcast(event, data)
	}

	zkSocket := handlers.SetupZKDevice(eventLogger, broadcastFunc)

	// Initialize the server first with a nil WhatsApp client
	apiServer := server.NewServer(nil)

	// WhatsApp setup: pass broadcast and QR channel registration functions
	whatsapp, whatsappDbPath := handlers.SetupWhatsapp(
		broadcastFunc,
		func(getter func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error)) {
			routes.GlobalWebSocketHandler.RegisterQRChannelGetter(getter)
		},
	)

	if routes.GlobalWebSocketHandler != nil {
		routes.GlobalWebSocketHandler.UpdateWhatsAppClient(whatsapp)
	}

	done := make(chan bool, 1)
	go server.GracefulShutdown(apiServer, zkSocket, whatsapp, whatsappDbPath, done)

	log := logFile // for logging in main
	_ = log

	if err := apiServer.ListenAndServe(); err != nil {
		panic("Failed to start API server: " + err.Error())
	}

	<-done
}
