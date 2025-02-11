package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestLocalizer_GetPrinter(t *testing.T) {
	tests := []struct {
		name          string
		lang          string
		setupPrinters func() map[string]*message.Printer
		expected      *message.Printer
	}{
		{
			name: "existing lang",
			lang: "en",
			setupPrinters: func() map[string]*message.Printer {
				return map[string]*message.Printer{
					"en": message.NewPrinter(language.English),
				}
			},
			expected: message.NewPrinter(language.English),
		},
		{
			name: "non-existent lang falls back to default",
			lang: "fr",
			setupPrinters: func() map[string]*message.Printer {
				return map[string]*message.Printer{
					"en": message.NewPrinter(language.English),
				}
			},
			expected: message.NewPrinter(language.English),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printers := tt.setupPrinters()
			localizer := &Localizer{printers: printers}

			// Act
			result := localizer.GetPrinter(tt.lang)

			// Assert
			assert.NotNil(t, result)
		})
	}
}

func TestNewLocalizer(t *testing.T) {
	tests := []struct {
		name              string
		supportedLangTags map[string]language.Tag
		verifyResult      func(t *testing.T, localizer *Localizer)
	}{
		{
			name: "create localizer with supported languages",
			supportedLangTags: map[string]language.Tag{
				"en": language.English,
				"es": language.Spanish,
			},
			verifyResult: func(t *testing.T, localizer *Localizer) {
				assert.NotNil(t, localizer.printers["en"])
				assert.NotNil(t, localizer.printers["es"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock supportedLanguages
			supportedLanguages = tt.supportedLangTags

			// Act
			localizer := NewLocalizer()

			// Assert
			require.NotNil(t, localizer)
			assert.NotNil(t, localizer.printers)
			tt.verifyResult(t, localizer)
		})
	}
}
