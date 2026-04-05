// Package store maneja el acceso a datos
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// CreateLoginHistory guarda un nuevo registro de login
func (s *Store) CreateLoginHistory(ctx context.Context, userID int64, ipAddress, userAgent string, deviceInfo models.DeviceInfo) (*models.LoginHistory, error) {
	query := `
		INSERT INTO login_history (user_id, ip_address, user_agent, device_type, device_name, browser, os, location, login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, ip_address, user_agent, device_type, device_name, browser, os, location, login_at, email_sent
	`

	history := &models.LoginHistory{
		UserID:     userID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		DeviceType: deviceInfo.DeviceType,
		DeviceName: deviceInfo.DeviceName,
		Browser:    deviceInfo.Browser,
		OS:         deviceInfo.OS,
		LoginAt:    time.Now(),
	}

	err := s.pool.QueryRow(ctx, query,
		userID, ipAddress, userAgent,
		deviceInfo.DeviceType, deviceInfo.DeviceName,
		deviceInfo.Browser, deviceInfo.OS,
		"", // location - podemos añadir geolocalización después
		history.LoginAt,
	).Scan(
		&history.ID, &history.UserID, &history.IPAddress,
		&history.UserAgent, &history.DeviceType, &history.DeviceName,
		&history.Browser, &history.OS, &history.Location,
		&history.LoginAt, &history.EmailSent,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating login history: %w", err)
	}

	return history, nil
}

// GetUserLoginHistory obtiene el historial de logins de un usuario
func (s *Store) GetUserLoginHistory(ctx context.Context, userID int64, limit int) ([]models.LoginHistory, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, user_id, ip_address, user_agent, device_type, device_name, browser, os, location, login_at, email_sent
		FROM login_history
		WHERE user_id = $1
		ORDER BY login_at DESC
		LIMIT $2
	`

	rows, err := s.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying login history: %w", err)
	}
	defer rows.Close()

	var history []models.LoginHistory
	for rows.Next() {
		var h models.LoginHistory
		err := rows.Scan(
			&h.ID, &h.UserID, &h.IPAddress,
			&h.UserAgent, &h.DeviceType, &h.DeviceName,
			&h.Browser, &h.OS, &h.Location,
			&h.LoginAt, &h.EmailSent,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning login history: %w", err)
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// IsNewDeviceForUser verifica si es un dispositivo nuevo para el usuario
func (s *Store) IsNewDeviceForUser(ctx context.Context, userID int64, deviceInfo models.DeviceInfo) bool {
	query := `
		SELECT COUNT(*)
		FROM login_history
		WHERE user_id = $1
		AND device_name = $2
		AND os = $3
		AND login_at > NOW() - INTERVAL '30 days'
	`

	var count int
	err := s.pool.QueryRow(ctx, query, userID, deviceInfo.DeviceName, deviceInfo.OS).Scan(&count)
	if err != nil {
		return true // Si hay error, asumimos que es nuevo por seguridad
	}

	return count == 0
}

// MarkLoginEmailSent marca que ya se envió el email de notificación
func (s *Store) MarkLoginEmailSent(ctx context.Context, historyID int64) error {
	query := `
		UPDATE login_history
		SET email_sent = TRUE
		WHERE id = $1
	`

	_, err := s.pool.Exec(ctx, query, historyID)
	if err != nil {
		return fmt.Errorf("error marking email as sent: %w", err)
	}
	return nil
}

// ─── User Notification Settings ─────────────────────────────

// CreateNotificationSettings crea configuración de notificaciones para un usuario
func (s *Store) CreateNotificationSettings(ctx context.Context, userID int64) (*models.UserNotificationSettings, error) {
	query := `
		INSERT INTO user_notification_settings (user_id, email_on_login, email_on_new_device)
		VALUES ($1, TRUE, TRUE)
		RETURNING id, user_id, email_on_login, email_on_new_device, welcome_email_sent, created_at, updated_at
	`

	var settings models.UserNotificationSettings
	err := s.pool.QueryRow(ctx, query, userID).Scan(
		&settings.ID, &settings.UserID,
		&settings.EmailOnLogin, &settings.EmailOnNewDevice,
		&settings.WelcomeEmailSent,
		&settings.CreatedAt, &settings.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating notification settings: %w", err)
	}

	return &settings, nil
}

// GetNotificationSettings obtiene las configuraciones de notificación de un usuario
func (s *Store) GetNotificationSettings(ctx context.Context, userID int64) (*models.UserNotificationSettings, error) {
	query := `
		SELECT id, user_id, email_on_login, email_on_new_device, welcome_email_sent, created_at, updated_at
		FROM user_notification_settings
		WHERE user_id = $1
	`

	var settings models.UserNotificationSettings
	err := s.pool.QueryRow(ctx, query, userID).Scan(
		&settings.ID, &settings.UserID,
		&settings.EmailOnLogin, &settings.EmailOnNewDevice,
		&settings.WelcomeEmailSent,
		&settings.CreatedAt, &settings.UpdatedAt,
	)
	if err != nil {
		// Si no existe, crear configuración por defecto
		return s.CreateNotificationSettings(ctx, userID)
	}

	return &settings, nil
}

// UpdateNotificationSettings actualiza las configuraciones
func (s *Store) UpdateNotificationSettings(ctx context.Context, userID int64, emailOnLogin, emailOnNewDevice bool) error {
	query := `
		UPDATE user_notification_settings
		SET email_on_login = $2, email_on_new_device = $3
		WHERE user_id = $1
	`

	_, err := s.pool.Exec(ctx, query, userID, emailOnLogin, emailOnNewDevice)
	if err != nil {
		return fmt.Errorf("error updating notification settings: %w", err)
	}
	return nil
}

// MarkWelcomeEmailSent marca que ya se envió el email de bienvenida
func (s *Store) MarkWelcomeEmailSent(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_notification_settings
		SET welcome_email_sent = TRUE
		WHERE user_id = $1
	`

	_, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error marking welcome email as sent: %w", err)
	}
	return nil
}
