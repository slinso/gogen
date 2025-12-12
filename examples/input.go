package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system.
// It contains basic profile information and authentication details.
type User struct {
	// ID is the unique identifier for the user
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email" validate:"required,email"`
	Name      string     `json:"name"`
	Age       int        `json:"age,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Tags      []string   `json:"tags"`
}

// Role represents a user role in the system.
type Role string

// Timestamps contains common timestamp fields.
type Timestamps struct {
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// Address represents a physical address.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
	ZipCode string `json:"zipCode,omitempty"`
}

// OrderStatus represents the status of an order.
type OrderStatus int

// Order represents a customer order.
type Order struct {
	ID       uuid.UUID   `json:"id"`
	UserID   uuid.UUID   `json:"userId"`
	Status   OrderStatus `json:"status"`
	Items    []OrderItem `json:"items"`
	Shipping *Address    `json:"shipping,omitempty"`
	Total    float64     `json:"total"`
	Timestamps
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
}

// ProductCategory is an alias for string.
type ProductCategory = string

// Product represents a product in the catalog.
type Product struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Price       float64         `json:"price"`
	Category    ProductCategory `json:"category"`
	InStock     bool            `json:"inStock"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}
