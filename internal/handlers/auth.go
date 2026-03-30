package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Register maneja el registro de usuarios
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validación básica
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		RespondWithError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	if len(req.Password) < 8 {
		RespondWithError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Verificar si el usuario ya existe
	existingUser, err := h.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existingUser != nil {
		RespondWithError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Hashear contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error processing password")
		return
	}

	// Crear usuario
	user, err := h.store.CreateUser(r.Context(), req.Email, string(hashedPassword), req.FirstName, req.LastName)
	if err != nil {
		if store.IsUniqueViolation(err) {
			RespondWithError(w, http.StatusConflict, "Email already registered")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	// Generar tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error generating refresh token")
		return
	}

	// Respuesta
	response := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    h.config.JWT.ExpirationHours * 3600,
		User:         user.ToResponse(),
	}

	RespondWithJSON(w, http.StatusCreated, response)
}

// Login maneja el inicio de sesión
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Buscar usuario
	user, err := h.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if user == nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generar tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error generating refresh token")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    h.config.JWT.ExpirationHours * 3600,
		User:         user.ToResponse(),
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// RefreshToken renueva el access token
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validar refresh token
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if claims.TokenType != "refresh" {
		RespondWithError(w, http.StatusUnauthorized, "Invalid token type")
		return
	}

	// Obtener usuario
	user, err := h.store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if user == nil {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Generar nuevo access token
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		ExpiresIn:    h.config.JWT.ExpirationHours * 3600,
		User:         user.ToResponse(),
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// GetCurrentUser retorna el usuario actual
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if user == nil {
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, user.ToResponse())
}
