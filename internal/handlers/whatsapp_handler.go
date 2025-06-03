package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"maya-canteen/internal/models"
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

	results, err := client.IsOnWhatsApp([]string{phoneNumber})
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

// formatBalanceMessage formats the balance notification message with user details
func (h *WhatsAppHandler) formatBalanceMessage(template string, name string, balance float64) string {
	message := template
	message = strings.ReplaceAll(message, "{name}", name)
	message = strings.ReplaceAll(message, "{balance}", fmt.Sprintf("%.2f", balance))
	return message
}

// formatTransactionHistory formats transaction history in both CSV and text format
func (h *WhatsAppHandler) formatTransactionHistory(transactions []models.Transaction) (string, string) {
	var csvContent strings.Builder
	var textContent strings.Builder

	// CSV Header
	csvContent.WriteString("Date,Type,Amount,Description\n")

	// Text Header
	textContent.WriteString("Transaction History:\n")
	textContent.WriteString("Date | Type | Amount | Description\n")
	textContent.WriteString("--------------------------------\n")

	for _, t := range transactions {
		// CSV Format
		desc := strings.ReplaceAll(t.Description, "\"", "\"\"") // escape quotes in description
		csvContent.WriteString(fmt.Sprintf("%s,%s,%.2f,\"%s\"\n",
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.TransactionType,
			t.Amount,
			desc))

		// Text Format
		textContent.WriteString(fmt.Sprintf("%s | %s | %.2f | %s\n",
			t.CreatedAt.Format("2006-01-02"),
			t.TransactionType,
			t.Amount,
			t.Description))
	}

	return csvContent.String(), textContent.String()
}

// sendBalanceNotification sends a balance notification to a single user
func (h *WhatsAppHandler) sendBalanceNotification(user models.User, userBalance models.UserBalance, messageTemplate string, startDate, endDate time.Time) error {
	// Format and send balance message
	message := h.formatBalanceMessage(messageTemplate, user.Name, float64(userBalance.Balance))

	fmt.Printf("Sending balance notification to %s: %s\n", user.Phone, message)

	//
	// Get transactions for the period
	// transactions, err := h.DB.GetTransactionsByDateRange(startDate, endDate)
	// if err != nil {
	// 	return fmt.Errorf("failed to get transactions: %v", err)
	// }
	//
	// // Filter transactions for this user
	// var userTransactions []models.Transaction
	// for _, t := range transactions {
	// 	if t.UserID == user.ID {
	// 		userTransactions = append(userTransactions, t)
	// 	}
	// }
	//
	// Format transaction history
	// csvContent, textContent := h.formatTransactionHistory(userTransactions)

	// Send the balance message
	if err := h.SendWhatsAppMessage(user.Phone, message); err != nil {
		return fmt.Errorf("failed to send balance message: %v", err)
	}

	// // Send transaction history in text format
	// if err := h.SendWhatsAppMessage(user.Phone, textContent); err != nil {
	// 	return fmt.Errorf("failed to send transaction text: %v", err)
	// }
	//
	// Send transaction history in CSV format
	// fileName := fmt.Sprintf("transactions_%s_%d.csv", startDate.Format("January"), startDate.Year())
	// if err := h.SendDocumentMessage(user.Phone, fileName, []byte(csvContent), "text/csv"); err != nil {
	// 	return fmt.Errorf("failed to send transaction CSV: %v", err)
	// }
	//
	return nil
}

