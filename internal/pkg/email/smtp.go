// Package email maneja el envío de emails
package email

import (
	"fmt"
	"net/smtp"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// SMTPService implementación usando SMTP
type SMTPService struct {
	cfg Config
}

// NewSMTPService crea un servicio de email SMTP
func NewSMTPService(cfg Config) *SMTPService {
	return &SMTPService{cfg: cfg}
}

// IsConfigured verifica si el servicio está configurado
func (s *SMTPService) IsConfigured() bool {
	return s.cfg.SMTPHost != "" && s.cfg.SMTPUser != "" && s.cfg.SMTPPass != ""
}

// SendWelcome envía email de bienvenida
func (s *SMTPService) SendWelcome(email, firstName string) error {
	html, err := renderTemplate(welcomeTemplate, map[string]interface{}{
		"UserName": firstName,
	})
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}

	return s.sendEmail(email, "¡Bienvenido a Ergonomía Pro! 💺", html)
}

// SendLoginNotification envía notificación de login
func (s *SMTPService) SendLoginNotification(data models.LoginNotificationEmail) error {
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

// sendEmail envía un email usando SMTP
func (s *SMTPService) sendEmail(to, subject, html string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("smtp not configured")
	}

	from := s.cfg.FromEmail
	if s.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromEmail)
	}

	// Headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""

	// Construir mensaje
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + html

	// Autenticación
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)

	// Enviar
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	if s.cfg.SMTPPort == 0 {
		addr = fmt.Sprintf("%s:587", s.cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{to}, []byte(msg))
}
