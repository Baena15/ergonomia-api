package models

import (
	"time"
)

// Product representa un producto de afiliados
type Product struct {
	ID              int64              `json:"id" db:"id"`
	Name            string             `json:"name" db:"name"`
	Slug            string             `json:"slug" db:"slug"`
	Description     string             `json:"description" db:"description"`
	Category        string             `json:"category" db:"category"`
	Subcategory     string             `json:"subcategory" db:"subcategory"`
	Price           float64            `json:"price" db:"price"`
	Currency        string             `json:"currency" db:"currency"`
	Rating          float64            `json:"rating" db:"rating"`
	Reviews         int                `json:"reviews" db:"reviews"`
	ImageURL        *string             `json:"image_url" db:"image_url"`
	Pros            []string            `json:"pros" db:"pros"`
	Cons            []string            `json:"cons" db:"cons"`
	IdealFor        []string            `json:"ideal_for" db:"ideal_for"`
	AffiliateLinks  map[string]string   `json:"affiliate_links" db:"affiliate_links"`
	IsActive        bool               `json:"is_active" db:"is_active"`
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" db:"updated_at"`
}

// ProductFilters filtros para búsqueda de productos
type ProductFilters struct {
	Category    string
	Subcategory string
	MinPrice    float64
	MaxPrice    float64
	MinRating   float64
	Search      string
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
}

// ProductListResponse respuesta paginada
type ProductListResponse struct {
	Products   []Product `json:"products"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

// CreateProductRequest datos para crear producto
type CreateProductRequest struct {
	Name           string            `json:"name" validate:"required,max=200"`
	Slug           string            `json:"slug" validate:"required,max=200"`
	Description    string            `json:"description" validate:"required"`
	Category       string            `json:"category" validate:"required"`
	Subcategory    string            `json:"subcategory"`
	Price          float64           `json:"price" validate:"required,gt=0"`
	Currency       string            `json:"currency" validate:"required,len=3"`
	Rating         float64           `json:"rating" validate:"gte=0,lte=5"`
	Reviews        int               `json:"reviews" validate:"gte=0"`
	ImageURL       string            `json:"image_url"`
	Pros           []string          `json:"pros"`
	Cons           []string          `json:"cons"`
	IdealFor       []string          `json:"ideal_for"`
	AffiliateLinks map[string]string `json:"affiliate_links"`
}
