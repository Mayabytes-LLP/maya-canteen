package handlers

import (
	"log"
	"maya-canteen/internal/database"
	"maya-canteen/internal/errors"
	"maya-canteen/internal/handlers/common"
	"maya-canteen/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	common.BaseHandler
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(db database.Service) *TransactionHandler {
	return &TransactionHandler{
		BaseHandler: common.NewBaseHandler(db),
	}
}

// CreateTransaction handles POST /api/transactions
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction
	if err := h.DecodeJSON(r, &transaction); err != nil {
		h.HandleError(w, err)
		return
	}

	if err := h.DB.CreateTransaction(&transaction); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusCreated, transaction)
}

// GetAllTransactions handles GET /api/transactions
func (h *TransactionHandler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := h.DB.GetAllTransactions()
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, transactions)
}

// GetLatestTransactions handles GET /api/transactions/latest
func (h *TransactionHandler) GetLatestTransactions(w http.ResponseWriter, r *http.Request) {
	// Get limit from query parameter, default to 10 if not provided
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default limit

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			h.HandleError(w, errors.InvalidInput("Invalid limit parameter. Must be a number."))
			return
		}

		// Ensure limit is positive
		if parsedLimit <= 0 {
			h.HandleError(w, errors.InvalidInput("Limit must be a positive number."))
			return
		}

		limit = parsedLimit
	}

	transactions, err := h.DB.GetLatestTransactions(limit)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, transactions)
}

// GetTransaction handles GET /api/transactions/{id}
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	transaction, err := h.DB.GetTransaction(id)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	if transaction == nil {
		h.HandleError(w, errors.NotFound("Transaction", id))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, transaction)
}

// UpdateTransaction handles PUT /api/transactions/{id}
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	var transaction models.Transaction
	if err := h.DecodeJSON(r, &transaction); err != nil {
		h.HandleError(w, err)
		return
	}
	transaction.ID = id

	if err := h.DB.UpdateTransaction(&transaction); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, transaction)
}

// DeleteTransaction handles DELETE /api/transactions/{id}
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	if err := h.DB.DeleteTransaction(id); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusNoContent, nil)
}

// GetTransactionsByUserID handles GET /api/users/{user_id}/transactions
func (h *TransactionHandler) GetTransactionsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := h.ParseID(vars, "user_id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	transactions, err := h.DB.GetTransactionsByUserID(userID)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, transactions)
}

// DateRangeRequest represents the request body for date range queries
type DateRangeRequest struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// GetTransactionsByDateRange handles POST /api/transactions/date-range
func (h *TransactionHandler) GetTransactionsByDateRange(w http.ResponseWriter, r *http.Request) {
	var dateRange DateRangeRequest
	if err := h.DecodeJSON(r, &dateRange); err != nil {
		h.HandleError(w, err)
		return
	}

	// Parse dates from the shadcn date picker format (ISO 8601: YYYY-MM-DD)
	startDate, err := time.Parse("2006-01-02", dateRange.StartDate)
	if err != nil {
		h.HandleError(w, errors.InvalidInput("Invalid start date format. Expected YYYY-MM-DD"))
		return
	}

	endDate, err := time.Parse("2006-01-02", dateRange.EndDate)
	if err != nil {
		h.HandleError(w, errors.InvalidInput("Invalid end date format. Expected YYYY-MM-DD"))
		return
	}

	// Validate date range
	if endDate.Before(startDate) {
		h.HandleError(w, errors.InvalidInput("End date cannot be before start date"))
		return
	}

	transactions, err := h.DB.GetTransactionsByDateRange(startDate, endDate)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, transactions)
}

// GetUsersBalances handles GET /api/users/balances
func (h *TransactionHandler) GetUsersBalances(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to fetch user balances")
	balances, err := h.DB.GetUsersBalances()
	if err != nil {
		log.Printf("Error fetching user balances: %v", err)
		h.HandleError(w, errors.Internal(err))
		return
	}

	log.Printf("Fetched user balances: %v", balances)
	common.RespondWithSuccess(w, http.StatusOK, balances)
}
