package handlers

import (
	logStd "log"
	"maya-canteen/internal/gozk"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// SetupZKDevice initializes and manages the ZK device connection and event capture.
// Accepts a broadcastFunc to decouple from routes and avoid import cycles.
func SetupZKDevice(eventLogger *logStd.Logger, broadcastFunc func(event string, data map[string]any)) *gozk.ZK {
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
				broadcastFunc("device_status", map[string]any{
					"status": "connected",
				})

			go func() {
				for {
					time.Sleep(3 * time.Second)
						broadcastFunc("device_status", map[string]any{
							"status": "connected",
						})
				}
			}()
			for event := range c {
				log.Printf("Event: %v", event)
				if event.UserID != "" {
					log.Infof("[WebSocket] Broadcasting attendance event - UserID: %v, Time: %v", event.UserID, event.AttendedAt)
					eventLogger.Printf("Event: %v", event)
						broadcastFunc("attendance_event", map[string]any{
							"user_id":   event.UserID,
							"timestamp": event.AttendedAt.Format(time.RFC3339),
						})
				}
			}

			log.Info("ZK Device disconnected. Retrying in 3 seconds...")
				broadcastFunc("device_status", map[string]any{
					"status": "disconnected",
				})
			time.Sleep(3 * time.Second)
		}
	}()

	return zkSocket
}
