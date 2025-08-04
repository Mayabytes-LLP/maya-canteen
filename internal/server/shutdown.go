package server

import (
	"context"
	"maya-canteen/internal/gozk"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// GracefulShutdown handles cleanup and shutdown of all services.
func GracefulShutdown(apiServer *http.Server, zkSocket *gozk.ZK, whatsapp *whatsmeow.Client, whatsappDbPath string, done chan bool) {
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
	if whatsapp.IsConnected() {
		log.Infoln("Logging out from WhatsApp...")
		ctx := context.Background()
		err := whatsapp.Logout(ctx)
		if err != nil {
			log.Errorf("Error logging out from WhatsApp: %v", err)
		} else {
			log.Infoln("WhatsApp logout successful")
		}
	}
	log.Infoln("Disconnecting WhatsApp client...")
	whatsapp.Disconnect()
	log.Infoln("WhatsApp client disconnected")
	log.Infof("Attempting to delete WhatsApp store file: %s", whatsappDbPath)
	if _, err := os.Stat(whatsappDbPath); err == nil {
		deleteErr := os.Remove(whatsappDbPath)
		if deleteErr != nil {
			log.Errorf("Error deleting WhatsApp store file: %v", deleteErr)
		} else {
			log.Infof("WhatsApp store file deleted successfully: %s", whatsappDbPath)
		}
	} else {
		log.Infof("WhatsApp store file not found: %s", whatsappDbPath)
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
