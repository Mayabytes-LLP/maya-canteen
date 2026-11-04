package server

import (
	"context"
	"maya-canteen/internal/gozk"
	"maya-canteen/internal/handlers"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func GracefulShutdown(apiServer *http.Server, zkSocket *gozk.ZK, getWhatsApp func() handlers.WhatsAppClient, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Infoln("Shutting down gracefully, press Ctrl+C again to force")
	log.Infoln("Stopping ZK device capture...")
	zkSocket.StopCapture()
	log.Infoln("ZK device capture stopped")
	log.Infoln("Disconnecting ZK device...")
	zkSocket.Disconnect()
	log.Infoln("ZK device disconnected")
	// Only disconnect: logging out or deleting the store would unlink the device
	// and force a QR re-scan after every restart.
	if whatsapp := getWhatsApp(); whatsapp != nil {
		log.Infoln("Disconnecting WhatsApp (session is kept for the next start)...")
		whatsapp.Disconnect()
	} else {
		log.Infoln("WhatsApp client is nil, skipping disconnect")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Infoln("Shutting down API server...")
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown with error: %v", err)
	} else {
		log.Infoln("API server shutdown gracefully")
	}
	log.Infoln("Shutdown sequence completed")
	close(done)
}
