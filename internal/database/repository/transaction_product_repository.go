package repository

import (
	"database/sql"
	"maya-canteen/internal/models"
	"time"

	log "github.com/sirupsen/logrus"
)

// TransactionProductRepository handles all database operations related to transaction products
type TransactionProductRepository struct {
	db *sql.DB
}

// NewTransactionProductRepository creates a new transaction product repository
func NewTransactionProductRepository(db *sql.DB) *TransactionProductRepository {
	return &TransactionProductRepository{db: db}
}

// InitTable initializes the transaction_products table
func (r *TransactionProductRepository) InitTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS transaction_products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transaction_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			product_name TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL,
			is_single_unit BOOLEAN NOT NULL DEFAULT false,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
		)
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Errorf("Error creating transaction_products table: %v", err)
		return err
	}

	log.Info("Created Transaction Products Table")
	return nil
}

// Create inserts a new transaction product into the database
func (r *TransactionProductRepository) Create(transactionProduct *models.TransactionProduct) error {
	query := `
		INSERT INTO transaction_products (
			transaction_id,
			product_id,
			product_name,
			quantity,
			unit_price,
			is_single_unit,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(
		query,
		transactionProduct.TransactionID,
		transactionProduct.ProductID,
		transactionProduct.ProductName,
		transactionProduct.Quantity,
		transactionProduct.UnitPrice,
		transactionProduct.IsSingleUnit,
		now,
		now,
	)
	if err != nil {
		log.Errorf("Error inserting transaction product: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Error getting last insert ID: %v", err)
		return err
	}
	transactionProduct.ID = id
	transactionProduct.CreatedAt = now
	transactionProduct.UpdatedAt = now
	return nil
}

// GetByTransactionID retrieves all products for a specific transaction
func (r *TransactionProductRepository) GetByTransactionID(transactionID int64) ([]models.TransactionProduct, error) {
	query := `
		SELECT
			id,
			transaction_id,
			product_id,
			product_name,
			quantity,
			unit_price,
			is_single_unit,
			created_at,
			updated_at
		FROM transaction_products
		WHERE transaction_id = ?
		ORDER BY id ASC
	`
	rows, err := r.db.Query(query, transactionID)
	if err != nil {
		log.Errorf("Error executing transaction product query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var products []models.TransactionProduct
	for rows.Next() {
		var product models.TransactionProduct
		err := rows.Scan(
			&product.ID,
			&product.TransactionID,
			&product.ProductID,
			&product.ProductName,
			&product.Quantity,
			&product.UnitPrice,
			&product.IsSingleUnit,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("Error with transaction product rows: %v", err)
		return nil, err
	}
	return products, nil
}

// GetProductSalesSummary retrieves sales statistics for all products
func (r *TransactionProductRepository) GetProductSalesSummary(startDate, endDate time.Time) ([]models.ProductSalesSummary, error) {
	// Adjust endDate to include the entire day
	endDate = endDate.Add(24 * time.Hour).Add(-1 * time.Second)

	query := `
		SELECT
			p.id AS product_id,
			p.name AS product_name,
			p.type AS product_type,
			SUM(tp.quantity) AS total_quantity,
			SUM(tp.quantity * tp.unit_price) AS total_sales,
			SUM(CASE WHEN tp.is_single_unit = 1 THEN tp.quantity ELSE 0 END) AS single_unit_sold,
			SUM(CASE WHEN tp.is_single_unit = 0 THEN tp.quantity ELSE 0 END) AS full_unit_sold
		FROM transaction_products tp
		JOIN products p ON tp.product_id = p.id
		JOIN transactions t ON tp.transaction_id = t.id
		WHERE t.transaction_type = 'purchase'
		AND t.created_at BETWEEN ? AND ?
		GROUP BY p.id, p.name, p.type
		ORDER BY total_sales DESC
	`
	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		log.Errorf("Error executing product sales summary query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var summaries []models.ProductSalesSummary
	for rows.Next() {
		var summary models.ProductSalesSummary
		err := rows.Scan(
			&summary.ProductID,
			&summary.ProductName,
			&summary.ProductType,
			&summary.TotalQuantity,
			&summary.TotalSales,
			&summary.SingleUnitSold,
			&summary.FullUnitSold,
		)
		if err != nil {
			log.Errorf("Error scanning product sales summary row: %v", err)
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("Error with product sales summary rows: %v", err)
		return nil, err
	}
	return summaries, nil
}

// GetTransactionProductDetails retrieves product details with transaction context
func (r *TransactionProductRepository) GetTransactionProductDetails(startDate, endDate time.Time) ([]models.TransactionProductDetail, error) {
	// Adjust endDate to include the entire day
	endDate = endDate.Add(24 * time.Hour).Add(-1 * time.Second)

	query := `
		SELECT
			tp.id,
			tp.transaction_id,
			tp.product_id,
			tp.product_name,
			p.type AS product_type,
			tp.quantity,
			tp.unit_price,
			(tp.quantity * tp.unit_price) AS total_price,
			tp.is_single_unit,
			tp.created_at,
			tp.updated_at
		FROM transaction_products tp
		JOIN products p ON tp.product_id = p.id
		JOIN transactions t ON tp.transaction_id = t.id
		WHERE t.transaction_type = 'purchase'
		AND t.created_at BETWEEN ? AND ?
		ORDER BY tp.transaction_id DESC, tp.id ASC
	`
	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		log.Errorf("Error executing transaction product detail query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var details []models.TransactionProductDetail
	for rows.Next() {
		var detail models.TransactionProductDetail
		err := rows.Scan(
			&detail.ID,
			&detail.TransactionID,
			&detail.ProductID,
			&detail.ProductName,
			&detail.ProductType,
			&detail.Quantity,
			&detail.UnitPrice,
			&detail.TotalPrice,
			&detail.IsSingleUnit,
			&detail.CreatedAt,
			&detail.UpdatedAt,
		)
		if err != nil {
			log.Errorf("Error scanning transaction product detail row: %v", err)
			return nil, err
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("Error with transaction product detail rows: %v", err)
		return nil, err
	}
	return details, nil
}
