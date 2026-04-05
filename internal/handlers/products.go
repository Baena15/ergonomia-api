package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
	"github.com/Gentleman-Programming/ergonomia-api/internal/store"
	"github.com/go-chi/chi/v5"
)

// ListProducts lista productos con filtros
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Parsear query params
	filters := models.ProductFilters{
		Category:  r.URL.Query().Get("category"),
		Subcategory: r.URL.Query().Get("subcategory"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}
	if filters.Limit == 0 {
		filters.Limit = 20
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
		if p, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filters.MinPrice = p
		}
	}

	if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
		if p, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filters.MaxPrice = p
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}

	result, err := h.store.ListProducts(r.Context(), filters)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error fetching products")
		return
	}

	RespondWithJSON(w, http.StatusOK, result)
}

// GetProduct obtiene un producto por slug
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	slug := getSlug(r)
	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Slug is required")
		return
	}

	product, err := h.store.GetProductBySlug(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error fetching product")
		return
	}
	if product == nil {
		RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, product)
}

// CreateProduct crea un nuevo producto (admin)
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validación
	if req.Name == "" || req.Slug == "" || req.Description == "" {
		RespondWithError(w, http.StatusBadRequest, "Name, slug and description are required")
		return
	}

	if req.Currency == "" {
		req.Currency = "EUR"
	}

	product, err := h.store.CreateProduct(r.Context(), &req)
	if err != nil {
		if store.IsUniqueViolation(err) {
			RespondWithError(w, http.StatusConflict, "Product with this slug already exists")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Error creating product")
		return
	}

	RespondWithJSON(w, http.StatusCreated, product)
}

// UpdateProduct actualiza un producto (admin)
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := parseID(idStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validación
	if req.Name == "" || req.Slug == "" || req.Description == "" {
		RespondWithError(w, http.StatusBadRequest, "Name, slug and description are required")
		return
	}

	product, err := h.store.UpdateProduct(r.Context(), id, &req)
	if err != nil {
		if err.Error() == "product not found" {
			RespondWithError(w, http.StatusNotFound, "Product not found")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Error updating product")
		return
	}

	RespondWithJSON(w, http.StatusOK, product)
}

// DeleteProduct elimina un producto (admin - soft delete)
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := parseID(idStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.store.DeleteProduct(r.Context(), id); err != nil {
		if err.Error() == "product not found" {
			RespondWithError(w, http.StatusNotFound, "Product not found")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Error deleting product")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

// GetAdminStats retorna estadísticas para el admin dashboard
func (h *Handler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetAdminStats(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error fetching stats")
		return
	}

	RespondWithJSON(w, http.StatusOK, stats)
}
