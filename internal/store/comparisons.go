// ─── Comparisons Repository ───────────────────────────────────

package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
	"github.com/jackc/pgx/v5"
)

// generateSlug genera un slug único para comparaciones
func generateSlug() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateComparison crea una nueva comparación
func (s *Store) CreateComparison(ctx context.Context, userID int64, title string, productIDs []int64) (*models.Comparison, error) {
	slug := generateSlug()

	query := `
		INSERT INTO comparisons (user_id, slug, title, product_ids)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, slug, title, product_ids, created_at, updated_at
	`

	var c models.Comparison
	err := s.pool.QueryRow(ctx, query, userID, slug, title, productIDs).Scan(
		&c.ID, &c.UserID, &c.Slug, &c.Title, &c.ProductIDs, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating comparison: %w", err)
	}

	return &c, nil
}

// GetComparisonBySlug obtiene una comparación por su slug público
func (s *Store) GetComparisonBySlug(ctx context.Context, slug string) (*models.ComparisonDetail, error) {
	// Obtener la comparación
	query := `
		SELECT id, user_id, slug, title, product_ids, created_at, updated_at
		FROM comparisons WHERE slug = $1
	`

	var c models.Comparison
	err := s.pool.QueryRow(ctx, query, slug).Scan(
		&c.ID, &c.UserID, &c.Slug, &c.Title, &c.ProductIDs, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting comparison: %w", err)
	}

	// Obtener los productos
	products, err := s.GetProductsByIDs(ctx, c.ProductIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting comparison products: %w", err)
	}

	return &models.ComparisonDetail{
		Comparison: c,
		Products:   products,
	}, nil
}

// GetComparisonByID obtiene una comparación por ID (para el dueño)
func (s *Store) GetComparisonByID(ctx context.Context, id int64) (*models.Comparison, error) {
	query := `
		SELECT id, user_id, slug, title, product_ids, created_at, updated_at
		FROM comparisons WHERE id = $1
	`

	var c models.Comparison
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.UserID, &c.Slug, &c.Title, &c.ProductIDs, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting comparison: %w", err)
	}

	return &c, nil
}

// ListComparisons lista las comparaciones de un usuario
func (s *Store) ListComparisons(ctx context.Context, userID int64, limit, offset int) ([]models.Comparison, int, error) {
	// Contar
	var total int
	countQuery := `SELECT COUNT(*) FROM comparisons WHERE user_id = $1`
	if err := s.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting comparisons: %w", err)
	}

	// Listar
	query := `
		SELECT id, user_id, slug, title, product_ids, created_at, updated_at
		FROM comparisons WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	if limit == 0 {
		limit = 20
	}

	rows, err := s.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing comparisons: %w", err)
	}
	defer rows.Close()

	var comparisons []models.Comparison
	for rows.Next() {
		var c models.Comparison
		err := rows.Scan(
			&c.ID, &c.UserID, &c.Slug, &c.Title, &c.ProductIDs,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning comparison: %w", err)
		}
		comparisons = append(comparisons, c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating comparisons: %w", err)
	}

	return comparisons, total, nil
}

// DeleteComparison elimina una comparación
func (s *Store) DeleteComparison(ctx context.Context, id, userID int64) error {
	query := `DELETE FROM comparisons WHERE id = $1 AND user_id = $2`
	result, err := s.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("error deleting comparison: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comparison not found or unauthorized")
	}

	return nil
}

// GetProductsByIDs obtiene múltiples productos por sus IDs
func (s *Store) GetProductsByIDs(ctx context.Context, ids []int64) ([]models.Product, error) {
	if len(ids) == 0 {
		return []models.Product{}, nil
	}

	query := `
		SELECT id, name, slug, description, category, subcategory, price, currency,
		       rating, reviews, image_url, pros, cons, ideal_for, affiliate_links,
		       is_active, created_at, updated_at
		FROM products
		WHERE id = ANY($1) AND is_active = true
		ORDER BY array_position($1, id)
	`

	rows, err := s.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("error getting products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var affiliateJSON []byte

		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Category, &p.Subcategory,
			&p.Price, &p.Currency, &p.Rating, &p.Reviews, &p.ImageURL,
			&p.Pros, &p.Cons, &p.IdealFor, &affiliateJSON,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}

		if affiliateJSON != nil {
			p.AffiliateLinks = parseAffiliateLinks(affiliateJSON)
		}

		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// UpdateComparison actualiza una comparación
func (s *Store) UpdateComparison(ctx context.Context, id, userID int64, title string, productIDs []int64) (*models.Comparison, error) {
	query := `
		UPDATE comparisons
		SET title = $1, product_ids = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, slug, title, product_ids, created_at, updated_at
	`

	var c models.Comparison
	err := s.pool.QueryRow(ctx, query, title, productIDs, id, userID).Scan(
		&c.ID, &c.UserID, &c.Slug, &c.Title, &c.ProductIDs, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("comparison not found or unauthorized")
	}
	if err != nil {
		return nil, fmt.Errorf("error updating comparison: %w", err)
	}

	return &c, nil
}
