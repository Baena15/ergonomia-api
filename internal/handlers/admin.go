package handlers

import (
	"net/http"
)

// RunMigrations ejecuta las migraciones pendientes
// POST /api/v1/admin/migrate
func (h *Handler) RunMigrations(w http.ResponseWriter, r *http.Request) {
	// Ejecutar migración de login_history
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

	_, err := h.store.Pool().Exec(r.Context(), migrationSQL)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Migration failed: "+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":    "Migrations completed successfully",
		"migrations": "login_history, user_notification_settings",
	})
}
