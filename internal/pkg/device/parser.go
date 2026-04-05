// Package device parsea User-Agent para detectar dispositivos
package device

import (
	"strings"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// Parser detecta información del dispositivo desde User-Agent
type Parser struct{}

// NewParser crea un nuevo parser de dispositivos
func NewParser() *Parser {
	return &Parser{}
}

// Parse analiza el User-Agent y devuelve información del dispositivo
func (p *Parser) Parse(userAgent string) models.DeviceInfo {
	if userAgent == "" {
		return models.DeviceInfo{
			DeviceType: "unknown",
			DeviceName: "Unknown Device",
			Browser:    "Unknown",
			OS:         "Unknown",
		}
	}

	ua := strings.ToLower(userAgent)
	info := models.DeviceInfo{
		DeviceType: "desktop",
		IsDesktop:  true,
	}

	// Detectar tipo de dispositivo
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "iphone") || 
	   strings.Contains(ua, "ipod") || strings.Contains(ua, "android") && strings.Contains(ua, "mobile") {
		info.DeviceType = "mobile"
		info.IsMobile = true
		info.IsDesktop = false
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") || 
	          (strings.Contains(ua, "android") && !strings.Contains(ua, "mobile")) {
		info.DeviceType = "tablet"
		info.IsTablet = true
		info.IsDesktop = false
	}

	// Detectar navegador
	switch {
	case strings.Contains(ua, "chrome"):
		info.Browser = "Chrome"
		info.BrowserVer = extractVersion(ua, "chrome/")
	case strings.Contains(ua, "firefox"):
		info.Browser = "Firefox"
		info.BrowserVer = extractVersion(ua, "firefox/")
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		info.Browser = "Safari"
		info.BrowserVer = extractVersion(ua, "version/")
	case strings.Contains(ua, "edge") || strings.Contains(ua, "edg/"):
		info.Browser = "Edge"
		info.BrowserVer = extractVersion(ua, "edge/")
		if info.BrowserVer == "" {
			info.BrowserVer = extractVersion(ua, "edg/")
		}
	case strings.Contains(ua, "opera") || strings.Contains(ua, "opr/"):
		info.Browser = "Opera"
	default:
		info.Browser = "Unknown"
	}

	// Detectar sistema operativo
	switch {
	case strings.Contains(ua, "windows"):
		info.OS = "Windows"
		info.OSVersion = extractVersion(ua, "windows nt ")
	case strings.Contains(ua, "mac os") || strings.Contains(ua, "macos"):
		info.OS = "macOS"
		info.OSVersion = extractVersion(ua, "mac os x ")
		info.OSVersion = strings.Replace(info.OSVersion, "_", ".", -1)
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		info.OS = "iOS"
		info.OSVersion = extractVersion(ua, "os ")
		info.OSVersion = strings.Replace(info.OSVersion, "_", ".", -1)
	case strings.Contains(ua, "android"):
		info.OS = "Android"
		info.OSVersion = extractVersion(ua, "android ")
	case strings.Contains(ua, "linux"):
		info.OS = "Linux"
	default:
		info.OS = "Unknown"
	}

	// Detectar nombre del dispositivo
	switch {
	case strings.Contains(ua, "iphone"):
		info.DeviceName = "iPhone"
	case strings.Contains(ua, "ipad"):
		info.DeviceName = "iPad"
	case strings.Contains(ua, "samsung") || strings.Contains(ua, "galaxy"):
		info.DeviceName = "Samsung Galaxy"
	case strings.Contains(ua, "pixel"):
		info.DeviceName = "Google Pixel"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "macbook"):
		info.DeviceName = "MacBook"
	case strings.Contains(ua, "windows"):
		info.DeviceName = "Windows PC"
	default:
		if info.IsMobile {
			info.DeviceName = "Mobile Device"
		} else if info.IsTablet {
			info.DeviceName = "Tablet"
		} else {
			info.DeviceName = "Desktop Computer"
		}
	}

	return info
}

// extractVersion extrae la versión de una cadena user-agent
func extractVersion(ua, prefix string) string {
	idx := strings.Index(ua, prefix)
	if idx == -1 {
		return ""
	}
	
	start := idx + len(prefix)
	if start >= len(ua) {
		return ""
	}
	
	// Encontrar el final de la versión (espacio, punto y coma, paréntesis, etc.)
	end := start
	for end < len(ua) && isVersionChar(ua[end]) {
		end++
	}
	
	if end > start {
		return ua[start:end]
	}
	return ""
}

// isVersionChar verifica si un caracter es válido para una versión
func isVersionChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == '.' || c == '_'
}

// FormatDeviceInfo formatea la información del dispositivo para mostrar
func FormatDeviceInfo(info models.DeviceInfo) string {
	parts := []string{}

	if info.DeviceName != "" && info.DeviceName != "Desktop Computer" {
		parts = append(parts, info.DeviceName)
	}

	if info.OS != "" && info.OS != "Unknown" {
		osStr := info.OS
		if info.OSVersion != "" {
			osStr += " " + info.OSVersion
		}
		parts = append(parts, osStr)
	}

	if info.Browser != "" && info.Browser != "Unknown" {
		browserStr := info.Browser
		if info.BrowserVer != "" {
			browserStr += " " + info.BrowserVer
		}
		parts = append(parts, browserStr)
	}

	if len(parts) == 0 {
		return "Dispositivo desconocido"
	}

	return strings.Join(parts, " • ")
}

// GetDeviceIcon devuelve un emoji según el tipo de dispositivo
func GetDeviceIcon(info models.DeviceInfo) string {
	switch {
	case info.IsMobile:
		return "📱"
	case info.IsTablet:
		return "📲"
	case strings.Contains(strings.ToLower(info.DeviceName), "mac"):
		return "🖥️"
	default:
		return "💻"
	}
}
