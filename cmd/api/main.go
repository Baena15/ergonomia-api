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
				r.Post("/migrate", h.RunMigrations)
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
