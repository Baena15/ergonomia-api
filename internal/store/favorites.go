// ─── Favorites Repository ─────────────────────────────────────

package store

import (
	"context"
	"fmt"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
	"github.com/jackc/pgx/v5"
)

// CreateFavorite añade un producto a favoritos del usuario
func (s *Store) CreateFavorite(ctx context.Context, userID, productID int64) (*models.Favorite, error) {
	query := `
		INSERT INTO favorites (user_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
		RETURNING id, user_id, product_id, created_at
	`

	var f models.Favorite
	err := s.pool.QueryRow(ctx, query, userID, productID).Scan(
		&f.ID, &f.UserID, &f.ProductID, &f.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Ya existía, obtener el existente
			return s.GetFavorite(ctx, userID, productID)
		}
		return nil, fmt.Errorf("error creating favorite: %w", err)
	}

	return &f, nil
}

// GetFavorite obtiene un favorito específico
func (s *Store) GetFavorite(ctx context.Context, userID, productID int64) (*models.Favorite, error) {
	query := `
		SELECT id, user_id, product_id, created_at
		FROM favorites WHERE user_id = $1 AND product_id = $2
	`

	var f models.Favorite
	err := s.pool.QueryRow(ctx, query, userID, productID).Scan(
		&f.ID, &f.UserID, &f.ProductID, &f.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting favorite: %w", err)
	}

	return &f, nil
}

// ListFavorites obtiene todos los favoritos de un usuario con datos del producto
func (s *Store) ListFavorites(ctx context.Context, userID int64, limit, offset int) ([]models.FavoriteDetail, int, error) {
	// Contar total
	var total int
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	if err := s.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting favorites: %w", err)
	}

	// Obtener favoritos con datos del producto
	query := `
		SELECT 
			f.id, f.user_id, f.product_id, f.created_at,
			p.id, p.name, p.slug, p.description, p.category, p.subcategory,
			p.price, p.currency, p.rating, p.reviews, p.image_url,
			p.pros, p.cons, p.ideal_for, p.affiliate_links,
			p.is_active, p.created_at, p.updated_at
		FROM favorites f
		JOIN products p ON f.product_id = p.id
		WHERE f.user_id = $1 AND p.is_active = true
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	if limit == 0 {
		limit = 20
	}

	rows, err := s.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing favorites: %w", err)
	}
	defer rows.Close()

	var favorites []models.FavoriteDetail
	for rows.Next() {
		var fd models.FavoriteDetail
		var affiliateJSON []byte

		err := rows.Scan(
			&fd.Favorite.ID, &fd.Favorite.UserID, &fd.Favorite.ProductID, &fd.Favorite.CreatedAt,
			&fd.Product.ID, &fd.Product.Name, &fd.Product.Slug, &fd.Product.Description,
			&fd.Product.Category, &fd.Product.Subcategory, &fd.Product.Price, &fd.Product.Currency,
			&fd.Product.Rating, &fd.Product.Reviews, &fd.Product.ImageURL,
			&fd.Product.Pros, &fd.Product.Cons, &fd.Product.IdealFor, &affiliateJSON,
			&fd.Product.IsActive, &fd.Product.CreatedAt, &fd.Product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning favorite: %w", err)
		}

		// Parse affiliate links
		if affiliateJSON != nil {
			fd.Product.AffiliateLinks = parseAffiliateLinks(affiliateJSON)
		}

		favorites = append(favorites, fd)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, total, nil
}

// DeleteFavorite elimina un favorito
func (s *Store) DeleteFavorite(ctx context.Context, userID, productID int64) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND product_id = $2`
	result, err := s.pool.Exec(ctx, query, userID, productID)
	if err != nil {
		return fmt.Errorf("error deleting favorite: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("favorite not found")
	}

	return nil
}

// IsFavorite verifica si un producto es favorito del usuario
func (s *Store) IsFavorite(ctx context.Context, userID, productID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND product_id = $2)`
	var exists bool
	err := s.pool.QueryRow(ctx, query, userID, productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking favorite: %w", err)
	}
	return exists, nil
}

// Helper para parsear JSON de affiliate links
func parseAffiliateLinks(data []byte) map[string]string {
	// Implementación simple - en producción usar json.Unmarshal
	links := make(map[string]string)
	// TODO: Implementar parsing real de JSON
	return links
}
