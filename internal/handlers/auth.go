package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
	"github.com/Gentleman-Programming/ergonomia-api/internal/store"
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

	// Crear configuración de notificaciones por defecto
	_, _ = h.store.CreateNotificationSettings(r.Context(), user.ID)

	// Enviar email de bienvenida (async - no bloquear la respuesta)
	go func() {
		if err := h.emailService.SendWelcome(user.Email, user.FirstName); err != nil {
			// Log error pero no fallar el registro
			// En producción usar un logger proper
			println("Error sending welcome email:", err.Error())
		} else {
			// Marcar como enviado
			_ = h.store.MarkWelcomeEmailSent(context.Background(), user.ID)
		}
	}()

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

	// Detectar información del dispositivo
	userAgent := r.Header.Get("User-Agent")
	ipAddress := getClientIP(r)
	deviceInfo := h.deviceParser.Parse(userAgent)

	// Guardar en historial de logins
	loginHistory, err := h.store.CreateLoginHistory(r.Context(), user.ID, ipAddress, userAgent, deviceInfo)
	if err != nil {
		// Log error pero no fallar el login
		println("Error saving login history:", err.Error())
	}

	// Verificar si es un dispositivo nuevo
	isNewDevice := h.store.IsNewDeviceForUser(r.Context(), user.ID, deviceInfo)

	// Obtener configuraciones de notificación
	notifSettings, _ := h.store.GetNotificationSettings(r.Context(), user.ID)

	// Enviar email de notificación si está habilitado y es nuevo dispositivo o login general
	shouldSendEmail := notifSettings != nil && 
		((notifSettings.EmailOnNewDevice && isNewDevice) || notifSettings.EmailOnLogin)

	if shouldSendEmail && loginHistory != nil {
		go func() {
			emailData := models.LoginNotificationEmail{
				UserName:    user.FirstName,
				Email:       user.Email,
				DeviceInfo:  deviceInfo,
				IPAddress:   ipAddress,
				Location:    "Ubicación desconocida", // Podemos añadir geolocalización después
				LoginTime:   loginHistory.LoginAt,
				IsNewDevice: isNewDevice,
			}

			if err := h.emailService.SendLoginNotification(emailData); err != nil {
				println("Error sending login notification:", err.Error())
			} else {
				// Marcar como enviado
				_ = h.store.MarkLoginEmailSent(context.Background(), loginHistory.ID)
			}
		}()
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

// getClientIP obtiene la IP real del cliente (considerando proxies)
func getClientIP(r *http.Request) string {
	// Verificar headers de proxy comunes
	headers := []string{"X-Forwarded-For", "X-Real-Ip"}
	for _, header := range headers {
		if ip := r.Header.Get(header); ip != "" {
			// X-Forwarded-For puede contener múltiples IPs, tomar la primera
			if idx := len(ip); idx > 0 {
				if comma := len(ip); comma > 0 {
					parts := splitIP(ip)
					if len(parts) > 0 {
						return parts[0]
					}
				}
			}
			return ip
		}
	}

	// Si no hay headers de proxy, usar RemoteAddr
	ip := r.RemoteAddr
	// Quitar el puerto si existe
	if idx := len(ip); idx > 0 {
		for i := len(ip) - 1; i >= 0; i-- {
			if ip[i] == ':' {
				return ip[:i]
			}
		}
	}
	return ip
}

// splitIP separa IPs por coma (para X-Forwarded-For)
func splitIP(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, trimSpace(s[start:i]))
			start = i + 1
		}
	}
	result = append(result, trimSpace(s[start:]))
	return result
}

// trimSpace elimina espacios al inicio y final
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
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
