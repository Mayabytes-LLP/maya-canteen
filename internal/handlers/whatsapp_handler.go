package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"maya-canteen/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

const (
	defaultBalanceMessageTemplate = "**Balance Update** \n\nDear {name},\nYour current canteen balance is: *PKR {balance}*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) half month of Canteen bill\n\nThis is an automated message from Maya Canteen Management System."
	csvHeader                     = "Date,Type,Amount,Description\n"
	textTransactionHeader         = "Transaction History:\n"
	textTransactionHeaderLine     = "Date | Type | Amount | Description\n"
	textTransactionSeparator      = "--------------------------------\n"
	notificationDelay             = 300 * time.Millisecond
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

// getWhatsAppRecipient checks client status and validates the recipient's phone number.
func (h *WhatsAppHandler) getWhatsAppRecipient(phoneNumber string) (types.JID, error) {
	client := h.GetWhatsAppClient()
	if client == nil {
		return types.JID{}, fmt.Errorf("WhatsApp client is not initialized")
	}
	if !client.IsLoggedIn() {
		log.Warn("WhatsApp client is not connected. Cannot send message.")
		return types.JID{}, fmt.Errorf("WhatsApp client is not connected")
	}

	results, err := client.IsOnWhatsApp([]string{phoneNumber})
	if err != nil {
		return types.JID{}, fmt.Errorf("failed to check WhatsApp status for %s: %v", phoneNumber, err)
	}
	if len(results) == 0 || !results[0].IsIn {
		log.Warnf("Phone number %s is not on WhatsApp", phoneNumber)
		return types.JID{}, fmt.Errorf("phone number is not on WhatsApp")
	}
	if len(results) > 1 {
		log.Warnf("Multiple JIDs found for phone number %s: %v", phoneNumber, results)
		return types.JID{}, fmt.Errorf("multiple JIDs found for phone number")
	}

	return results[0].JID, nil
}

// SendWhatsAppMessage sends a message to a user's WhatsApp number
func (h *WhatsAppHandler) SendWhatsAppMessage(phoneNumber, message string) error {
	recipient, err := h.getWhatsAppRecipient(phoneNumber)
	if err != nil {
		return err
	}

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

	client := h.GetWhatsAppClient()
	_, err = client.SendMessage(ctx, recipient, msg)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %v", err)
	}

	return nil
}

// formatBalanceMessage formats the balance notification message with user details
func (h *WhatsAppHandler) formatBalanceMessage(template string, name string, balance float64) string {
	var builder strings.Builder
	builder.WriteString(template)
	message := builder.String()
	message = strings.ReplaceAll(message, "{name}", name)
	message = strings.ReplaceAll(message, "{balance}", fmt.Sprintf("%.2f", balance))
	return message
}

// formatTransactionHistory formats transaction history in both CSV and text format
func (h *WhatsAppHandler) formatTransactionHistory(transactions []models.Transaction) (string, string) {
	var csvContent strings.Builder
	var textContent strings.Builder

	// CSV Header
	csvContent.WriteString(csvHeader)

	// Text Header
	textContent.WriteString(textTransactionHeader)
	textContent.WriteString(textTransactionHeaderLine)
	textContent.WriteString(textTransactionSeparator)

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
func (h *WhatsAppHandler) sendBalanceNotification(user models.User, userBalance models.UserBalance, messageTemplate string, startDate, endDate time.Time, includeTransactions bool) error {
	// Format balance message
	message := h.formatBalanceMessage(messageTemplate, user.Name, float64(userBalance.Balance))

	var combinedMessage string
	var csvContent string

	if includeTransactions {
		// Get transactions for the period
		transactions, err := h.DB.GetTransactionsByDateRange(startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to get transactions: %v", err)
		}

		// Filter transactions for this user
		var userTransactions []models.Transaction
		for _, t := range transactions {
			if t.UserID == user.ID {
				userTransactions = append(userTransactions, t)
			}
		}

		var textContent string
		if len(userTransactions) > 0 {
			csvContent, textContent = h.formatTransactionHistory(userTransactions)
		} else {
			csvContent = ""
			textContent = "No transactions found for this period."
		}

		// Combine balance message with transaction history (text)
		combinedMessage = message + "\n\n" + textContent
	} else {
		combinedMessage = message
	}

	// Always send the combined message (balance + transaction history if included)
	if err := h.SendWhatsAppMessage(user.Phone, combinedMessage); err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %v", err)
	}

	// If there are transactions and includeTransactions is true, send the CSV as a document (as a second message)
	if includeTransactions && csvContent != "" {
		fileName := fmt.Sprintf("transactions_%s_%d.csv", startDate.Format("January"), startDate.Year())
		if err := h.SendDocumentMessage(user.Phone, fileName, []byte(csvContent), "text/csv"); err != nil {
			return fmt.Errorf("failed to send transaction CSV: %v", err)
		}
	}

	return nil
}

