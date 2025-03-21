package routes

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers"

	"github.com/gorilla/mux"
)

// RegisterTransactionRoutes registers all transaction-related routes
func RegisterTransactionRoutes(router *mux.Router, db database.Service) {
	// Create handler
	transactionHandler := handlers.NewTransactionHandler(db)

	// Register routes
	router.HandleFunc("/api/transactions", transactionHandler.CreateTransaction).Methods("POST")
	router.HandleFunc("/api/transactions", transactionHandler.GetAllTransactions).Methods("GET")
	router.HandleFunc("/api/transactions/latest", transactionHandler.GetLatestTransactions).Methods("GET")
	router.HandleFunc("/api/transactions/date-range", transactionHandler.GetTransactionsByDateRange).Methods("POST")
	router.HandleFunc("/api/transactions/{id}", transactionHandler.GetTransaction).Methods("GET")
	router.HandleFunc("/api/transactions/{id}", transactionHandler.UpdateTransaction).Methods("PUT")
	router.HandleFunc("/api/transactions/{id}", transactionHandler.DeleteTransaction).Methods("DELETE")
	router.HandleFunc("/api/users/{user_id}/transactions", transactionHandler.GetTransactionsByUserID).Methods("GET")
	router.HandleFunc("/api/users/{user_id}/balance", transactionHandler.GetUserBalanceByUserID).Methods("GET")
	router.HandleFunc("/api/users/balances", transactionHandler.GetUsersBalances).Methods("GET")
}
