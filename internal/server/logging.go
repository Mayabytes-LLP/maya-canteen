package server

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// SetupLogFile creates or opens a log file for writing logs.
func SetupLogFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Infof("Failed to open log file: %v", err)
		return nil, err
	}
	return file, nil
}
