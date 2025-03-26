package models

import (
	"time"
)

// TransactionProduct represents a product included in a transaction
type TransactionProduct struct {
	ID            int64     `json:"id"`
	TransactionID int64     `json:"transaction_id"`
	ProductID     int64     `json:"product_id"`
	ProductName   string    `json:"product_name"`
	Quantity      int       `json:"quantity"`
	UnitPrice     float64   `json:"unit_price"`
	IsSingleUnit  bool      `json:"is_single_unit"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TransactionProductDetail represents transaction product with additional product details
type TransactionProductDetail struct {
	ID            int64     `json:"id"`
	TransactionID int64     `json:"transaction_id"`
	ProductID     int64     `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductType   string    `json:"product_type"`
	Quantity      int       `json:"quantity"`
	UnitPrice     float64   `json:"unit_price"`
	TotalPrice    float64   `json:"total_price"`
	IsSingleUnit  bool      `json:"is_single_unit"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProductSalesSummary represents summary statistics for product sales
type ProductSalesSummary struct {
	ProductID      int64   `json:"product_id"`
	ProductName    string  `json:"product_name"`
	ProductType    string  `json:"product_type"`
	TotalQuantity  int     `json:"total_quantity"`
	TotalSales     float64 `json:"total_sales"`
	SingleUnitSold int     `json:"single_unit_sold"`
	FullUnitSold   int     `json:"full_unit_sold"`
}
