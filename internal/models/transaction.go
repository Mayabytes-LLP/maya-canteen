package models

import (
	"time"
)

// Transaction represents a financial transaction in the system
type Transaction struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"` // e.g., "deposit", "withdrawal", "purchase"
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// EmployeeTransaction represents a financial transaction with user details
type EmployeeTransaction struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	UserName        string    `json:"user_name"`
	EmployeeID      string    `json:"employee_id"`
	Department      string    `json:"department"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GetID returns the transaction ID
func (t *Transaction) GetID() int64 {
	return t.ID
}

// SetID sets the transaction ID
func (t *Transaction) SetID(id int64) {
	t.ID = id
}

// SetCreatedAt sets the created timestamp
func (t *Transaction) SetCreatedAt(timestamp any) {
	if ts, ok := timestamp.(time.Time); ok {
		t.CreatedAt = ts
	}
}

// SetUpdatedAt sets the updated timestamp
func (t *Transaction) SetUpdatedAt(timestamp any) {
	if ts, ok := timestamp.(time.Time); ok {
		t.UpdatedAt = ts
	}
}
