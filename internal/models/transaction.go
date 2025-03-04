package models

import (
	"database/sql"
	"time"
)

type Transaction struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	EmployeeID string    `json:"employee_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Products   []Product `json:"products"`
}

type TransactionModel struct {
	DB *sql.DB
}

func NewTransactionModel(db *sql.DB) *TransactionModel {
	return &TransactionModel{DB: db}
}

func (m *TransactionModel) InitTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS transactions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      employee_id TEXT NOT NULL,
      created_at DATETIME NOT NULL,
      updated_at DATETIME NOT NULL,
      products TEXT NOT NULL
    )
  `
	_, err := m.DB.Exec(query)
	return err
}

func (m *TransactionModel) Create(transaction *Transaction) error {
	query := `
    INSERT INTO transactions (name, employee_id, created_at, updated_at, products)
    VALUES (?, ?, ?, ?, ?)
  `

	now := time.Now()
	result, err := m.DB.Exec(query, transaction.Name, transaction.EmployeeID, now, now, transaction.Products)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	transaction.ID = id
	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	return nil
}
