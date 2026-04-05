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

	// Auto-migración al iniciar
	if err := runMigrations(context.Background(), str); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
	} else {
		log.Println("✅ Database migrations completed")
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

		// Endpoint público para migraciones (temporal)
		r.Post("/admin/migrate", h.RunMigrations)
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

// runMigrations ejecuta las migraciones automáticas al iniciar
func runMigrations(ctx context.Context, str *store.Store) error {
	migrationSQL := `
-- Tabla de historial de logins
CREATE TABLE IF NOT EXISTS login_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_type VARCHAR(50),
    device_name VARCHAR(100),
    browser VARCHAR(100),
    os VARCHAR(100),
    location VARCHAR(100),
    login_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    email_sent BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_login_history_user ON login_history(user_id);
CREATE INDEX IF NOT EXISTS idx_login_history_login_at ON login_history(login_at DESC);

-- Tabla para configuraciones de notificación
CREATE TABLE IF NOT EXISTS user_notification_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    email_on_login BOOLEAN DEFAULT TRUE,
    email_on_new_device BOOLEAN DEFAULT TRUE,
    welcome_email_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Trigger para updated_at
DROP TRIGGER IF EXISTS update_user_notification_settings_updated_at ON user_notification_settings;
CREATE TRIGGER update_user_notification_settings_updated_at 
    BEFORE UPDATE ON user_notification_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Configuraciones por defecto para usuarios existentes (solo si no existen)
INSERT INTO user_notification_settings (user_id, email_on_login, email_on_new_device, welcome_email_sent)
SELECT id, TRUE, TRUE, TRUE FROM users
ON CONFLICT (user_id) DO NOTHING;
`
	_, err := str.Pool().Exec(ctx, migrationSQL)
	return err
}
