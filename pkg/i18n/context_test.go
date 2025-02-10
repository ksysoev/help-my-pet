package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestSetLocale_ValidLanguage(t *testing.T) {
	localizer := &Localizer{
		printers: map[string]*message.Printer{
			"en": message.NewPrinter(language.English),
		},
	}
	ctx := context.Background()
	newCtx := SetLocale(ctx, localizer, "en")

	printer := newCtx.Value(localizerContextKey{}).(*message.Printer)
	assert.NotNil(t, printer)
	assert.Equal(t, message.NewPrinter(language.English), printer)
}

func TestSetLocale_InvalidLanguage(t *testing.T) {
	localizer := &Localizer{
		printers: map[string]*message.Printer{
			"en": message.NewPrinter(language.English),
		},
	}
	ctx := context.Background()
	newCtx := SetLocale(ctx, localizer, "fr")

	printer := newCtx.Value(localizerContextKey{}).(*message.Printer)
	assert.NotNil(t, printer)
	assert.Equal(t, message.NewPrinter(language.English), printer)
}

func TestGetLocale_ContextWithPrinter(t *testing.T) {
	ctx := context.WithValue(context.Background(), localizerContextKey{}, message.NewPrinter(language.English))

	printer := GetLocale(ctx)
	assert.NotNil(t, printer)
	assert.Equal(t, message.NewPrinter(language.English), printer)
}

func TestGetLocale_ContextWithoutPrinter(t *testing.T) {
	ctx := context.Background()

	printer := GetLocale(ctx)
	assert.NotNil(t, printer)
	assert.Equal(t, defaultPrinter, printer)
}

func TestGetLocale_NilContext(t *testing.T) {
	//nolint:staticcheck // Ignore SA1019: nil context is expected
	printer := GetLocale(nil)
	assert.NotNil(t, printer)
	assert.Equal(t, defaultPrinter, printer)
}
