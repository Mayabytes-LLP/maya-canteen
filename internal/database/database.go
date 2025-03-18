package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"maya-canteen/internal/database/repository"
	"maya-canteen/internal/models"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// GetDB returns the underlying database connection
	GetDB() *sql.DB

	// User-related operations
	InitUserTable() error
	CreateUser(user *models.User) error
	GetAllUsers() ([]models.User, error)
	GetUser(id int64) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int64) error

	// Transaction-related operations
	InitTransactionTable() error
	CreateTransaction(transaction *models.Transaction) error
	GetAllTransactions() ([]models.Transaction, error)
	GetLatestTransactions(limit int) ([]models.Transaction, error)
	GetTransaction(id int64) (*models.Transaction, error)
	UpdateTransaction(transaction *models.Transaction) error
	DeleteTransaction(id int64) error
	GetTransactionsByUserID(userID int64, limit int) ([]models.EmployeeTransaction, error)
	GetTransactionsByDateRange(startDate, endDate time.Time) ([]models.Transaction, error)
	GetUsersBalances() ([]models.UserBalance, error)

	// Product-related operations
	InitProductTable() error
	CreateProduct(product *models.Product) error
	GetAllProducts() ([]models.Product, error)
	GetProduct(id int64) (*models.Product, error)
	UpdateProduct(product *models.Product) error
	DeleteProduct(id int64) error
}

type service struct {
	db                    *sql.DB
	repositoryFactory     *repository.RepositoryFactory
	userRepository        repository.UserRepositoryInterface
	transactionRepository repository.TransactionRepositoryInterface
	productRepository     repository.ProductRepositoryInterface
}

var (
	dburl      = os.Getenv("BLUEPRINT_DB_URL")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	// check if path is absolute or just contain the file name example ./db/cateen.db if db folder does not exist create it
	dbPath := dburl[:strings.LastIndex(dburl, "/")]
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Println("Creating DB Path", dbPath)
		os.MkdirAll(dbPath, os.ModePerm)
	} else {
		log.Println("DB Path exists at ", dbPath)
	}

	log.Println("DB URL", dburl)
	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatalf("Error opening database: %v", err)
	}

	// Create repository factory
	repoFactory := repository.NewRepositoryFactory(db)

	dbInstance = &service{
		db:                    db,
		repositoryFactory:     repoFactory,
		userRepository:        repoFactory.NewUserRepository(),
		transactionRepository: repoFactory.NewTransactionRepository(),
		productRepository:     repoFactory.NewProductRepository(),
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}

// GetDB returns the underlying database connection
func (s *service) GetDB() *sql.DB {
	return s.db
}

// User-related operations
func (s *service) InitUserTable() error {
	return s.userRepository.InitTable()
}

func (s *service) CreateUser(user *models.User) error {
	return s.userRepository.Create(user)
}

func (s *service) GetAllUsers() ([]models.User, error) {
	return s.userRepository.GetAll()
}

func (s *service) GetUser(id int64) (*models.User, error) {
	return s.userRepository.Get(id)
}

func (s *service) UpdateUser(user *models.User) error {
	return s.userRepository.Update(user)
}

func (s *service) DeleteUser(id int64) error {
	return s.userRepository.Delete(id)
}

// Transaction-related operations
func (s *service) InitTransactionTable() error {
	return s.transactionRepository.InitTable()
}

func (s *service) CreateTransaction(transaction *models.Transaction) error {
	return s.transactionRepository.Create(transaction)
}

func (s *service) GetAllTransactions() ([]models.Transaction, error) {
	return s.transactionRepository.GetAll()
}

func (s *service) GetLatestTransactions(limit int) ([]models.Transaction, error) {
	return s.transactionRepository.GetLatest(limit)
}

func (s *service) GetTransaction(id int64) (*models.Transaction, error) {
	return s.transactionRepository.Get(id)
}

func (s *service) UpdateTransaction(transaction *models.Transaction) error {
	return s.transactionRepository.Update(transaction)
}

func (s *service) DeleteTransaction(id int64) error {
	return s.transactionRepository.Delete(id)
}

func (s *service) GetTransactionsByUserID(userID int64, limit int) ([]models.EmployeeTransaction, error) {
	return s.transactionRepository.GetByUserID(userID, limit)
}

func (s *service) GetTransactionsByDateRange(startDate, endDate time.Time) ([]models.Transaction, error) {
	return s.transactionRepository.GetByDateRange(startDate, endDate)
}

func (s *service) GetUsersBalances() ([]models.UserBalance, error) {
	return s.transactionRepository.GetUsersBalances()
}

// Product-related operations
func (s *service) InitProductTable() error {
	return s.productRepository.InitTable()
}

func (s *service) CreateProduct(product *models.Product) error {
	return s.productRepository.Create(product)
}

func (s *service) GetAllProducts() ([]models.Product, error) {
	return s.productRepository.GetAll()
}

func (s *service) GetProduct(id int64) (*models.Product, error) {
	return s.productRepository.Get(id)
}

func (s *service) UpdateProduct(product *models.Product) error {
	return s.productRepository.Update(product)
}

func (s *service) DeleteProduct(id int64) error {
	return s.productRepository.Delete(id)
}
