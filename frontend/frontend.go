package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

//go:embed dist
var DistFS embed.FS

// Add this to debug embedding issues
func debugEmbeddedFS(embeddedFS embed.FS) {
	entries, _ := embeddedFS.ReadDir(".")
	for _, entry := range entries {
		log.Infof("Embedded entry: %s\n", entry.Name())
	}
}

// Get the subtree of the embedded files with `dist` directory as a root.
func BuildHTTPFS() http.FileSystem {
	build, err := fs.Sub(DistFS, "dist")
	if err != nil {
		log.Fatal(err)
	}
	return http.FS(build)
}

// ReadIndexHTML reads the index.html file from the frontend filesystem
func ReadIndexHTML() ([]byte, error) {
	if _, err := os.Stat("./dist"); err == nil {
		// Development: read from disk
		log.Infof("Reading index.html from disk")
		return os.ReadFile("./dist/index.html")
	}

	// Production: read from embedded filesystem
	fsys, err := fs.Sub(DistFS, "dist")
	if err != nil {
		log.Errorf("Error reading embedded filesystem: %v", err)
		return nil, err
	}

	return fs.ReadFile(fsys, "index.html")
}

// isAPIRoute returns true if the path is an API endpoint
func isAPIRoute(path string) bool {
	// Add your API route prefixes here
	apiPrefixes := []string{
		"/api/",
		"/ws",
	}

	for _, prefix := range apiPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// ServeStaticFiles configures routes for serving the frontend
// This function now handles http.Handler (compatible with gorilla/mux) instead of http.ServeMux
func ServeStaticFiles(handler http.Handler) http.Handler {
	// Debug embedded filesystem
	debugEmbeddedFS(DistFS)

	// Create a new handler that will first check if the request should be handled
	// by our API routes, and if not, serve static files
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an API request
		if isAPIRoute(r.URL.Path) {
			// Pass API requests to the original handler
			handler.ServeHTTP(w, r)
			return
		}

		// For static file requests like /assets/...
		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/assets/" {
			// Serve static files
			if _, err := os.Stat("./dist"); err == nil {
				// Development: serve from filesystem
				http.StripPrefix("/", http.FileServer(http.Dir("./dist"))).ServeHTTP(w, r)
				return
			}

			// Production: serve from embedded filesystem
			fsys, err := fs.Sub(DistFS, "dist")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
			return
		}

		// For other routes, serve index.html (SPA routing)
		var indexHTML []byte

		if _, err := os.Stat("./dist"); err == nil {
			// Development: read from disk
			indexHTML, err = os.ReadFile("./dist/index.html")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
			// Production: read from embedded filesystem
			fsys, err := fs.Sub(DistFS, "dist")
			if err == nil {
				indexHTML, err = fs.ReadFile(fsys, "index.html")
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(indexHTML)
	})
}
