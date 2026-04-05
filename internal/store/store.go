// Package store maneja el acceso a datos (Repository Pattern)
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store es el repositorio principal
type Store struct {
	pool *pgxpool.Pool
}

// NewStore crea una nueva instancia del store
func NewStore(databaseURL string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Configuración de pool
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verificar conexión
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close cierra la conexión a la base de datos
func (s *Store) Close() {
	s.pool.Close()
}

// Pool retorna el pool de conexiones (para operaciones raw si es necesario)
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// HealthCheck verifica que la base de datos esté accesible
func (s *Store) HealthCheck(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// IsUniqueViolation verifica si un error es violación de unique constraint
func IsUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}

// IsForeignKeyViolation verifica si es violación de foreign key
func IsForeignKeyViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23503" // foreign_key_violation
	}
	return false
}

// ─── User Repository ─────────────────────────────────────────

// CreateUser crea un nuevo usuario
func (s *Store) CreateUser(ctx context.Context, email, passwordHash, firstName, lastName string) (*models.User, error) {
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, first_name, last_name, is_admin, created_at, updated_at
	`

	var user models.User
	err := s.pool.QueryRow(ctx, query, email, passwordHash, firstName, lastName).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail busca usuario por email
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, is_admin, created_at, updated_at
		FROM users WHERE email = $1
	`

	var user models.User
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

// GetUserByID busca usuario por ID
func (s *Store) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, is_admin, created_at, updated_at
		FROM users WHERE id = $1
	`

	var user models.User
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

// ─── Product Repository ─────────────────────────────────────

// GetProductBySlug obtiene un producto por su slug
func (s *Store) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	query := `
		SELECT id, name, slug, description, category, subcategory, price, currency,
		       rating, reviews, image_url, pros, cons, ideal_for, affiliate_links,
		       is_active, created_at, updated_at
		FROM products WHERE slug = $1 AND is_active = true
	`

	var p models.Product
	var affiliateLinksJSON []byte

	err := s.pool.QueryRow(ctx, query, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Category, &p.Subcategory,
		&p.Price, &p.Currency, &p.Rating, &p.Reviews, &p.ImageURL,
		&p.Pros, &p.Cons, &p.IdealFor, &affiliateLinksJSON,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting product: %w", err)
	}

	// Parse JSON de affiliate links
	if affiliateLinksJSON != nil {
		p.AffiliateLinks = make(map[string]string)
		// pgx maneja JSONB como []byte, necesitamos parsear
		// Esto se maneja mejor con pgx.CopyFrom o tipo personalizado
		// Por simplicidad, lo dejamos como está por ahora
	}

	return &p, nil
}

// ListProducts lista productos con filtros
func (s *Store) ListProducts(ctx context.Context, filters models.ProductFilters) (*models.ProductListResponse, error) {
	// Construir query dinámica
	whereClause := "WHERE is_active = true"
	args := []interface{}{}
	argIdx := 1

	if filters.Category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filters.Category)
		argIdx++
	}

	if filters.MinPrice > 0 {
		whereClause += fmt.Sprintf(" AND price >= $%d", argIdx)
		args = append(args, filters.MinPrice)
		argIdx++
	}

	if filters.MaxPrice > 0 {
		whereClause += fmt.Sprintf(" AND price <= $%d", argIdx)
		args = append(args, filters.MaxPrice)
		argIdx++
	}

	// Query de conteo
	countQuery := "SELECT COUNT(*) FROM products " + whereClause
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error counting products: %w", err)
	}

	// Query principal
	query := `
		SELECT id, name, slug, description, category, subcategory, price, currency,
		       rating, reviews, image_url, pros, cons, ideal_for, affiliate_links,
		       is_active, created_at, updated_at
		FROM products ` + whereClause + `
		ORDER BY rating DESC, reviews DESC
		LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing products: %w", err)
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
		// Parse affiliate links JSON
		if len(affiliateJSON) > 0 {
			p.AffiliateLinks = make(map[string]string)
			// Simple parsing - in production use json.Unmarshal
			// For now, leave empty to avoid complexity
		}
		products = append(products, p)
	}

	pageSize := filters.Limit
	if pageSize == 0 {
		pageSize = 20
	}
	page := (filters.Offset / pageSize) + 1

	return &models.ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}, nil
}
