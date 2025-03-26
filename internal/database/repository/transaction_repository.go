package repository

import (
	"database/sql"
	"maya-canteen/internal/models"
	"time"

	log "github.com/sirupsen/logrus"
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
	if err != nil {
		log.Errorf("Error creating transactions table: %v", err)
	}

	log.Info("Created Transactions Table")
	return nil
}

// Create inserts a new transaction into the database
func (r *TransactionRepository) Create(transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (
      user_id, 
      amount, 
      description, 
      transaction_type, 
      created_at, 
      updated_at
    )
		VALUES (
      ?, ?, ?, ?, ?, ?
    )
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
		log.Errorf("Error creating transaction: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Error getting last insert ID: %v", err)
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
			log.Errorf("Error scanning row: %v", err)
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// Get retrieves a single transaction by ID
func (r *TransactionRepository) Get(id int64) (*models.Transaction, error) {
	query := `
    SELECT * 
    FROM transactions 
    WHERE id = ?
  `
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
		log.Errorf("Transaction with ID %d not found", id)
		return nil, nil
	}
	if err != nil {
		log.Errorf("Error scanning transaction row: %v", err)
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
	if err != nil {
		log.Errorf("Error deleting transaction: %v", err)
		return err
	}

	return nil
}

// GetByUserID retrieves all transactions for a specific user
func (r *TransactionRepository) GetByUserID(userID int64, limit int) ([]models.EmployeeTransaction, error) {
	query := `
	  SELECT
        users.name,
        users.employee_id,
        users.department,
        transactions.id,
        transactions.user_id,
        transactions.amount,
        transactions.description,
        transactions.transaction_type,
        transactions.created_at,
        transactions.updated_at
	  FROM transactions
	  LEFT JOIN users ON transactions.user_id = users.id
	  WHERE users.employee_id = ?
	  ORDER BY transactions.created_at DESC
		LIMIT ?;
	`
	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		log.Errorf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var transactions []models.EmployeeTransaction
	for rows.Next() {
		var transaction models.EmployeeTransaction
		err := rows.Scan(
			&transaction.UserName,
			&transaction.EmployeeID,
			&transaction.Department,
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Description,
			&transaction.TransactionType,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			log.Errorf("Error scanning row: %v", err)
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("Error with transaction rows: %v", err)
		return nil, err
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
		log.Errorf("Error executing query: %v", err)
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
			log.Errorf("Error scanning row: %v", err)
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
		log.Errorf("Error executing query: %v", err)
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
			log.Errorf("Error scanning row: %v", err)
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// GetUsersBalances retrieves the total balance for each user
func (r *TransactionRepository) GetUsersBalances() ([]models.UserBalance, error) {
	query := `
        SELECT 
          users.id,
          users.name, 
          users.employee_id, 
          users.department, 
          users.phone,
          COALESCE(SUM(CASE WHEN transactions.transaction_type = 'deposit' THEN transactions.amount ELSE -transactions.amount END), 0) AS balance
        FROM users
        LEFT JOIN transactions ON users.id = transactions.user_id
        GROUP BY users.id
    `

	rows, err := r.db.Query(query)
	if err != nil {
		log.Errorf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var balances []models.UserBalance
	for rows.Next() {
		var balance models.UserBalance
		err := rows.Scan(
			&balance.UserID,
			&balance.UserName,
			&balance.EmployeeID,
			&balance.Department,
			&balance.Phone,
			&balance.Balance,
		)
		if err != nil {
			log.Errorf("Error scanning row: %v", err)
			return nil, err
		}
		balances = append(balances, balance)
	}
	return balances, nil
}

func (r *TransactionRepository) GetUserBalanceByID(userID int64) (models.UserBalance, error) {
	query := `
		SELECT 
      users.id, 
      users.name, 
      users.employee_id, 
      users.department, 
      users.phone,
		  COALESCE(SUM(CASE WHEN transactions.transaction_type = 'deposit' THEN transactions.amount ELSE -transactions.amount END), 0) AS balance
		FROM users
		LEFT JOIN transactions ON users.id = transactions.user_id
		WHERE users.id = ?
		GROUP BY users.id
	`
	var balance models.UserBalance
	err := r.db.QueryRow(query, userID).Scan(
		&balance.UserID,
		&balance.UserName,
		&balance.EmployeeID,
		&balance.Department,
		&balance.Phone,
		&balance.Balance,
	)
	if err != nil {
		log.Errorf("Error scanning row: %v", err)
		return models.UserBalance{}, err
	}
	return balance, nil
}
