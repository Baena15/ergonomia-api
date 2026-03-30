// Package handlers contiene los HTTP handlers de la API
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Gentleman-Programming/ergonomia-api/internal/config"
	"github.com/Gentleman-Programming/ergonomia-api/internal/store"
	"github.com/Gentleman-Programming/ergonomia-api/pkg/auth"
	"github.com/go-chi/chi/v5"
)

// Handler agrupa todos los handlers
type Handler struct {
	store      *store.Store
	jwtService *auth.Service
	config     *config.Config
}

// New crea una nueva instancia de handlers
func New(store *store.Store, jwtService *auth.Service, cfg *config.Config) *Handler {
	return &Handler{
		store:      store,
		jwtService: jwtService,
		config:     cfg,
	}
}

// RespondWithJSON envía una respuesta JSON
func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// RespondWithError envía una respuesta de error
func RespondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, map[string]string{"error": message})
}

// RespondWithMessage envía un mensaje simple
func RespondWithMessage(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, map[string]string{"message": message})
}

// parseID convierte string a int64
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// getSlug obtiene el slug de la URL
func getSlug(r *http.Request) string {
	return chi.URLParam(r, "slug")
}
