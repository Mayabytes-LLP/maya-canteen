package handlers

import (
	"testing"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

type fakeClientMgr struct {
	connected, paired bool
	disconnects       int
}

func (f *fakeClientMgr) Disconnect()       { f.disconnects++; f.connected = false }
func (f *fakeClientMgr) IsConnected() bool { return f.connected }
func (f *fakeClientMgr) IsPaired() bool    { return f.paired }

func TestEventHandlerLoggedOutRebuildsClient(t *testing.T) {
	mgr := &fakeClientMgr{connected: true}
	rebuilt := make(chan struct{}, 1)
	var statuses []string
	broadcast := func(ev string, data map[string]any) {
		if ev == "whatsapp_status" {
			statuses = append(statuses, data["status"].(string))
		}
	}
	eventHandler(&events.LoggedOut{}, broadcast, mgr, func() { rebuilt <- struct{}{} })

	select {
	case <-rebuilt:
	case <-time.After(time.Second):
		t.Fatal("LoggedOut must trigger the client rebuild callback")
	}
	if mgr.disconnects != 1 || len(statuses) != 1 || statuses[0] != "disconnected" {
		t.Fatalf("unexpected state: disconnects=%d statuses=%v", mgr.disconnects, statuses)
	}
}

func TestEventHandlerDisconnectedStatus(t *testing.T) {
	for _, paired := range []bool{true, false} {
		var got string
		eventHandler(&events.Disconnected{}, func(ev string, d map[string]any) { got = d["status"].(string) }, &fakeClientMgr{paired: paired}, nil)
		want := "disconnected"
		if paired {
			want = "connecting" // transient: whatsmeow auto-reconnects paired clients
		}
		if got != want {
			t.Errorf("paired=%v: status %q, want %q", paired, got, want)
		}
	}
}
