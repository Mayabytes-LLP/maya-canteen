package main

import (
	"context"
	"fmt"
	"log"
	"maya-canteen/internal/gozk"
	"maya-canteen/internal/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func setupLogFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func gracefulShutdown(apiServer *http.Server, zkSocket *gozk.ZK, done chan bool) {
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
	eventLogger := log.New(logFile, "", log.LstdFlags)

	apiServer := server.NewServer()

	zkSocket := gozk.NewZK("192.168.1.153", 4370, 0, gozk.DefaultTimezone)
	if err := zkSocket.Connect(); err != nil {
		log.Fatalf("Failed to connect to ZK device: %v", err)
	}

	props, err := zkSocket.GetProperties()
	if err != nil {
		log.Fatalf("Failed to get ZK device properties: %v", err)
	}

	fmt.Printf("ZK device Total Users: %v\n", props.TotalUsers)
	fmt.Printf("ZK device Total Fingers: %v\n", props.TotalFingers)
	fmt.Printf("ZK device Total Records: %v\n", props.TotalRecords)

	fmt.Println("***************************************")
	fmt.Println("***************************************")

	// users, err := zkSocket.GetZktecoUsers()
	// if err != nil {
	// 	log.Printf("Failed to get ZK device users: %v", err)
	// } else {
	//
	// 	usersFile, err := setupLogFile("zk_users.log")
	// 	if err != nil {
	// 		log.Fatalf("Failed to open log file: %v", err)
	// 	}
	// 	defer usersFile.Close()
	// 	usersLogger := log.New(usersFile, "", log.LstdFlags)
	//
	// 	for _, user := range users {
	// 		usersLogger.Printf("%v", user)
	// 	}
	//
	// 	fmt.Printf("Users: %d\n", len(users))
	//
	// }

	attendances, err := zkSocket.GetAttendances()
	if err != nil {
		log.Fatalf("Failed to get ZK device users: %v", err)
	}
	fmt.Printf("Attendances: %d\n", len(attendances))

	fmt.Println("***************************************")
	fmt.Println("***************************************")

	eventLogger.Println("Starting live capture")
	c, err := zkSocket.LiveCapture()
	if err != nil {
		log.Fatalf("Failed to start live capture: %v", err)
	}

	go func() {
		for event := range c {
			// check if event contains user data
			if event.UserID != 0 {
				// send the event to the web socket

			}
		}
	}()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(apiServer, zkSocket, done)

	// Start the API server
	log.Println("Starting API server")
	if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", apiServer.Addr, err)
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
