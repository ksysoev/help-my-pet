package i18n

import (
	"context"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type localizerContextKey struct{}

var defaultPrinter = message.NewPrinter(language.MustParse("en-GB"))

func SetLocale(ctx context.Context, l10n *Localizer, lang string) context.Context {
	ctx = context.WithValue(ctx, localizerContextKey{}, l10n.GetPrinter(lang))
	return ctx
}

// GetLocale retrieves the localized message printer from the provided context.
// It defaults to a preconfigured printer if the context is nil or does not contain a valid localizer.
// Returns a *message.Printer for message localization.
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
