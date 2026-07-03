package main

import (
	"context"
	logStd "log"
	"maya-canteen/internal/handlers"
	"maya-canteen/internal/server"
	"maya-canteen/internal/server/routes"
	"net"
	"os"
	"strconv"

	"go.mau.fi/whatsmeow"
)

func mustListen(port int) net.Listener {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic("Failed to listen on port " + strconv.Itoa(port) + ": " + err.Error())
	}
	return ln
}

func main() {
	logFile, err := server.SetupLogFile("zk_events.log")
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}
	defer logFile.Close()

	os.Setenv("BLUEPRINT_DB_URL", server.SetupDBPath())

	eventLogger := logStd.New(logFile, "", logStd.LstdFlags)

	broadcastFunc := func(event string, data map[string]any) {
		routes.GlobalWebSocketHandler.Broadcast(event, data)
	}

	zkSocket := handlers.SetupZKDevice(eventLogger, broadcastFunc)

	listener := mustListen(server.DefaultPort())
	apiServer := server.NewServer(nil)

	whatsapp, whatsappDbPath := handlers.SetupWhatsapp(
		broadcastFunc,
		func(getter func(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error)) {
			routes.GlobalWebSocketHandler.RegisterQRChannelGetter(getter)
		},
	)

	whatsappInterface := handlers.NewWhatsAppClientInterface(whatsapp)

	if routes.GlobalWebSocketHandler != nil {
		routes.GlobalWebSocketHandler.UpdateWhatsAppClient(whatsappInterface)
	}

	done := make(chan bool, 1)
	go server.GracefulShutdown(apiServer, zkSocket, whatsappInterface, whatsappDbPath, done)

	log := logFile
	_ = log

	if err := apiServer.Serve(listener); err != nil {
		panic("Failed to start API server: " + err.Error())
	}

	<-done
}