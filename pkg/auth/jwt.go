// Package auth maneja la autenticación y autorización
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims estructura de claims JWT
type Claims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// Service maneja operaciones JWT
type Service struct {
	secret                    []byte
	accessTokenExpiration     time.Duration
	refreshTokenExpiration    time.Duration
}

// NewService crea un nuevo servicio JWT
func NewService(secret string, accessHours, refreshDays int) *Service {
	return &Service{
		secret:                 []byte(secret),
		accessTokenExpiration:  time.Duration(accessHours) * time.Hour,
		refreshTokenExpiration: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

// GenerateAccessToken genera un token de acceso
func (s *Service) GenerateAccessToken(userID int64, email string, isAdmin bool) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		IsAdmin:   isAdmin,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ergonomia-api",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken genera un token de refresco
func (s *Service) GenerateRefreshToken(userID int64) (string, error) {
	claims := Claims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ergonomia-api",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken valida y parsea un token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// HashPassword hashea una contraseña (usamos bcrypt en el handler)
// Esto es un placeholder, el hashing real está en los handlers
func HashPassword(password string) (string, error) {
	// Delegado a golang.org/x/crypto/bcrypt en los handlers
	return password, nil
}
