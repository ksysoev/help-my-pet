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
						Start:     "Text de démarrage",
						RateLimit: "Text de limite",
					},
				},
			},
			lang:     "fr",
			msgType:  StartMessage,
			expected: "Text de démarrage",
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
		{
			name: "returns global limit message for specified language",
			config: &Config{
				Languages: map[string]Messages{
					"es": {
						Error:       "Error en español",
						Start:       "Inicio en español",
						RateLimit:   "Límite en español",
						GlobalLimit: "Límite global en español",
					},
				},
			},
			lang:     "es",
			msgType:  GlobalLimitMessage,
			expected: "Límite global en español",
		},
		{
			name: "returns global limit message from English fallback",
			config: &Config{
				Languages: map[string]Messages{
					"en": {
						Error:       "English error",
						Start:       "English start",
						RateLimit:   "English rate limit",
						GlobalLimit: "English global limit",
					},
				},
			},
			lang:     "fr",
			msgType:  GlobalLimitMessage,
			expected: "English global limit",
		},
		{
			name: "returns default global limit message when not defined",
			config: &Config{
				Languages: map[string]Messages{},
			},
			lang:     "en",
			msgType:  GlobalLimitMessage,
			expected: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
		},
		{
			name: "handles fallback to English",
			config: &Config{
				Languages: map[string]Messages{
					"en": {
						Error:       "Error",
						Start:       "Start",
						RateLimit:   "Rate limit",
						GlobalLimit: "Global limit",
					},
				},
			},
			lang:     "fr",
			msgType:  GlobalLimitMessage,
			expected: "Global limit",
		},
		{
			name: "message too long",
			config: &Config{
				Languages: map[string]Messages{},
			},
			lang:     "en",
			msgType:  MessageTooLong,
			expected: "I apologize, but your message is too long for me to process. Please try to make it shorter and more concise.",
		},
		{
			name: "unsupported media in specific language",
			config: &Config{
				Languages: map[string]Messages{
					"es": {
						Error:            "Error en español",
						UnsupportedMedia: "No puedo procesar imágenes, videos, audio o documentos. Por favor, envía tu pregunta solo como texto.",
					},
				},
			},
			lang:     "es",
			msgType:  UnsupportedMediaType,
			expected: "No puedo procesar imágenes, videos, audio o documentos. Por favor, envía tu pregunta solo como texto.",
		},
		{
			name: "unsupported media fallback to English",
			config: &Config{
				Languages: map[string]Messages{
					"en": {
						Error:            "English error",
						UnsupportedMedia: "I cannot process images, videos, audio, or documents. Please send your question as text only.",
					},
				},
			},
			lang:     "fr",
			msgType:  UnsupportedMediaType,
			expected: "I cannot process images, videos, audio, or documents. Please send your question as text only.",
		},
		{
			name: "unsupported media default message",
			config: &Config{
				Languages: map[string]Messages{},
			},
			lang:     "en",
			msgType:  UnsupportedMediaType,
			expected: "Sorry, I cannot process images, videos, audio, or documents. Please send your question as text only.",
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
