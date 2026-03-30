package models

import (
	"time"
)

// Comparison representa una comparación guardada por un usuario
type Comparison struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Slug      string    `json:"slug" db:"slug"`
	Title     string    `json:"title" db:"title"`
	ProductIDs []int64  `json:"product_ids" db:"product_ids"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ComparisonDetail comparación con productos completos
type ComparisonDetail struct {
	Comparison Comparison `json:"comparison"`
	Products   []Product  `json:"products"`
}

// CreateComparisonRequest datos para crear comparación
type CreateComparisonRequest struct {
	Title      string  `json:"title" validate:"required,max=200"`
	ProductIDs []int64 `json:"product_ids" validate:"required,min=2,max=4"`
}

// Favorite representa un producto guardado por un usuario
type Favorite struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ProductID int64     `json:"product_id" db:"product_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FavoriteDetail favorito con datos del producto
type FavoriteDetail struct {
	Favorite Favorite `json:"favorite"`
	Product  Product  `json:"product"`
}
