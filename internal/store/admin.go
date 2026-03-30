// ─── Admin & Stats Repository ─────────────────────────────────

package store

import (
	"context"
	"fmt"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// AdminStats estadísticas para el dashboard
type AdminStats struct {
	TotalUsers       int64 `json:"total_users"`
	TotalProducts    int64 `json:"total_products"`
	ActiveProducts   int64 `json:"active_products"`
	TotalFavorites   int64 `json:"total_favorites"`
	TotalComparisons int64 `json:"total_comparisons"`
	UsersToday       int64 `json:"users_today"`
}

// GetAdminStats obtiene estadísticas del sistema
func (s *Store) GetAdminStats(ctx context.Context) (*AdminStats, error) {
	stats := &AdminStats{}

	queries := []struct {
		sql   string
		target *int64
	}{
		{"SELECT COUNT(*) FROM users", &stats.TotalUsers},
		{"SELECT COUNT(*) FROM products", &stats.TotalProducts},
		{"SELECT COUNT(*) FROM products WHERE is_active = true", &stats.ActiveProducts},
		{"SELECT COUNT(*) FROM favorites", &stats.TotalFavorites},
		{"SELECT COUNT(*) FROM comparisons", &stats.TotalComparisons},
		{"SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE", &stats.UsersToday},
	}

	for _, q := range queries {
		if err := s.pool.QueryRow(ctx, q.sql).Scan(q.target); err != nil {
			return nil, fmt.Errorf("error getting stats: %w", err)
		}
	}

	return stats, nil
}

// CreateProduct crea un nuevo producto
func (s *Store) CreateProduct(ctx context.Context, p *models.CreateProductRequest) (*models.Product, error) {
	query := `
		INSERT INTO products (name, slug, description, category, subcategory, price, currency,
		                     rating, reviews, image_url, pros, cons, ideal_for, affiliate_links)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, name, slug, description, category, subcategory, price, currency,
		         rating, reviews, image_url, pros, cons, ideal_for, affiliate_links,
		         is_active, created_at, updated_at
	`

	var product models.Product
	var affiliateJSON []byte

	err := s.pool.QueryRow(ctx, query,
		p.Name, p.Slug, p.Description, p.Category, p.Subcategory,
		p.Price, p.Currency, p.Rating, p.Reviews, p.ImageURL,
		p.Pros, p.Cons, p.IdealFor, p.AffiliateLinks,
	).Scan(
		&product.ID, &product.Name, &product.Slug, &product.Description,
		&product.Category, &product.Subcategory, &product.Price, &product.Currency,
		&product.Rating, &product.Reviews, &product.ImageURL,
		&product.Pros, &product.Cons, &product.IdealFor, &affiliateJSON,
		&product.IsActive, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating product: %w", err)
	}

	product.AffiliateLinks = p.AffiliateLinks
	return &product, nil
}

// UpdateProduct actualiza un producto
func (s *Store) UpdateProduct(ctx context.Context, id int64, p *models.CreateProductRequest) (*models.Product, error) {
	query := `
		UPDATE products
		SET name = $1, slug = $2, description = $3, category = $4, subcategory = $5,
		    price = $6, currency = $7, rating = $8, reviews = $9, image_url = $10,
		    pros = $11, cons = $12, ideal_for = $13, affiliate_links = $14,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $15
		RETURNING id, name, slug, description, category, subcategory, price, currency,
		         rating, reviews, image_url, pros, cons, ideal_for, affiliate_links,
		         is_active, created_at, updated_at
	`

	var product models.Product
	var affiliateJSON []byte

	err := s.pool.QueryRow(ctx, query,
		p.Name, p.Slug, p.Description, p.Category, p.Subcategory,
		p.Price, p.Currency, p.Rating, p.Reviews, p.ImageURL,
		p.Pros, p.Cons, p.IdealFor, p.AffiliateLinks, id,
	).Scan(
		&product.ID, &product.Name, &product.Slug, &product.Description,
		&product.Category, &product.Subcategory, &product.Price, &product.Currency,
		&product.Rating, &product.Reviews, &product.ImageURL,
		&product.Pros, &product.Cons, &product.IdealFor, &affiliateJSON,
		&product.IsActive, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating product: %w", err)
	}

	product.AffiliateLinks = p.AffiliateLinks
	return &product, nil
}

// DeleteProduct elimina (soft delete) o desactiva un producto
func (s *Store) DeleteProduct(ctx context.Context, id int64) error {
	// Soft delete: marcar como inactivo
	query := `UPDATE products SET is_active = false WHERE id = $1`
	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// HardDeleteProduct elimina permanentemente un producto
func (s *Store) HardDeleteProduct(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error hard deleting product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// GetProductByID obtiene un producto por ID
func (s *Store) GetProductByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
		SELECT id, name, slug, description, category, subcategory, price, currency,
		       rating, reviews, image_url, pros, cons, ideal_for, affiliate_links,
		       is_active, created_at, updated_at
		FROM products WHERE id = $1
	`

	var p models.Product
	var affiliateJSON []byte

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Category, &p.Subcategory,
		&p.Price, &p.Currency, &p.Rating, &p.Reviews, &p.ImageURL,
		&p.Pros, &p.Cons, &p.IdealFor, &affiliateJSON,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	p.AffiliateLinks = parseAffiliateLinks(affiliateJSON)
	return &p, nil
}

// SetUserAdmin cambia el estado de admin de un usuario
func (s *Store) SetUserAdmin(ctx context.Context, userID int64, isAdmin bool) error {
	query := `UPDATE users SET is_admin = $1 WHERE id = $2`
	result, err := s.pool.Exec(ctx, query, isAdmin, userID)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
