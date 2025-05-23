package server

import (
	"io"
	"maya-canteen/internal/server/routes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	// Create a system handler directly
	systemHandlers := routes.NewSystemHandlers(nil)

	// Test the HelloWorldHandler
	server := httptest.NewServer(http.HandlerFunc(systemHandlers.HelloWorldHandler))
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()
	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
	expected := "{\"success\":true,\"data\":{\"message\":\"Hello World\"}}\n"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}
