package handlers

import (
	"net/http"
)

// PublicMigrate crea las tablas necesarias (endpoint temporal público)
func (h *Handler) PublicMigrate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Crear tabla login_history
	_, err := h.store.Pool().Exec(ctx, `
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
		)
	`)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create login_history: "+err.Error())
		return
	}

	// Crear índices
	h.store.Pool().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_login_history_user ON login_history(user_id)`)
	h.store.Pool().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_login_history_login_at ON login_history(login_at DESC)`)

	// Crear tabla user_notification_settings
	_, err = h.store.Pool().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_notification_settings (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			email_on_login BOOLEAN DEFAULT TRUE,
			email_on_new_device BOOLEAN DEFAULT TRUE,
			welcome_email_sent BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id)
		)
	`)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create notification_settings: "+err.Error())
		return
	}

	// Insertar configuraciones por defecto
	h.store.Pool().Exec(ctx, `
		INSERT INTO user_notification_settings (user_id, email_on_login, email_on_new_device, welcome_email_sent)
		SELECT id, TRUE, TRUE, TRUE FROM users
		ON CONFLICT (user_id) DO NOTHING
	`)

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Tables created successfully",
	})
}
