// Package middleware - Rate Limiting

package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implementa rate limiting simple en memoria
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Limpiar entradas antiguas periódicamente
	go rl.cleanup()

	return rl
}

// Allow verifica si una petición está permitida
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Filtrar solo peticiones en la ventana de tiempo
	var recent []time.Time
	for _, t := range rl.requests[key] {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}

	// Verificar límite
	if len(recent) >= rl.limit {
		rl.requests[key] = recent
		return false
	}

	// Añadir petición actual
	recent = append(recent, now)
	rl.requests[key] = recent

	return true
}

// cleanup elimina entradas antiguas cada cierto tiempo
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		windowStart := time.Now().Add(-rl.window)

		for key, times := range rl.requests {
			var recent []time.Time
			for _, t := range times {
				if t.After(windowStart) {
					recent = append(recent, t)
				}
			}

			if len(recent) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = recent
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit middleware
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(limit, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Usar IP como key, o userID si está autenticado
			key := r.RemoteAddr
			if userID := r.Context().Value("userID"); userID != nil {
				key = userID.(string)
			}

			if !limiter.Allow(key) {
				http.Error(w, `{"error": "Rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByIP rate limiting solo por IP (para rutas públicas)
func RateLimitByIP(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(limit, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(r.RemoteAddr) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", window.String())
				http.Error(w, `{"error": "Too many requests"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
