// Package config maneja la configuración de la aplicación
package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
	CORS     CORSConfig
}

// DatabaseConfig configuración de PostgreSQL
type DatabaseConfig struct {
	URL string
}

// JWTConfig configuración de tokens
type JWTConfig struct {
	Secret                   string
	ExpirationHours          int
	RefreshTokenExpirationDays int
}

// ServerConfig configuración del servidor HTTP
type ServerConfig struct {
	Port string
	Env  string
}

// CORSConfig configuración de CORS
type CORSConfig struct {
	AllowedOrigins []string
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	// Cargar .env solo en desarrollo
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	return &Config{
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://ergonomia:ergonomia_dev@localhost:5433/ergonomia?sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret:                     getEnv("JWT_SECRET", "default-secret-change-in-production-min-32-chars"),
			ExpirationHours:            getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshTokenExpirationDays: getEnvAsInt("REFRESH_TOKEN_EXPIRATION_DAYS", 7),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:8080"}),
		},
	}
}

// IsProduction retorna true si estamos en producción
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// getEnv obtiene variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt obtiene variable de entorno como int
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvAsSlice obtiene variable de entorno como slice
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
