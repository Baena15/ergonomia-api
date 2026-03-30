package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// ListFavorites lista los favoritos del usuario con detalles de productos
func (h *Handler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	// Parsear paginación
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	favorites, total, err := h.store.ListFavorites(r.Context(), userID, limit, offset)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error fetching favorites")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"favorites": favorites,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// AddFavorite añade un producto a favoritos
func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	var req struct {
		ProductID int64 `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ProductID == 0 {
		RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	favorite, err := h.store.CreateFavorite(r.Context(), userID, req.ProductID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error adding favorite")
		return
	}

	RespondWithJSON(w, http.StatusCreated, favorite)
}

// RemoveFavorite elimina un producto de favoritos
func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	productIDStr := chi.URLParam(r, "productID")
	productID, err := parseID(productIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if err := h.store.DeleteFavorite(r.Context(), userID, productID); err != nil {
		RespondWithError(w, http.StatusNotFound, "Favorite not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Removed from favorites",
	})
}
