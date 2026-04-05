// Package device parsea User-Agent para detectar dispositivos
package device

import (
	"regexp"
	"strings"

	"github.com/Gentleman-Programming/ergonomia-api/internal/models"
)

// Parser detecta información del dispositivo desde User-Agent
type Parser struct {
	mobilePatterns  []*regexp.Regexp
	tabletPatterns  []*regexp.Regexp
	browserPatterns map[string]*regexp.Regexp
	osPatterns      map[string]*regexp.Regexp
	devicePatterns  map[string]*regexp.Regexp
}

// NewParser crea un nuevo parser de dispositivos
func NewParser() *Parser {
	return &Parser{
		mobilePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)mobile|iphone|ipod|android.*mobile|windows phone`),
		},
		tabletPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)tablet|ipad|android(?!.*mobile)|kindle|silk`),
		},
		browserPatterns: map[string]*regexp.Regexp{
			"Chrome":  regexp.MustCompile(`(?i)chrome/(\d+\.\d+)`),
			"Firefox": regexp.MustCompile(`(?i)firefox/(\d+\.\d+)`),
			"Safari":  regexp.MustCompile(`(?i)safari/(\d+\.\d+)`),
			"Edge":    regexp.MustCompile(`(?i)edge/(\d+\.\d+)|edg/(\d+\.\d+)`),
			"Opera":   regexp.MustCompile(`(?i)opera/(\d+\.\d+)|opr/(\d+\.\d+)`),
		},
		osPatterns: map[string]*regexp.Regexp{
			"Windows":   regexp.MustCompile(`(?i)windows nt (\d+\.\d+)`),
			"macOS":     regexp.MustCompile(`(?i)mac os x (\d+[._]\d+)`),
			"iOS":       regexp.MustCompile(`(?i)os (\d+[._]\d+)`),
			"Android":   regexp.MustCompile(`(?i)android (\d+\.\d+)`),
			"Linux":     regexp.MustCompile(`(?i)linux`),
			"Chrome OS": regexp.MustCompile(`(?i)crOS`),
		},
		devicePatterns: map[string]*regexp.Regexp{
			"iPhone":         regexp.MustCompile(`(?i)iphone`),
			"iPad":           regexp.MustCompile(`(?i)ipad`),
			"Samsung Galaxy": regexp.MustCompile(`(?i)samsung|galaxy`),
			"Google Pixel":   regexp.MustCompile(`(?i)pixel`),
			"OnePlus":        regexp.MustCompile(`(?i)oneplus`),
			"Huawei":         regexp.MustCompile(`(?i)huawei`),
			"Xiaomi":         regexp.MustCompile(`(?i)xiaomi|redmi`),
			"MacBook":        regexp.MustCompile(`(?i)macintosh|macbook`),
			"Windows PC":     regexp.MustCompile(`(?i)windows`),
		},
	}
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

	info := models.DeviceInfo{
		DeviceType: "desktop",
		IsDesktop:  true,
	}

	// Detectar si es móvil
	for _, pattern := range p.mobilePatterns {
		if pattern.MatchString(userAgent) {
			info.DeviceType = "mobile"
			info.IsMobile = true
			info.IsDesktop = false
			break
		}
	}

	// Detectar si es tablet
	for _, pattern := range p.tabletPatterns {
		if pattern.MatchString(userAgent) {
			info.DeviceType = "tablet"
			info.IsTablet = true
			info.IsMobile = false
			info.IsDesktop = false
			break
		}
	}

	// Detectar navegador
	for browser, pattern := range p.browserPatterns {
		if matches := pattern.FindStringSubmatch(userAgent); matches != nil {
			info.Browser = browser
			if len(matches) > 1 && matches[1] != "" {
				info.BrowserVer = strings.Replace(matches[1], "_", ".", -1)
			} else if len(matches) > 2 && matches[2] != "" {
				info.BrowserVer = strings.Replace(matches[2], "_", ".", -1)
			}
			break
		}
	}
	if info.Browser == "" {
		info.Browser = "Unknown"
	}

	// Detectar sistema operativo
	for os, pattern := range p.osPatterns {
		if matches := pattern.FindStringSubmatch(userAgent); matches != nil {
			info.OS = os
			if len(matches) > 1 {
				info.OSVersion = strings.Replace(matches[1], "_", ".", -1)
			}
			break
		}
	}
	if info.OS == "" {
		info.OS = "Unknown"
	}

	// Detectar nombre del dispositivo
	for device, pattern := range p.devicePatterns {
		if pattern.MatchString(userAgent) {
			info.DeviceName = device
			break
		}
	}
	if info.DeviceName == "" {
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
