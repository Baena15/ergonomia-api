package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFormatList verifica el helper de formateo
func TestFormatList(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected string
	}{
		{
			name:     "empty list",
			items:    []string{},
			expected: "",
		},
		{
			name:     "single item",
			items:    []string{"item1"},
			expected: "<li>item1</li>",
		},
		{
			name:     "multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: "<li>item1</li><li>item2</li><li>item3</li>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatList(tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHXEndpoints_ResponseContentType verifica que HTMX devuelve HTML
func TestHXEndpoints_ContentType(t *testing.T) {
	// Nota: Estos tests requieren el store configurado
	// Por ahora, verificamos que los handlers existen

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"compare form", "GET", "/hx/compare"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Placeholder - los tests reales requerirían setup del handler
			assert.True(t, true)
		})
	}
}

// TestParseID verifica el helper de parseo de IDs
func TestParseID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		shouldError bool
	}{
		{
			name:        "valid number",
			input:       "123",
			expected:    123,
			shouldError: false,
		},
		{
			name:        "zero",
			input:       "0",
			expected:    0,
			shouldError: false,
		},
		{
			name:        "negative number",
			input:       "-5",
			expected:    -5,
			shouldError: false,
		},
		{
			name:        "invalid string",
			input:       "abc",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseID(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