// parseBalanceNotificationRequest parses the request body and returns message template, startDate, endDate, and includeTransactions
func parseBalanceNotificationRequest(r *http.Request) (string, time.Time, time.Time, bool, error) {
	const defaultTemplate = defaultBalanceMessageTemplate
	type reqBody struct {
		MessageTemplate     string `json:"message_template"`
		Month               string `json:"month"`
		Year                int    `json:"year"`
		IncludeTransactions bool   `json:"include_transactions"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return "", time.Time{}, time.Time{}, false, fmt.Errorf("invalid request body: %w", err)
	}
	messageTemplate := body.MessageTemplate
	if messageTemplate == "" {
		messageTemplate = defaultTemplate
	}
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
		return "", time.Time{}, time.Time{}, false, fmt.Errorf("invalid month format: %w", err)
	}
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	return messageTemplate, startDate, endDate, body.IncludeTransactions, nil
}

// sendBalanceNotifications sends notifications to a slice of users and returns success/fail counts and details
func sendBalanceNotifications(
	h *WhatsAppHandler,
	users []models.User,
	balances []models.UserBalance,
	messageTemplate string,
	startDate, endDate time.Time,
	includeTransactions bool,
	delay time.Duration,
) (successCount int, failCount int, failedUsers []string) {
	if len(users) != len(balances) {
		log.Errorf("users and balances slices have different lengths: %d vs %d", len(users), len(balances))
		failCount = len(users)
		for _, user := range users {
			failedUsers = append(failedUsers, fmt.Sprintf("%s (internal error: mismatched slices)", user.Name))
		}
		return
	}

	for i, user := range users {
		userBalance := balances[i]
		if user.Phone == "" {
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (no phone number)", user.Name))
			continue
		}
		err := h.sendBalanceNotification(user, userBalance, messageTemplate, startDate, endDate, includeTransactions)
		if err != nil {
			log.Printf("Failed to send WhatsApp notification to %s (%s): %v", user.Name, user.Phone, err)
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (%v)", user.Name, err))
		} else {
			successCount++
		}
		if delay > 0 && i < len(users)-1 {
			time.Sleep(delay)
		}
	}
	return
}

// notifyUserBalances is a modular handler for sending WhatsApp notifications to one or all users.
// If employeeID is 0, it sends to all users; otherwise, to the specified user.
func (h *WhatsAppHandler) notifyUserBalances(w http.ResponseWriter, r *http.Request, employeeID int64) {
	client := h.GetWhatsAppClient()
	if client == nil || !client.IsLoggedIn() || !client.IsConnected() {
		log.Warn("WhatsApp client is not available")
		common.RespondWithError(w, http.StatusInternalServerError, "WhatsApp client is not available")
		return
	}

	// Parse request body and date range
	messageTemplate, startDate, endDate, includeTransactions, err := parseBalanceNotificationRequest(r)
	if err != nil {
		log.Errorf("Failed to parse notification request: %v", err)
		common.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var users []models.User
	var balances []models.UserBalance
	var notificationDelayToUse time.Duration = 0
	var target string

	if employeeID != 0 {
		// Single user
		user, err := h.DB.GetUser(employeeID)
		if err != nil {
			log.Errorf("User with employee ID %d not found: %v", employeeID, err)
			common.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("User with employee ID %d not found", employeeID))
			return
		}
		if user.Phone == "" {
			log.Warnf("User with employee ID %d does not have a phone number", employeeID)
			common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("User with employee ID %d does not have a phone number", employeeID))
			return
		}
		userBalance, err := h.DB.GetUserBalanceByUserID(user.ID)
		if err != nil {
			log.Errorf("Failed to get user balance for user ID %d: %v", user.ID, err)
			common.RespondWithError(w, http.StatusInternalServerError, "Failed to get user balance")
			return
		}
		users = []models.User{*user}
		balances = []models.UserBalance{userBalance}
		target = user.Name
	} else {
		// All users
		userBalances, err := h.DB.GetUsersBalances()
		if err != nil {
			log.Errorf("Failed to get all user balances: %v", err)
			common.RespondWithError(w, http.StatusInternalServerError, "Failed to get users' balances")
			return
		}
		for _, balance := range userBalances {
			if !balance.UserActive || balance.Phone == "" {
				continue
			}
			users = append(users, models.User{
				ID:    balance.UserID,
				Name:  balance.UserName,
				Phone: balance.Phone,
			})
			balances = append(balances, models.UserBalance{
				Balance: balance.Balance,
			})
		}
		notificationDelayToUse = notificationDelay
		target = "all users"
	}

	successCount, failCount, failedUsers := sendBalanceNotifications(h, users, balances, messageTemplate, startDate, endDate, includeTransactions, notificationDelayToUse)

	// Consistent response structure for both single and all
	resp := map[string]any{
		"success": failCount == 0,
		"message": fmt.Sprintf("Sent %d notification(s) to %s, %d failed", successCount, target, failCount),
		"details": map[string]any{
			"success_count": successCount,
			"fail_count":    failCount,
			"failed_users":  failedUsers,
		},
	}
	status := http.StatusOK
	if failCount > 0 {
		status = http.StatusInternalServerError
	}
	common.RespondWithJSON(w, status, resp)
}

// NotifyUserBalance handles sending WhatsApp notification to a single user
func (h *WhatsAppHandler) NotifyUserBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	employeeID, err := h.ParseID(vars, "id")
	if err != nil {
		log.Warnf("Failed to parse employee ID from request: %v", err)
		common.RespondWithError(w, http.StatusBadRequest, "Employee ID is required")
		return
	}
	h.notifyUserBalances(w, r, employeeID)
}

// NotifyAllUsersBalances handles sending WhatsApp notifications to all users
func (h *WhatsAppHandler) NotifyAllUsersBalances(w http.ResponseWriter, r *http.Request) {
	h.notifyUserBalances(w, r, 0)
}

// SendDocumentMessage sends a document message to a user's WhatsApp number
func (h *WhatsAppHandler) SendDocumentMessage(phoneNumber string, fileName string, fileData []byte, mimeType string) error {
	recipient, err := h.getWhatsAppRecipient(phoneNumber)
	if err != nil {
		return err
	}

	log.Infof("Sending WhatsApp document to %s: %s", recipient, fileName)

	client := h.GetWhatsAppClient()
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
