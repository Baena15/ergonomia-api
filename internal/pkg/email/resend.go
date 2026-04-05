// Package email maneja el envío de emails
package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// ResendService implementación usando Resend API
type ResendService struct {
	cfg    Config
	client *http.Client
}

// NewResendService crea un servicio de email usando Resend
func NewResendService(cfg Config) *ResendService {
	return &ResendService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsConfigured verifica si el servicio está configurado
func (s *ResendService) IsConfigured() bool {
	return s.cfg.APIKey != "" && s.cfg.FromEmail != ""
}

// resendRequest estructura para la API de Resend
type resendRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// resendResponse estructura de respuesta de Resend
type resendResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// sendEmail envía un email usando la API de Resend
func (s *ResendService) sendEmail(to, subject, html string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("resend not configured")
	}

	from := s.cfg.FromEmail
	if s.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromEmail)
	}

	payload := resendRequest{
		From:    from,
		To:      to,
		Subject: subject,
		HTML:    html,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp struct {
			StatusCode int    `json:"statusCode"`
			Message    string `json:"message"`
			Name       string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("resend API error (status %d)", resp.StatusCode)
		}
		return fmt.Errorf("resend API error: %s", errResp.Message)
	}

	var result resendResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}

// SendWelcome envía email de bienvenida
func (s *ResendService) SendWelcome(email, firstName string) error {
	html, err := renderTemplate(welcomeTemplate, map[string]interface{}{
		"UserName": firstName,
	})
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}

	return s.sendEmail(email, "¡Bienvenido a Ergonomía Pro! 💺", html)
}

// SendLoginNotification envía notificación de login
func (s *ResendService) SendLoginNotification(data models.LoginNotificationEmail) error {
	subject := "Nuevo inicio de sesión detectado"
	if data.IsNewDevice {
		subject = "🔒 Nuevo dispositivo detectado en tu cuenta"
	}

	html, err := renderTemplate(loginNotificationTemplate, map[string]interface{}{
		"UserName":    data.UserName,
		"DeviceIcon":  getDeviceIconForEmail(data.DeviceInfo),
		"DeviceInfo":  data.DeviceInfo,
		"Location":    data.Location,
		"IPAddress":   data.IPAddress,
		"LoginTime":   formatTime(data.LoginTime),
		"IsNewDevice": data.IsNewDevice,
	})
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}

	return s.sendEmail(data.Email, subject, html)
}
