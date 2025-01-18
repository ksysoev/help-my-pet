package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetMessage(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		lang     string
		msgType  Message
		expected string
	}{
		{
			name: "handles nil Languages map",
			config: &Config{
				Languages: nil,
			},
			lang:     "en",
			msgType:  ErrorMessage,
			expected: "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "returns message for specified language",
			config: &Config{
				Languages: map[string]Messages{
					"es": {
						Error:     "Error en español",
						Start:     "Inicio en español",
						RateLimit: "Límite en español",
					},
				},
			},
			lang:     "es",
			msgType:  ErrorMessage,
			expected: "Error en español",
		},
		{
			name: "falls back to English when language not found",
			config: &Config{
				Languages: map[string]Messages{
					"en": {
						Error:     "English error",
						Start:     "English start",
						RateLimit: "English rate limit",
					},
				},
			},
			lang:     "fr",
			msgType:  ErrorMessage,
			expected: "English error",
		},
		{
			name: "uses default message when no config available",
			config: &Config{
				Languages: map[string]Messages{},
			},
			lang:     "en",
			msgType:  ErrorMessage,
			expected: "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "handles empty language by using English",
			config: &Config{
				Languages: map[string]Messages{
					"en": {
						Error:     "English error",
						Start:     "English start",
						RateLimit: "English rate limit",
					},
				},
			},
			lang:     "",
			msgType:  ErrorMessage,
			expected: "English error",
		},
		{
			name: "returns start message for specified language",
			config: &Config{
				Languages: map[string]Messages{
					"de": {
						Error:     "Fehler",
						Start:     "Willkommen",
						RateLimit: "Limit erreicht",
					},
				},
			},
			lang:     "de",
			msgType:  StartMessage,
			expected: "Willkommen",
		},
		{
			name: "returns rate limit message for specified language",
			config: &Config{
				Languages: map[string]Messages{
					"fr": {
						Error:     "Erreur",
						Start:     "Bienvenue",
						RateLimit: "Limite atteinte",
					},
				},
			},
			lang:     "fr",
			msgType:  RateLimitMessage,
			expected: "Limite atteinte",
		},
		{
			name: "returns default message for unknown message type",
			config: &Config{
				Languages: map[string]Messages{
					"en": {
						Error:     "English error",
						Start:     "English start",
						RateLimit: "English rate limit",
					},
				},
			},
			lang:     "en",
			msgType:  "unknown",
			expected: "An error occurred.",
		},
		{
			name: "returns all message types in primary language",
			config: &Config{
				Languages: map[string]Messages{
					"fr": {
						Error:     "Erreur message",
						Start:     "Message de démarrage",
						RateLimit: "Message de limite",
					},
				},
			},
			lang:     "fr",
			msgType:  StartMessage,
			expected: "Message de démarrage",
		},
		{
			name: "returns all message types in English fallback",
			config: &Config{
				Languages: map[string]Messages{
					"en": {
						Error:     "Error message",
						Start:     "Start message",
						RateLimit: "Rate limit message",
					},
				},
			},
			lang:     "unknown",
			msgType:  RateLimitMessage,
			expected: "Rate limit message",
		},
		{
			name: "returns all default messages when no config",
			config: &Config{
				Languages: map[string]Messages{},
			},
			lang:     "unknown",
			msgType:  StartMessage,
			expected: "Welcome to Help My Pet Bot! How can I help you today?",
		},
		{
			name: "returns error message from default fallback",
			config: &Config{
				Languages: map[string]Messages{},
			},
			lang:     "unknown",
			msgType:  ErrorMessage,
			expected: "Sorry, I encountered an error while processing your request. Please try again later.",
		},
		{
			name: "returns rate limit message from default fallback",
			config: &Config{
				Languages: map[string]Messages{},
			},
			lang:     "unknown",
			msgType:  RateLimitMessage,
			expected: "You have reached the maximum number of requests per hour. Please try again later.",
		},
		{
			name: "prefers primary language over English when both exist",
			config: &Config{
				Languages: map[string]Messages{
					"es": {
						Error:     "Error en español",
						Start:     "Inicio en español",
						RateLimit: "Límite en español",
					},
					"en": {
						Error:     "English error",
						Start:     "English start",
						RateLimit: "English rate limit",
					},
				},
			},
			lang:     "es",
			msgType:  ErrorMessage,
			expected: "Error en español",
		},
		{
			name:     "handles nil config",
			config:   nil,
			lang:     "en",
			msgType:  ErrorMessage,
			expected: "Sorry, I encountered an error while processing your request. Please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				tt.config = &Config{}
			}
			result := tt.config.GetMessage(tt.lang, tt.msgType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
