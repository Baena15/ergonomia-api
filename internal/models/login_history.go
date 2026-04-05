// Package models define las entidades del dominio
package models

import (
	"time"
)

// LoginHistory representa un registro de inicio de sesión
type LoginHistory struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	DeviceType  string    `json:"device_type" db:"device_type"`
	DeviceName  string    `json:"device_name" db:"device_name"`
	Browser     string    `json:"browser" db:"browser"`
	OS          string    `json:"os" db:"os"`
	Location    string    `json:"location" db:"location"`
	LoginAt     time.Time `json:"login_at" db:"login_at"`
	EmailSent   bool      `json:"email_sent" db:"email_sent"`
}

// DeviceInfo contiene información parseada del dispositivo
type DeviceInfo struct {
	DeviceType string // desktop, mobile, tablet
	DeviceName string // iPhone, Samsung Galaxy, etc.
	Browser    string // Chrome, Safari, Firefox
	BrowserVer string // 120.0
	OS         string // Windows, macOS, iOS, Android
	OSVersion  string // 10, 14.2
	IsMobile   bool
	IsTablet   bool
	IsDesktop  bool
}

// UserNotificationSettings configuración de notificaciones del usuario
type UserNotificationSettings struct {
	ID                int64     `json:"id" db:"id"`
	UserID            int64     `json:"user_id" db:"user_id"`
	EmailOnLogin      bool      `json:"email_on_login" db:"email_on_login"`
	EmailOnNewDevice  bool      `json:"email_on_new_device" db:"email_on_new_device"`
	WelcomeEmailSent  bool      `json:"welcome_email_sent" db:"welcome_email_sent"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// LoginNotificationEmail datos para email de notificación
type LoginNotificationEmail struct {
	UserName    string
	Email       string
	DeviceInfo  DeviceInfo
	IPAddress   string
	Location    string
	LoginTime   time.Time
	IsNewDevice bool
}

// WelcomeEmail datos para email de bienvenida
type WelcomeEmail struct {
	UserName string
	Email    string
}
