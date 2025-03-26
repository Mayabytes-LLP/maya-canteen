package handlers

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/errors"
	"maya-canteen/internal/handlers/common"
	"maya-canteen/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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

// TransactionRequest represents the request body for creating a transaction with products
type TransactionRequest struct {
	UserID          int64                   `json:"user_id"`
	Amount          float64                 `json:"amount"`
	Description     string                  `json:"description"`
	TransactionType string                  `json:"transaction_type"`
	Products        []TransactionProductDTO `json:"products,omitempty"`
}

// TransactionProductDTO represents a product in a transaction request
type TransactionProductDTO struct {
	ProductID    int64   `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	IsSingleUnit bool    `json:"is_single_unit"`
}

// CreateTransaction handles POST /api/transactions
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var request TransactionRequest

	if err := h.DecodeJSON(r, &request); err != nil {
		log.Errorf("Error decoding JSON: %v", err)
		h.HandleError(w, err)
		return
	}

	// Create the transaction model
	transaction := models.Transaction{
		UserID:          request.UserID,
		Amount:          request.Amount,
		Description:     request.Description,
		TransactionType: request.TransactionType,
	}

	// If it's a deposit or has no products, use the simple transaction creation
	if request.TransactionType == "deposit" || len(request.Products) == 0 {
		if err := h.DB.CreateTransaction(&transaction); err != nil {
			h.HandleError(w, errors.Internal(err))
			return
		}
	} else {
		// Convert DTO to models.TransactionProduct
		var transactionProducts []models.TransactionProduct
		for _, productDTO := range request.Products {
			product := models.TransactionProduct{
				ProductID:    productDTO.ProductID,
				ProductName:  productDTO.ProductName,
				Quantity:     productDTO.Quantity,
				UnitPrice:    productDTO.UnitPrice,
				IsSingleUnit: productDTO.IsSingleUnit,
			}
			transactionProducts = append(transactionProducts, product)
		}

		// Create transaction with products
		if err := h.DB.CreateTransactionWithProducts(&transaction, transactionProducts); err != nil {
			log.Errorf("Error creating transaction with products: %v", err)
			h.HandleError(w, errors.Internal(err))
			return
		}
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

	// Get associated products if this is a purchase transaction
	var transactionProducts []models.TransactionProduct = nil
	if transaction.TransactionType == "purchase" {
		transactionProducts, err = h.DB.GetTransactionProducts(id)
		if err != nil {
			h.HandleError(w, errors.Internal(err))
			return
		}
	}

	// Create response with transaction and its products
	response := struct {
		*models.Transaction
		Products []models.TransactionProduct `json:"products,omitempty"`
	}{
		Transaction: transaction,
		Products:    transactionProducts,
	}

	common.RespondWithSuccess(w, http.StatusOK, response)
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

	transactions, err := h.DB.GetTransactionsByUserID(userID, limit)
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
		log.Errorf("Error decoding JSON: %v", err)
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

	common.RespondWithSuccess(w, http.StatusOK, balances)
}

// GetUserBalanceByUserID handles GET /api/users/{user_id}/balance
func (h *TransactionHandler) GetUserBalanceByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := h.ParseID(vars, "user_id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	balance, err := h.DB.GetUserBalanceByUserID(userID)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, balance)
}

// GetTransactionProducts handles GET /api/transactions/{id}/products
func (h *TransactionHandler) GetTransactionProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionID, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	products, err := h.DB.GetTransactionProducts(transactionID)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, products)
}

// GetProductSalesSummary handles POST /api/reports/product-sales
func (h *TransactionHandler) GetProductSalesSummary(w http.ResponseWriter, r *http.Request) {
	var dateRange DateRangeRequest
	if err := h.DecodeJSON(r, &dateRange); err != nil {
		h.HandleError(w, err)
		return
	}

	// Parse dates
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

	summary, err := h.DB.GetProductSalesSummary(startDate, endDate)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, summary)
}

// GetTransactionProductDetails handles POST /api/reports/transaction-products
func (h *TransactionHandler) GetTransactionProductDetails(w http.ResponseWriter, r *http.Request) {
	var dateRange DateRangeRequest
	if err := h.DecodeJSON(r, &dateRange); err != nil {
		h.HandleError(w, err)
		return
	}

	// Parse dates
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

	details, err := h.DB.GetTransactionProductDetails(startDate, endDate)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, details)
}
