package models

import (
	"time"
)

// ProductType represents the type of product
type ProductType string

const (
	ProductTypeRegular   ProductType = "regular"
	ProductTypeCigarette ProductType = "cigarette"
)

// Product represents a product in the system
type Product struct {
	ID              int64       `json:"id"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Price           float64     `json:"price"`
	Type            ProductType `json:"type"`
	Active          bool        `json:"active"`
	IsSingleUnit    bool        `json:"is_single_unit"`    // For cigarettes: true if single, false if packet
	SingleUnitPrice float64     `json:"single_unit_price"` // For cigarettes: true if single, false if packet
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// GetID returns the product ID
func (p *Product) GetID() int64 {
	return p.ID
}

// SetID sets the product ID
func (p *Product) SetID(id int64) {
	p.ID = id
}

// SetCreatedAt sets the created timestamp
func (p *Product) SetCreatedAt(timestamp any) {
	if t, ok := timestamp.(time.Time); ok {
		p.CreatedAt = t
	}
}

// SetUpdatedAt sets the updated timestamp
func (p *Product) SetUpdatedAt(timestamp any) {
	if t, ok := timestamp.(time.Time); ok {
		p.UpdatedAt = t
	}
}
