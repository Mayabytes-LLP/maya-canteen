package handlers

import (
	"strings"
	"testing"
	"time"

	"maya-canteen/internal/models"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"+923452324442", "+923452324442", false},
		{"923452324442", "+923452324442", false},
		{"0345 2324442", "+923452324442", false},
		{"0092-345-2324442", "+923452324442", false},
		{"+92 (345) 232-4442", "+923452324442", false},
		{"", "", true},
		{"12345", "", true},
		{"+1234567890123456", "", true},
	}
	for _, tt := range tests {
		got, err := normalizePhone(tt.in)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("normalizePhone(%q) = %q, %v; want %q (err=%v)", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestBulkSkipsDuplicatePhones(t *testing.T) {
	h := &WhatsAppHandler{GetWhatsAppClient: func() *Client { return nil }}
	users := []models.User{
		{ID: 1, Name: "A", Phone: "+923422949447"},
		{ID: 2, Name: "B", Phone: "03422949447"},
	}
	balances := make([]models.UserBalance, 2)
	_, fail, failed := sendBalanceNotifications(h, users, balances, "x", time.Now(), time.Now(), false, false, nil)
	if fail != 2 || len(failed) != 2 || !strings.Contains(failed[1], "duplicate phone number") {
		t.Fatalf("expected B to be skipped as duplicate, got fail=%d failed=%v", fail, failed)
	}
}
