package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
	"github.com/go-chi/chi/v5"
)

// ListComparisons lista las comparaciones del usuario
func (h *Handler) ListComparisons(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

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

	comparisons, total, err := h.store.ListComparisons(r.Context(), userID, limit, offset)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error fetching comparisons")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"comparisons": comparisons,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateComparison crea una nueva comparación
func (h *Handler) CreateComparison(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	var req models.CreateComparisonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.ProductIDs) < 2 {
		RespondWithError(w, http.StatusBadRequest, "At least 2 products required")
		return
	}

	if len(req.ProductIDs) > 4 {
		RespondWithError(w, http.StatusBadRequest, "Maximum 4 products allowed")
		return
	}

	if req.Title == "" {
		req.Title = "Comparación"
	}

	comparison, err := h.store.CreateComparison(r.Context(), userID, req.Title, req.ProductIDs)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error creating comparison")
		return
	}

	RespondWithJSON(w, http.StatusCreated, comparison)
}

// GetComparison obtiene una comparación por slug (público si se comparte)
func (h *Handler) GetComparison(w http.ResponseWriter, r *http.Request) {
	slug := getSlug(r)
	if slug == "" {
		RespondWithError(w, http.StatusBadRequest, "Slug is required")
		return
	}

	detail, err := h.store.GetComparisonBySlug(r.Context(), slug)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error fetching comparison")
		return
	}
	if detail == nil {
		RespondWithError(w, http.StatusNotFound, "Comparison not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, detail)
}

// DeleteComparison elimina una comparación del usuario
func (h *Handler) DeleteComparison(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	idStr := chi.URLParam(r, "id")
	id, err := parseID(idStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.store.DeleteComparison(r.Context(), id, userID); err != nil {
		RespondWithError(w, http.StatusNotFound, "Comparison not found or unauthorized")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Comparison deleted",
	})
}
