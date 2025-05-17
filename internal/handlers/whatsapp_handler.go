package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

type Client = whatsmeow.Client

// WhatsAppHandler manages the WhatsApp integration with our application
type WhatsAppHandler struct {
	common.BaseHandler
	GetWhatsAppClient func() *Client // Function to get the current WhatsApp client
}

// NewWhatsAppHandler creates a new WhatsApp handler with the given database service and client getter
func NewWhatsAppHandler(db database.Service, getClient func() *whatsmeow.Client) *WhatsAppHandler {
	return &WhatsAppHandler{
		BaseHandler:       common.NewBaseHandler(db),
		GetWhatsAppClient: getClient,
	}
}

// SendWhatsAppMessage sends a message to a user's WhatsApp number
func (h *WhatsAppHandler) SendWhatsAppMessage(phoneNumber, message string) error {
	// Check if WhatsApp client is connected
	client := h.GetWhatsAppClient()
	if client == nil {
		return fmt.Errorf("WhatsApp client is not initialized")
	}
	if !client.IsLoggedIn() {
		log.Warn("WhatsApp client is not connected. Cannot send message.")
		return fmt.Errorf("WhatsApp client is not connected")
	}

	results, err := h.GetWhatsAppClient().IsOnWhatsApp([]string{phoneNumber})
	if err != nil {
		return fmt.Errorf("failed to check WhatsApp status: %v", err)
	}
	if len(results) == 0 || !results[0].IsIn {
		log.Warnf("Phone number %s is not on WhatsApp", phoneNumber)
		return fmt.Errorf("phone number is not on WhatsApp")
	}

	if len(results) > 1 {
		log.Warnf("Multiple results for phone number %s: %v", phoneNumber, results)
		return fmt.Errorf("multiple results for phone number")
	}
	recipient := results[0].JID

	log.Infof("Sending WhatsApp message to %s: %s", recipient, message)

	// Create message with current timestamp
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
		},
	}

	// Send message with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.SendMessage(ctx, recipient, msg)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %v", err)
	}

	return nil
}

// NotifyUserBalance sends a balance notification to a specific user
func (h *WhatsAppHandler) NotifyUserBalance(w http.ResponseWriter, r *http.Request) {
	if h.GetWhatsAppClient() == nil {
		common.RespondWithError(w, http.StatusInternalServerError, "WhatsApp client is not initialized")
		return
	}
	if !h.GetWhatsAppClient().IsLoggedIn() {
		common.RespondWithError(w, http.StatusInternalServerError, "WhatsApp client is not connected")
		return
	}

	connected := h.GetWhatsAppClient().IsConnected()
	if !connected {
		common.RespondWithError(w, http.StatusInternalServerError, "WhatsApp client is not connected")
		return
	}
	log.Info("WhatsApp client disconnected")

	vars := mux.Vars(r)
	employeeID, err := h.ParseID(vars, "id")
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Employee ID is required")
		return
	}

	// Get user and balance information
	user, err := h.DB.GetUser(employeeID)
	if err != nil {
		common.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("User with employee ID %s not found", strconv.FormatInt(employeeID, 10)))
		return
	}

	// Validate that user has a phone number
	if user.Phone == "" {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("User with employee ID %s does not have a phone number", strconv.FormatInt(employeeID, 10)))
		return
	}

	// Get balance for the user
	userBalance, err := h.DB.GetUserBalanceByUserID(user.ID)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Failed to get user balance")
		return
	}

	// Parse message_template from POST body
	type reqBody struct {
		MessageTemplate string `json:"message_template"`
	}
	var body reqBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	messageTemplate := body.MessageTemplate
	if messageTemplate == "" {
		messageTemplate = "**Balance Update** \n\nDear {name},\nYour current canteen balance is: *PKR {balance}*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) half month of Canteen bill\n\nThis is an automated message from Maya Canteen Management System."
	}

	// Replace placeholders
	message := messageTemplate
	message = strings.ReplaceAll(message, "{name}", user.Name)
	message = strings.ReplaceAll(message, "{balance}", fmt.Sprintf("%.2f", float64(userBalance.Balance)))

	log.Infof("Sending WhatsApp message to %s: %s", user.Phone, message)

	// Send the message via WhatsApp
	err = h.SendWhatsAppMessage(user.Phone, message)
	if err != nil {
		log.Printf("Error sending WhatsApp balance notification to %s: %v", user.Phone, err)
		common.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send WhatsApp message: %v", err))
		return
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Balance notification sent to %s", user.Name),
	})
}

// NotifyAllUsersBalances sends balance notifications to all users
func (h *WhatsAppHandler) NotifyAllUsersBalances(w http.ResponseWriter, r *http.Request) {
	// Parse message_template from POST body
	type reqBody struct {
		MessageTemplate string `json:"message_template"`
	}
	var body reqBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	messageTemplate := body.MessageTemplate
	if messageTemplate == "" {
		messageTemplate = "**Balance Update** \n\nDear {name},\nYour current canteen balance is: *PKR {balance}*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) half month of Canteen bill\n\nThis is an automated message from Maya Canteen Management System."
	}

	// Get all users with balances
	userBalances, err := h.DB.GetUsersBalances()
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Failed to get user's balance")
		return
	}

	successCount := 0
	failCount := 0
	failedUsers := []string{}

	// Send notification to each user with a phone number
	for _, balance := range userBalances {
		// Skip users without phone numbers
		if balance.Phone == "" {
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (no phone number)", balance.UserName))
			continue
		}

		// Replace placeholders
		message := messageTemplate
		message = strings.ReplaceAll(message, "{name}", balance.UserName)
		message = strings.ReplaceAll(message, "{balance}", fmt.Sprintf("%.2f", balance.Balance))

		// Send message
		err = h.SendWhatsAppMessage(balance.Phone, message)
		if err != nil {
			log.Printf("Failed to send WhatsApp notification to %s (%s): %v", balance.UserName, balance.Phone, err)
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (%v)", balance.UserName, err))
		} else {
			successCount++
		}
		// Add a small delay between messages to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Sent %d notifications, %d failed", successCount, failCount),
		"details": map[string]any{
			"success_count": successCount,
			"fail_count":    failCount,
			"failed_users":  failedUsers,
		},
	})

	log.Infof("WhatsApp client pointer in WhatsAppHandler: %p", h.GetWhatsAppClient())

}