func (h *WhatsAppHandler) NotifyUserBalance(w http.ResponseWriter, r *http.Request) {
	client := h.GetWhatsAppClient()
	if client == nil || !client.IsLoggedIn() || !client.IsConnected() {
		common.RespondWithError(w, http.StatusInternalServerError, "WhatsApp client is not available")
		return
	}

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

	if user.Phone == "" {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("User with employee ID %s does not have a phone number", strconv.FormatInt(employeeID, 10)))
		return
	}

	userBalance, err := h.DB.GetUserBalanceByUserID(user.ID)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Failed to get user balance")
		return
	}

	// Parse request body
	type reqBody struct {
		MessageTemplate string `json:"message_template"`
		Month           string `json:"month"`
		Year            int    `json:"year"`
	}
	var body reqBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	messageTemplate := body.MessageTemplate
	if messageTemplate == "" {
		messageTemplate = "**Balance Update** \n\nDear {name},\nYour current canteen balance is: *PKR {balance}*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) half month of Canteen bill\n\nThis is an automated message from Maya Canteen Management System."
	}

	// Set up date range
	month := body.Month
	year := body.Year
	if month == "" {
		month = time.Now().Format("January")
	}
	if year == 0 {
		year = time.Now().Year()
	}

	startDate, err := time.Parse("January 2006", fmt.Sprintf("%s %d", month, year))
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Invalid month format")
		return
	}
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Send notification
	if err := h.sendBalanceNotification(*user, userBalance, messageTemplate, startDate, endDate); err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send notification: %v", err))
		return
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Balance notification and transactions sent to %s", user.Name),
	})
}

func (h *WhatsAppHandler) NotifyAllUsersBalances(w http.ResponseWriter, r *http.Request) {
	client := h.GetWhatsAppClient()
	if client == nil || !client.IsLoggedIn() || !client.IsConnected() {
		common.RespondWithError(w, http.StatusInternalServerError, "WhatsApp client is not available")
		return
	}

	// Parse request body
	type reqBody struct {
		MessageTemplate string `json:"message_template"`
		Month           string `json:"month"`
		Year            int    `json:"year"`
	}
	var body reqBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	messageTemplate := body.MessageTemplate
	if messageTemplate == "" {
		messageTemplate = "**Balance Update** \n\nDear {name},\nYour current canteen balance is: *PKR {balance}*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) half month of Canteen bill\n\nThis is an automated message from Maya Canteen Management System."
	}

	// Set up date range
	month := body.Month
	year := body.Year
	if month == "" {
		month = time.Now().Format("January")
	}
	if year == 0 {
		year = time.Now().Year()
	}

	startDate, err := time.Parse("January 2006", fmt.Sprintf("%s %d", month, year))
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Invalid month format")
		return
	}
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Get all users with balances
	userBalances, err := h.DB.GetUsersBalances()
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Failed to get user's balance")
		return
	}

	successCount := 0
	failCount := 0
	failedUsers := []string{}

	// Send notification to each user
	for _, balance := range userBalances {
		if !balance.UserActive || balance.Phone == "" {
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (inactive or no phone number)", balance.UserName))
			continue
		}

		user := models.User{
			ID:    balance.UserID,
			Name:  balance.UserName,
			Phone: balance.Phone,
		}

		userBalance := models.UserBalance{
			Balance: balance.Balance,
		}

		if err := h.sendBalanceNotification(user, userBalance, messageTemplate, startDate, endDate); err != nil {
			log.Printf("Failed to send WhatsApp notification to %s (%s): %v", balance.UserName, balance.Phone, err)
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (%v)", balance.UserName, err))
		} else {
			successCount++
		}

		// Add a small delay between messages to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
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
}

// SendDocumentMessage sends a document message to a user's WhatsApp number
func (h *WhatsAppHandler) SendDocumentMessage(phoneNumber string, fileName string, fileData []byte, mimeType string) error {
	// Check if WhatsApp client is connected
	client := h.GetWhatsAppClient()
	if client == nil {
		return fmt.Errorf("WhatsApp client is not initialized")
	}
	if !client.IsLoggedIn() {
		log.Warn("WhatsApp client is not connected. Cannot send message.")
		return fmt.Errorf("WhatsApp client is not connected")
	}

	results, err := client.IsOnWhatsApp([]string{phoneNumber})
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

	log.Infof("Sending WhatsApp document to %s: %s", recipient, fileName)

	// Upload the file to WhatsApp servers
	uploaded, err := client.Upload(context.Background(), fileData, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload document: %v", err)
	}

	// Create document message
	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			FileName:      proto.String(fileName),
			Mimetype:      proto.String(mimeType),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(fileData))),
			URL:           proto.String(uploaded.URL),
		},
	}

	// Send message with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.SendMessage(ctx, recipient, msg)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp document: %v", err)
	}

	return nil
}
