package handlers

import (
	"context"
	"fmt"
	"log"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type Client = whatsmeow.Client

// WhatsAppHandler manages the WhatsApp integration with our application
type WhatsAppHandler struct {
	common.BaseHandler
	whatsappClient *whatsmeow.Client
}

// NewWhatsAppHandler creates a new WhatsApp handler with the given database service and client
func NewWhatsAppHandler(db database.Service, client *whatsmeow.Client) *WhatsAppHandler {
	return &WhatsAppHandler{
		BaseHandler:    common.NewBaseHandler(db),
		whatsappClient: client,
	}
}

// SendWhatsAppMessage sends a message to a user's WhatsApp number
func (h *WhatsAppHandler) SendWhatsAppMessage(phoneNumber, message string) error {
	// Check if WhatsApp client is connected
	if h.whatsappClient == nil || !h.whatsappClient.IsConnected() {
		return fmt.Errorf("WhatsApp client is not connected")
	}

	// Parse phone number as JID (WhatsApp ID)
	recipient, err := types.ParseJID(phoneNumber + "@s.whatsapp.net")
	if err != nil {
		return fmt.Errorf("invalid phone number format: %v", err)
	}

	// Create message with current timestamp
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
		},
	}

	// Send message with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = h.whatsappClient.SendMessage(ctx, recipient, msg)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %v", err)
	}

	return nil
}

// NotifyUserBalance sends a balance notification to a specific user
func (h *WhatsAppHandler) NotifyUserBalance(w http.ResponseWriter, r *http.Request) {

	// Extract employee ID from URL params
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

	// Format the message
	message := fmt.Sprintf(
		"📊 *Balance Update* 📊\n\nDear %s,\n\nYour current canteen balance is: *PKR %.2f*\n\nThis is an automated message from Maya Canteen Management System.\nThank you for using our services!",
		user.Name,
		float64(userBalance.Balance), // Assuming 'Amount' is the numeric field in models.UserBalance
	)

	// Send the message via WhatsApp
	err = h.SendWhatsAppMessage(user.Phone, message)
	if err != nil {
		log.Printf("Error sending WhatsApp balance notification to %s: %v", user.Phone, err)
		common.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send WhatsApp message: %v", err))
		return
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Balance notification sent to %s", user.Name),
	})
}

// NotifyAllUsersBalances sends balance notifications to all users
func (h *WhatsAppHandler) NotifyAllUsersBalances(w http.ResponseWriter, r *http.Request) {
	// Get all users with balances
	userBalances, err := h.DB.GetUsersBalances()
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Failed to get users' balances")
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

		// Format message
		message := fmt.Sprintf(
			"📊 *Balance Update* 📊\n\nDear %s,\n\nYour current canteen balance is: *PKR %.2f*\n\nThis is an automated message from Maya Canteen Management System.\nThank you for using our services!",
			balance.UserName,
			balance.Balance,
		)

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

	common.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Sent %d notifications, %d failed", successCount, failCount),
		"details": map[string]interface{}{
			"success_count": successCount,
			"fail_count":    failCount,
			"failed_users":  failedUsers,
		},
	})
}
