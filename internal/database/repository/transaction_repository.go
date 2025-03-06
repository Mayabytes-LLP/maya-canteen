package repository

import (
	"database/sql"
	"log"
	"maya-canteen/internal/models"
	"time"
)

// TransactionRepository handles all database operations related to transactions
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// InitTable initializes the transactions table
func (r *TransactionRepository) InitTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			description TEXT,
			transaction_type TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`
	_, err := r.db.Exec(query)
	return err
}

// Create inserts a new transaction into the database
func (r *TransactionRepository) Create(transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (user_id, amount, description, transaction_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(
		query,
		transaction.UserID,
		transaction.Amount,
		transaction.Description,
		transaction.TransactionType,
		now,
		now,
	)
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

// GetAll retrieves all transactions from the database
func (r *TransactionRepository) GetAll() ([]models.Transaction, error) {
	query := `SELECT * FROM transactions ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.TransactionType,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// Get retrieves a single transaction by ID
func (r *TransactionRepository) Get(id int64) (*models.Transaction, error) {
	query := `SELECT * FROM transactions WHERE id = ?`
	var transaction models.Transaction
	err := r.db.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.Amount,
		&transaction.Description,
		&transaction.TransactionType,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// Update updates an existing transaction
func (r *TransactionRepository) Update(transaction *models.Transaction) error {
	query := `
		UPDATE transactions
		SET user_id = ?, amount = ?, description = ?, transaction_type = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.Exec(
		query,
		transaction.UserID,
		transaction.Amount,
		transaction.Description,
		transaction.TransactionType,
		now,
		transaction.ID,
	)
	if err != nil {
		return err
	}
	transaction.UpdatedAt = now
	return nil
}

// Delete removes a transaction by ID
func (r *TransactionRepository) Delete(id int64) error {
	query := `DELETE FROM transactions WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// GetByUserID retrieves all transactions for a specific user
func (r *TransactionRepository) GetByUserID(userID int64) ([]models.Transaction, error) {
	query := `SELECT * FROM transactions WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.TransactionType,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// GetByDateRange retrieves all transactions within a specific date range
func (r *TransactionRepository) GetByDateRange(startDate, endDate time.Time) ([]models.Transaction, error) {
	// Adjust endDate to include the entire day
	endDate = endDate.Add(24 * time.Hour).Add(-1 * time.Second)

	query := `SELECT * FROM transactions WHERE created_at BETWEEN ? AND ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.TransactionType,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// GetLatest retrieves the latest transactions with a limit
func (r *TransactionRepository) GetLatest(limit int) ([]models.Transaction, error) {
	query := `SELECT * FROM transactions ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.TransactionType,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// GetUsersBalances retrieves the total balance for each user
func (r *TransactionRepository) GetUsersBalances() ([]models.UserBalance, error) {
	query := `
        SELECT users.id, users.name, users.employee_id,
               COALESCE(SUM(CASE WHEN transactions.transaction_type = 'deposit' THEN transactions.amount ELSE -transactions.amount END), 0) AS balance
        FROM users
        LEFT JOIN transactions ON users.id = transactions.user_id
        GROUP BY users.id
    `
	// 	`
	// 	SELECT users.id, users.name, users.employee_id,
	// 	       SUM(CASE WHEN transactions.transaction_type = 'deposit' THEN transactions.amount ELSE -transactions.amount END) AS balance
	// 	FROM users
	// 	LEFT JOIN transactions ON users.id = transactions.user_id
	// 	GROUP BY users.id
	// `

	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var balances []models.UserBalance
	for rows.Next() {
		var balance models.UserBalance
		err := rows.Scan(&balance.UserID, &balance.UserName, &balance.EmployeeID, &balance.Balance)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		balances = append(balances, balance)
	}
	log.Printf("Fetched balances: %v", balances)
	return balances, nil
}
