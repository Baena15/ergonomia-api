// Package validator - Validaciones de entrada

package validator

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Validator contiene las reglas de validación
type Validator struct {
	Errors map[string]string
}

// New crea un nuevo validator
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid retorna true si no hay errores
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError añade un error
func (v *Validator) AddError(field, message string) {
	v.Errors[field] = message
}

// Check añade error si la condición es false
func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

// Validaciones comunes

// NotEmpty verifica que el string no esté vacío
func NotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// MinLength verifica longitud mínima
func MinLength(s string, min int) bool {
	return len(strings.TrimSpace(s)) >= min
}

// MaxLength verifica longitud máxima
func MaxLength(s string, max int) bool {
	return len(strings.TrimSpace(s)) <= max
}

// Between verifica que la longitud esté en rango
func Between(s string, min, max int) bool {
	length := len(strings.TrimSpace(s))
	return length >= min && length <= max
}

// Matches verifica contra regex
func Matches(s string, rx *regexp.Regexp) bool {
	return rx.MatchString(s)
}

// IsEmail verifica formato de email
func IsEmail(s string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return Matches(s, emailRegex)
}

// IsURL verifica formato de URL
func IsURL(s string) bool {
	urlRegex := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return Matches(s, urlRegex)
}

// IsSlug verifica formato de slug
func IsSlug(s string) bool {
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return Matches(s, slugRegex)
}

// PasswordStrength verifica fortaleza de contraseña
func PasswordStrength(password string) (score int, errors []string) {
	var errs []string

	// Longitud mínima
	if len(password) < 8 {
		errs = append(errs, "Al menos 8 caracteres")
	} else {
		score++
	}

	// Mayúsculas
	hasUpper := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		errs = append(errs, "Al menos una mayúscula")
	} else {
		score++
	}

	// Minúsculas
	hasLower := false
	for _, char := range password {
		if unicode.IsLower(char) {
			hasLower = true
			break
		}
	}
	if !hasLower {
		errs = append(errs, "Al menos una minúscula")
	} else {
		score++
	}

	// Números
	hasNumber := false
	for _, char := range password {
		if unicode.IsDigit(char) {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		errs = append(errs, "Al menos un número")
	} else {
		score++
	}

	// Caracteres especiales
	hasSpecial := false
	for _, char := range password {
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
			break
		}
	}
	if !hasSpecial {
		errs = append(errs, "Al menos un caracter especial")
	} else {
		score++
	}

	return score, errs
}

// Sanitize limpia un string
func Sanitize(s string) string {
	// Eliminar espacios extra
	s = strings.TrimSpace(s)
	// Normalizar espacios múltiples
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return s
}

// ValidateSlug crea un slug válido desde un string
func CreateSlug(s string) string {
	// Convertir a minúsculas
	s = strings.ToLower(s)
	// Reemplazar espacios por guiones
	s = strings.ReplaceAll(s, " ", "-")
	// Eliminar caracteres no alfanuméricos (excepto guiones)
	s = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(s, "")
	// Eliminar guiones múltiples
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	// Eliminar guiones al inicio y final
	s = strings.Trim(s, "-")

	return s
}

// Common validation errors
var (
	ErrInvalidEmail    = errors.New("email inválido")
	ErrInvalidPassword = errors.New("contraseña débil")
	ErrInvalidSlug     = errors.New("slug inválido")
	ErrEmptyField      = errors.New("campo requerido")
)
