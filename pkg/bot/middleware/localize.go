package middleware

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// WithLocalization wraps a Handler with language-specific localization support for incoming messages.
// It retrieves the user's language code from the incoming message and attaches a localizer instance to the context.
// Returns Middleware that ensures the context contains a localized message printer for the message's language.
func WithLocalization() Middleware {
	l10n := i18n.NewLocalizer()

	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			lang := ""
			if msg.From != nil {
				lang = msg.From.LanguageCode
			}

			ctx = i18n.SetLocale(ctx, l10n, lang)

			return next.Handle(ctx, msg)
		})
	}
}
