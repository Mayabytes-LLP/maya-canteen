package models

import (
	"database/sql"
	"time"
)

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductModel struct {
	DB *sql.DB
}

func NewProductModel(db *sql.DB) *ProductModel {
	return &ProductModel{DB: db}
}

func (m *ProductModel) InitTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS products (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      description TEXT,
      price REAL NOT NULL,
      created_at DATETIME NOT NULL,
      updated_at DATETIME NOT NULL
    )
  `
	_, err := m.DB.Exec(query)
	return err
}

func (m *ProductModel) Create(product *Product) error {
	query := `
    INSERT INTO products (name, description, price, created_at, updated_at) 
    VALUES (?, ?, ?, ?, ?)
  `

	now := time.Now()
	result, err := m.DB.Exec(query, product.Name, product.Description, product.Price, now, now)
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

func (m *ProductModel) GetAll() ([]Product, error) {
	query := `
    SELECT * FROM products
  `

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		product := Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (m *ProductModel) Get(id int64) (*Product, error) {
	query := `SELECT * FROM products WHERE id = ?`

	row := m.DB.QueryRow(query, id)
	product := Product{}
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (m *ProductModel) Update(product *Product) error {
	query := `
    UPDATE products 
    SET name = ?, description = ?, price = ?, updated_at = ? 
    WHERE id = ?
  `

	now := time.Now()
	_, err := m.DB.Exec(query, product.Name, product.Description, product.Price, now, product.ID)
	if err != nil {
		return err
	}
	product.UpdatedAt = now
	return nil
}

func (m *ProductModel) Delete(id int64) error {
	query := `DELETE FROM products WHERE id = ?`
	_, err := m.DB.Exec(query, id)
	return err
}
