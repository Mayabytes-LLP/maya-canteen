package server

import (
	"encoding/json"
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	var got struct {
		Status string            `json:"status"`
		Data   map[string]string `json:"data"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("response is not valid JSON: %v (%s)", err, body)
	}
	if got.Status != "success" || got.Data["message"] == "" {
		t.Errorf("unexpected response body: %s", body)
	}
}
