package middleware

import (
	"context"
	"errors"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// WithErrorHandling middleware handles errors from the message handler
func WithErrorHandling(getMessage func(lang string, msgType i18n.Message) string) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			if message == nil {
				return tgbotapi.MessageConfig{}, errors.New("message is nil")
			}

			msgConfig, err := next.Handle(ctx, message)
			if err != nil {
				var chatID int64
				if message.Chat != nil {
					chatID = message.Chat.ID
				}

				slog.ErrorContext(ctx, "Failed to handle message", slog.Any("error", err))

				// Get language code safely
				var langCode string
				if message.From != nil {
					langCode = message.From.LanguageCode
				}
				// Return error message to user
				return tgbotapi.NewMessage(chatID, getMessage(langCode, i18n.ErrorMessage)), nil
			}
			return msgConfig, nil
		})
	}
}
