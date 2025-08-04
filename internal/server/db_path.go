package server

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// SetupDBPath returns the database path, configurable via env var.
func SetupDBPath() string {
	dbPath := os.Getenv("BLUEPRINT_DB_URL")
	if dbPath == "" {
		executablePath, err := os.Executable()
		if err != nil {
			log.Infof("Error getting executable path: %v, using default", err)
		}
		return filepath.Join(filepath.Dir(executablePath), "db", "canteen.db")
	}
	return dbPath
}
