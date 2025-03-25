package repository

import (
	"database/sql"
	"fmt"
	"maya-canteen/internal/models"
	"time"
)

// ProductRepository handles all database operations related to products
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// InitTable initializes the products table
func (r *ProductRepository) InitTable() error {
	r.addActiveColumnIfNeeded()

	query := `
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			price REAL NOT NULL,
			type TEXT NOT NULL DEFAULT 'regular',
			active BOOLEAN NOT NULL DEFAULT true,
      is_single_unit BOOLEAN NOT NULL DEFAULT false,
			single_unit_price REAL NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`
	_, err := r.db.Exec(query)
	return err
}

// addActiveColumnIfNeeded checks if the active column exists and adds it if needed
func (r *ProductRepository) addActiveColumnIfNeeded() {
	// Check if the column exists
	var colExists bool
	err := r.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('products')
		WHERE name = 'active'
	`).Scan(&colExists)

	if err != nil || colExists {
		return // Either error occurred or column already exists
	}

	// Add the column if it doesn't exist
	_, err = r.db.Exec(`ALTER TABLE products ADD COLUMN active BOOLEAN NOT NULL DEFAULT 1`)
	if err != nil {
		fmt.Printf("Error adding active column: %v\n", err)
	} else {
		fmt.Println("Added active column to products table")
	}
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(product *models.Product) error {
	query := `
		INSERT INTO products (
			name,
			description,
			price,
			type,
			active,
			is_single_unit,
			single_unit_price,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.Exec(
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Type,
		product.Active,
		product.IsSingleUnit,
		product.SingleUnitPrice,
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
	product.ID = id
	product.CreatedAt = now
	product.UpdatedAt = now
	return nil
}

// GetAll retrieves all products from the database
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	query := `
	SELECT
		id,
		name,
		description,
		price,
		type,
		active,
		is_single_unit,
		single_unit_price,
		created_at,
		updated_at
  FROM products
	ORDER BY name ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Type,
			&product.Active,
			&product.IsSingleUnit,
			&product.SingleUnitPrice,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

// Get retrieves a single product by ID
func (r *ProductRepository) Get(id int64) (*models.Product, error) {
	query := `
		SELECT
			id,
			name,
			description,
			price,
			type,
			active,
			is_single_unit,
			single_unit_price,
			created_at,
			updated_at
		FROM products WHERE id = ?
		`
	var product models.Product
	err := r.db.QueryRow(query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Type,
		&product.Active,
		&product.IsSingleUnit,
		&product.SingleUnitPrice,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(product *models.Product) error {
	query := `
		UPDATE products
		SET
			name = ?,
			description = ?,
			price = ?,
			type = ?,
			active = ?,
			is_single_unit = ?,
			single_unit_price = ?,
			updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.Exec(
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Type,
		product.Active,
		product.IsSingleUnit,
		product.SingleUnitPrice,
		now,
		product.ID,
	)
	if err != nil {
		return err
	}
	product.UpdatedAt = now
	return nil
}

// Delete removes a product by ID
func (r *ProductRepository) Delete(id int64) error {
	query := `DELETE FROM products WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
