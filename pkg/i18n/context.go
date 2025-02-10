package i18n

import (
	"context"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type localizerContextKey struct{}

var defaultPrinter = message.NewPrinter(language.MustParse("en-GB"))

// SetLocale injects a language-specific printer into the provided context for localization.
// It retrieves the printer for the specified language from the Localizer instance.
// If the language is not supported, it falls back to the default language printer.
// Returns a new context containing the localized printer.
func SetLocale(ctx context.Context, l10n *Localizer, lang string) context.Context {
	return context.WithValue(ctx, localizerContextKey{}, l10n.GetPrinter(lang))
}

// GetLocale retrieves the localized message printer associated with the provided context.
// It uses the localizer stored in the context for localized messages or defaults to the English printer if unavailable.
// Accepts ctx, the request context that may contain localization details.
// Returns the *message.Printer for localized message formatting.
func GetLocale(ctx context.Context) *message.Printer {
	if ctx == nil {
		return defaultPrinter
	}

	l10n, ok := ctx.Value(localizerContextKey{}).(*message.Printer)
	if !ok {
		return defaultPrinter
	}

	return l10n
}
