package repository

import (
	"database/sql"
	"maya-canteen/internal/models"
	"time"
)

// Repository is the interface that all repositories must implement
type Repository interface {
	// Common methods that all repositories should implement
	InitTable() error
}

// UserRepositoryInterface defines operations for user data
type UserRepositoryInterface interface {
	Repository
	Create(user *models.User) error
	GetAll() ([]models.User, error)
	Get(id int64) (*models.User, error)
	Update(user *models.User) error
	Delete(id int64) error
}

// TransactionRepositoryInterface defines operations for transaction data
type TransactionRepositoryInterface interface {
	Repository
	Create(transaction *models.Transaction) error
	GetAll() ([]models.Transaction, error)
	Get(id int64) (*models.Transaction, error)
	Update(transaction *models.Transaction) error
	Delete(id int64) error
	GetByUserID(userID int64, limit int) ([]models.EmployeeTransaction, error)
	GetByDateRange(startDate, endDate time.Time) ([]models.Transaction, error)
	GetLatest(limit int) ([]models.Transaction, error)
	GetUsersBalances() ([]models.UserBalance, error)
	GetUserBalanceByID(userID int64) (models.UserBalance, error)
}

// ProductRepositoryInterface defines operations for product data
type ProductRepositoryInterface interface {
	Repository
	Create(product *models.Product) error
	GetAll() ([]models.Product, error)
	Get(id int64) (*models.Product, error)
	Update(product *models.Product) error
	Delete(id int64) error
}

// RepositoryFactory creates and returns repositories
type RepositoryFactory struct {
	db *sql.DB
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *sql.DB) *RepositoryFactory {
	return &RepositoryFactory{db: db}
}

// NewUserRepository creates a new user repository
func (f *RepositoryFactory) NewUserRepository() UserRepositoryInterface {
	return NewUserRepository(f.db)
}

// NewTransactionRepository creates a new transaction repository
func (f *RepositoryFactory) NewTransactionRepository() TransactionRepositoryInterface {
	return NewTransactionRepository(f.db)
}

// NewProductRepository creates a new product repository
func (f *RepositoryFactory) NewProductRepository() ProductRepositoryInterface {
	return NewProductRepository(f.db)
}
