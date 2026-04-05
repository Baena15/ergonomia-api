// Package email maneja el envío de emails
package email

import (
	"bytes"
	"html/template"
	"log"
	"time"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// Service interfaz para enviar emails
type Service interface {
	SendWelcome(email, firstName string) error
	SendLoginNotification(data models.LoginNotificationEmail) error
	IsConfigured() bool
}

// Config configuración del servicio de email
type Config struct {
	Provider   string // "resend", "smtp", "console" (para desarrollo)
	APIKey     string
	FromEmail  string
	FromName   string
	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
}

// NewService crea un servicio de email según la configuración
func NewService(cfg Config) Service {
	switch cfg.Provider {
	case "resend":
		return NewResendService(cfg)
	case "smtp":
		return NewSMTPService(cfg)
	case "console", "":
		return NewConsoleService(cfg)
	default:
		log.Printf("⚠️  Unknown email provider '%s', using console", cfg.Provider)
		return NewConsoleService(cfg)
	}
}

// ─── Templates ───────────────────────────────────────────────

const welcomeTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>¡Bienvenido a Ergonomía Pro!</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px; text-align: center; border-radius: 8px 8px 0 0; }
        .header h1 { color: white; margin: 0; font-size: 28px; }
        .content { background: #f9fafb; padding: 40px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .features { background: white; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .feature { margin: 10px 0; }
        .emoji { font-size: 20px; margin-right: 10px; }
        .footer { text-align: center; margin-top: 30px; color: #6b7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>💺 Ergonomía Pro</h1>
    </div>
    <div class="content">
        <h2>¡Hola {{.UserName}}! 👋</h2>
        <p>Bienvenido a <strong>Ergonomía Pro</strong>. Tu cuenta ha sido creada exitosamente y ahora puedes disfrutar de todas nuestras funciones exclusivas.</p>
        
        <div class="features">
            <h3>✨ Lo que puedes hacer ahora:</h3>
            <div class="feature"><span class="emoji">⭐</span> Guardar tus productos favoritos</div>
            <div class="feature"><span class="emoji">⚖️</span> Crear comparativas ilimitadas</div>
            <div class="feature"><span class="emoji">💰</span> Recibir alertas de precios</div>
            <div class="feature"><span class="emoji">📧</span> Acceder al newsletter exclusivo</div>
        </div>

        <center>
            <a href="https://baena15.github.io/Web-Ergonomia/" class="button">Ir a mi cuenta</a>
        </center>

        <p>Si tienes alguna pregunta, no dudes en contactarnos respondiendo a este email.</p>
        
        <p>¡Gracias por unirte!<br>
        <strong>El equipo de Ergonomía Pro</strong></p>
    </div>
    <div class="footer">
        <p>© 2026 Ergonomía Pro. Todos los derechos reservados.</p>
        <p>Este email fue enviado automáticamente. Por favor, no respondas a esta dirección.</p>
    </div>
</body>
</html>`

const loginNotificationTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Nuevo inicio de sesión - Ergonomía Pro</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .header h1 { color: white; margin: 0; font-size: 24px; }
        .content { background: #f9fafb; padding: 40px; border-radius: 0 0 8px 8px; }
        .alert { background: {{if .IsNewDevice}}#fef3c7{{else}}#dbeafe{{end}}; border-left: 4px solid {{if .IsNewDevice}}#f59e0b{{else}}#3b82f6{{end}}; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .device-info { background: white; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .device-row { display: flex; padding: 10px 0; border-bottom: 1px solid #e5e7eb; }
        .device-row:last-child { border-bottom: none; }
        .device-label { width: 100px; color: #6b7280; font-weight: 600; }
        .device-value { flex: 1; color: #111827; }
        .footer { text-align: center; margin-top: 30px; color: #6b7280; font-size: 14px; }
        .warning { background: #fee2e2; border-left: 4px solid #ef4444; padding: 15px; margin: 20px 0; border-radius: 4px; color: #991b1b; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.DeviceIcon}} Nuevo inicio de sesión</h1>
    </div>
    <div class="content">
        <h2>Hola {{.UserName}},</h2>
        
        <div class="alert">
            {{if .IsNewDevice}}
            <strong>⚠️ Nuevo dispositivo detectado</strong><br>
            Hemos detectado un inicio de sesión desde un dispositivo nuevo o poco frecuente.
            {{else}}
            <strong>ℹ️ Inicio de sesión detectado</strong><br>
            Te notificamos que se ha iniciado sesión en tu cuenta.
            {{end}}
        </div>

        <h3>📱 Detalles del acceso:</h3>
        <div class="device-info">
            <div class="device-row">
                <div class="device-label">Dispositivo:</div>
                <div class="device-value">{{.DeviceInfo.DeviceName}}</div>
            </div>
            <div class="device-row">
                <div class="device-label">Sistema:</div>
                <div class="device-value">{{.DeviceInfo.OS}} {{.DeviceInfo.OSVersion}}</div>
            </div>
            <div class="device-row">
                <div class="device-label">Navegador:</div>
                <div class="device-value">{{.DeviceInfo.Browser}} {{.DeviceInfo.BrowserVer}}</div>
            </div>
            <div class="device-row">
                <div class="device-label">Ubicación:</div>
                <div class="device-value">{{.Location}}</div>
            </div>
            <div class="device-row">
                <div class="device-label">IP:</div>
                <div class="device-value">{{.IPAddress}}</div>
            </div>
            <div class="device-row">
                <div class="device-label">Fecha/Hora:</div>
                <div class="device-value">{{.LoginTime}}</div>
            </div>
        </div>

        {{if .IsNewDevice}}
        <div class="warning">
            <strong>🔒 ¿No fuiste tú?</strong><br>
            Si no reconoces esta actividad, te recomendamos cambiar tu contraseña inmediatamente desde tu cuenta.
        </div>
        {{end}}

        <p>Si fuiste tú, puedes ignorar este email con tranquilidad.</p>
        
        <p>Saludos,<br>
        <strong>El equipo de Ergonomía Pro</strong></p>
    </div>
    <div class="footer">
        <p>© 2026 Ergonomía Pro. Todos los derechos reservados.</p>
        <p>Este email fue enviado automáticamente por seguridad.</p>
    </div>
</body>
</html>`

// renderTemplate renderiza un template con los datos proporcionados
func renderTemplate(tmpl string, data interface{}) (string, error) {
	t := template.Must(template.New("email").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// formatTime formatea la hora para español
func formatTime(t time.Time) string {
	return t.Format("02/01/2006 15:04:05 MST")
}

// ConsoleService implementación para desarrollo (imprime en consola)
type ConsoleService struct {
	cfg Config
}

// NewConsoleService crea un servicio de consola
func NewConsoleService(cfg Config) *ConsoleService {
	return &ConsoleService{cfg: cfg}
}

// IsConfigured siempre retorna true para consola
func (s *ConsoleService) IsConfigured() bool {
	return true
}

// SendWelcome envía email de bienvenida (a consola)
func (s *ConsoleService) SendWelcome(email, firstName string) error {
	html, err := renderTemplate(welcomeTemplate, map[string]interface{}{
		"UserName": firstName,
	})
	if err != nil {
		return err
	}

	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("📧 EMAIL DE BIENVENIDA (Console Mode)")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("Para: %s\n", email)
	log.Printf("Asunto: ¡Bienvenido a Ergonomía Pro!\n")
	log.Printf("HTML:\n%s\n", html)
	log.Println("═══════════════════════════════════════════════════════════")
	return nil
}

// SendLoginNotification envía notificación de login (a consola)
func (s *ConsoleService) SendLoginNotification(data models.LoginNotificationEmail) error {
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
		return err
	}

	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("📧 EMAIL DE NOTIFICACIÓN DE LOGIN (Console Mode)")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("Para: %s\n", data.Email)
	log.Printf("Asunto: Nuevo inicio de sesión detectado\n")
	log.Printf("HTML:\n%s\n", html)
	log.Println("═══════════════════════════════════════════════════════════")
	return nil
}

func getDeviceIconForEmail(info models.DeviceInfo) string {
	if info.IsMobile {
		return "📱"
	}
	if info.IsTablet {
		return "📲"
	}
	return "💻"
}
