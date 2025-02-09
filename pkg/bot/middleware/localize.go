package middleware

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type localizerContextKey struct{}

var defaultPrinter = message.NewPrinter(language.MustParse("en-GB"))

// WithLocalization wraps a Handler with language-specific localization support for incoming messages.
// It retrieves the user's language code from the incoming message and attaches a localizer instance to the context.
// Returns Middleware that ensures the context contains a localized message printer for the message's language.
func WithLocalization() Middleware {
	l10n := i18n.NewLocalizer()

	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			ctx = context.WithValue(ctx, localizerContextKey{}, l10n.GetPrinter(msg.From.LanguageCode))

			return next.Handle(ctx, msg)
		})
	}
}

// GetLocalizer retrieves the localized message printer from the provided context.
// It defaults to a preconfigured printer if the context is nil or does not contain a valid localizer.
// Returns a *message.Printer for message localization.
func GetLocalizer(ctx context.Context) *message.Printer {
	if ctx == nil {
		return defaultPrinter
	}

	l10n, ok := ctx.Value(localizerContextKey{}).(*message.Printer)
	if !ok {
		return defaultPrinter
	}

	return l10n
}
