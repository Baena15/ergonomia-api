// ergonomia-api - Backend HTMX + Go para web de afiliados
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Gentleman-Programming/ergonomia-api/internal/config"
	"github.com/Gentleman-Programming/ergonomia-api/internal/handlers"
	"github.com/Gentleman-Programming/ergonomia-api/internal/middleware"
	"github.com/Gentleman-Programming/ergonomia-api/internal/store"
	"github.com/Gentleman-Programming/ergonomia-api/pkg/auth"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	chiCors "github.com/go-chi/cors"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Log de inicio
	log.Printf("🔧 Environment: %s", cfg.Server.Env)
	log.Printf("📊 Database URL: %s", maskConnectionString(cfg.Database.URL))

	// Inicializar conexión a base de datos
	str, err := store.NewStore(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer str.Close()

	log.Println("✅ Connected to database")

	// Ejecutar migraciones automáticamente
	if err := runMigrations(str); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
	}

	// Inicializar servicios
	jwtService := auth.NewService(
		cfg.JWT.Secret,
		cfg.JWT.ExpirationHours,
		cfg.JWT.RefreshTokenExpirationDays,
	)

	// Inicializar handlers
	h := handlers.New(str, jwtService, cfg)

	// Configurar router
	r := chi.NewRouter()

	// Middleware global
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// CORS
	r.Use(chiCors.Handler(chiCors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := str.HealthCheck(ctx); err != nil {
			handlers.RespondWithError(w, http.StatusServiceUnavailable, "Database unavailable")
			return
		}

		handlers.RespondWithJSON(w, http.StatusOK, map[string]string{
			"status":    "ok",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (públicas)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/refresh", h.RefreshToken)
		})

		// Products (públicas)
		r.Route("/products", func(r chi.Router) {
			r.Get("/", h.ListProducts)
			r.Get("/{slug}", h.GetProduct)
		})

		// Rutas protegidas
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtService))

			// User
			r.Get("/me", h.GetCurrentUser)

			// Favorites
			r.Route("/favorites", func(r chi.Router) {
				r.Get("/", h.ListFavorites)
				r.Post("/", h.AddFavorite)
				r.Delete("/{productID}", h.RemoveFavorite)
			})

			// Comparisons
			r.Route("/comparisons", func(r chi.Router) {
				r.Get("/", h.ListComparisons)
				r.Post("/", h.CreateComparison)
				r.Get("/{slug}", h.GetComparison)
				r.Delete("/{id}", h.DeleteComparison)
			})
		})

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtService))
			r.Use(middleware.RequireAdmin)

			r.Route("/admin", func(r chi.Router) {
				r.Get("/stats", h.GetAdminStats)
				r.Route("/products", func(r chi.Router) {
					r.Post("/", h.CreateProduct)
					r.Put("/{id}", h.UpdateProduct)
					r.Delete("/{id}", h.DeleteProduct)
				})
			})
		})
	})

	// Static files (templates HTMX)
	fileServer := http.FileServer(http.Dir("./web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Rutas HTMX (server-side rendering)
	r.Route("/hx", func(r chi.Router) {
		// Productos - listados parciales
		r.Get("/products", h.HXListProducts)
		r.Get("/products/{slug}", h.HXProductDetail)

		// Comparador
		r.Get("/compare", h.HXCompareForm)
		r.Post("/compare", h.HXDoCompare)
	})

	// SPA fallback (para rutas del frontend)
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/index.html")
	})

	// Configurar servidor
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Server starting on http://0.0.0.0:%s", port)
		log.Printf("📖 API Documentation: http://0.0.0.0:%s (when frontend ready)", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Esperar señal de cierre
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️  Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// maskConnectionString oculta la contraseña en logs
func maskConnectionString(connStr string) string {
	// Reemplazar password en postgres://user:pass@host...
	if len(connStr) > 20 {
		// Buscar :// y @
		if idx := strings.Index(connStr, "://"); idx != -1 {
			if atIdx := strings.Index(connStr[idx+3:], "@"); atIdx != -1 {
				return connStr[:idx+3] + "****:****@" + connStr[idx+3+atIdx+1:]
			}
		}
	}
	return connStr
}

// runMigrations ejecuta migraciones SQL básicas
func runMigrations(str *store.Store) error {
	log.Println("🔄 Running migrations...")

	ctx := context.Background()
	pool := str.Pool()

	// Crear extensión UUID primero
	_, _ = pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// Crear tablas una por una
	statements := []string{
		// Tabla users
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		// Tabla products
		`CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			slug VARCHAR(200) UNIQUE NOT NULL,
			description TEXT NOT NULL,
			category VARCHAR(50) NOT NULL,
			subcategory VARCHAR(50),
			price DECIMAL(10, 2) NOT NULL,
			currency VARCHAR(3) DEFAULT 'EUR',
			rating DECIMAL(2, 1) CHECK (rating >= 0 AND rating <= 5),
			reviews INTEGER DEFAULT 0,
			image_url VARCHAR(500),
			pros TEXT[],
			cons TEXT[],
			ideal_for TEXT[],
			affiliate_links JSONB,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_category ON products(category)`,
		`CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_products_active ON products(is_active)`,

		// Tabla favorites
		`CREATE TABLE IF NOT EXISTS favorites (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, product_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_favorites_user ON favorites(user_id)`,

		// Tabla comparisons
		`CREATE TABLE IF NOT EXISTS comparisons (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			slug VARCHAR(100) UNIQUE NOT NULL,
			title VARCHAR(200) NOT NULL,
			product_ids INTEGER[],
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_comparations_user ON comparisons(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comparations_slug ON comparisons(slug)`,
	}

	// Ejecutar cada statement
	for i, stmt := range statements {
		_, err := pool.Exec(ctx, stmt)
		if err != nil {
			// Ignorar errores de "already exists"
			if !strings.Contains(err.Error(), "already exists") {
				log.Printf("  ⚠️  Migration %d: %v", i+1, err)
			}
		}
	}

	// Insertar datos de ejemplo solo si no hay productos
	var count int
	row := pool.QueryRow(ctx, "SELECT COUNT(*) FROM products")
	if err := row.Scan(&count); err == nil && count == 0 {
		log.Println("🌱 Seeding initial data...")
		// Insertar productos de ejemplo
		_, _ = pool.Exec(ctx, `INSERT INTO products (name, slug, description, category, subcategory, price, currency, rating, reviews, pros, cons, ideal_for, affiliate_links) VALUES
			('Reposapiés Ergonómico Ajustable HUANUO', 'reposapies-huanuo-ajustable', 'Reposapiés ajustable en altura e inclinación con textura masaje. Mejora la circulación y reduce la presión en las piernas.', 'general', 'reposapies', 29.99, 'EUR', 4.5, 2847, 
			 ARRAY['Ajustable en altura (10-17cm)', 'Superficie texturizada masaje', 'Ángulo de inclinación ajustable', 'Base antideslizante'],
			 ARRAY['Plástico algo rígido al principio', 'No tiene función de calor'],
			 ARRAY['oficina', 'trabajo-casa', 'piernas-cansadas'],
			 '{"amazon": "https://www.amazon.es/s?k=HUANUO+reposapies+ajustable&tag=baena15-21"}')`)
		
		_, _ = pool.Exec(ctx, `INSERT INTO products (name, slug, description, category, subcategory, price, currency, rating, reviews, pros, cons, ideal_for, affiliate_links) VALUES
			('SIHOO M57 Silla Ergonómica de Oficina', 'silla-sihoo-m57', 'Silla ergonómica con soporte lumbar dinámico, reposabrazos 3D ajustables y malla transpirable. Ideal para largas jornadas de trabajo.', 'espalda', 'sillas', 189.99, 'EUR', 4.7, 15234,
			 ARRAY['Soporte lumbar ajustable dinámicamente', 'Reposabrazos 3D', 'Malla transpirable de alta calidad', 'Reposacabezas ajustable', 'Excelente relación calidad-precio'],
			 ARRAY['Montaje puede llevar 30-45 minutos', 'Para usuarios hasta 100kg'],
			 ARRAY['programadores', 'oficina', 'dolor-espalda', 'largas-jornadas'],
			 '{"amazon": "https://www.amazon.es/s?k=SIHOO+M57+silla+ergonomica&tag=baena15-21"}')`)
		
		_, _ = pool.Exec(ctx, `INSERT INTO products (name, slug, description, category, subcategory, price, currency, rating, reviews, pros, cons, ideal_for, affiliate_links) VALUES
			('Cojín Lumbar Ergonómico Dreamer Car', 'cojin-lumbar-dreamer', 'Cojín lumbar de espuma viscoelástica con funda lavable. Diseño ortopédico que se adapta a la curva natural de la espalda.', 'espalda', 'cojines-lumbares', 24.95, 'EUR', 4.4, 8932,
			 ARRAY['Espuma viscoelástica de alta densidad', 'Diseño ergonómico probado', 'Funda transpirable y lavable', 'Correas ajustables', 'Portátil'],
			 ARRAY['Puede ser demasiado firme al principio', 'Tamaño estándar, no XL'],
			 ARRAY['dolor-espalda', 'oficina', 'coche', 'silla-rigida'],
			 '{"amazon": "https://www.amazon.es/s?k=cojin+lumbar+ergonomico&tag=baena15-21"}')`)
		
		_, _ = pool.Exec(ctx, `INSERT INTO products (name, slug, description, category, subcategory, price, currency, rating, reviews, pros, cons, ideal_for, affiliate_links) VALUES
			('Logitech ERGO K860 Teclado Ergonómico Split', 'teclado-logitech-ergo-k860', 'Teclado inalámbrico ergonómico con diseño split curvo y reposamanos integrado. Reduce la tensión en muñecas y antebrazos.', 'programadores', 'teclados', 119.99, 'EUR', 4.6, 3421,
			 ARRAY['Diseño split curvo natural', 'Reposamanos acolchado integrado', 'Teclas de perfil bajo silenciosas', 'Conectividad multi-dispositivo', 'Batería de 2 años'],
			 ARRAY['Precio elevado', 'Curva de aprendizaje de 1-2 semanas', 'No es mecánico'],
			 ARRAY['programadores', 'tunnel-carpiano', 'escritura-larga'],
			 '{"amazon": "https://www.amazon.es/s?k=Logitech+ERGO+K860&tag=baena15-21"}')`)
	}

	log.Println("✅ Migrations completed")
	return nil
}
