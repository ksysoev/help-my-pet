package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// Middleware represents a function that wraps a Handler with additional functionality
type Middleware func(next Handler) Handler

// withErrorHandling middleware handles errors from the message handler
func withErrorHandling(getMessage func(lang string, msgType i18n.Message) string, next Handler) Handler {
	return func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		if message == nil {
			return tgbotapi.MessageConfig{}, errors.New("message is nil")
		}

		msgConfig, err := next(ctx, message)
		if err != nil {
			var chatID int64
			if message.Chat != nil {
				chatID = message.Chat.ID
			}

			slog.Error("Failed to handle message",
				slog.Any("error", err),
				slog.Int64("chat_id", chatID),
			)

			// Get language code safely
			var langCode string
			if message.From != nil {
				langCode = message.From.LanguageCode
			}
			// Return error message to user
			return tgbotapi.NewMessage(chatID, getMessage(langCode, i18n.ErrorMessage)), nil
		}
		return msgConfig, nil
	}
}

// withThrottler creates a middleware that limits the number of concurrent message processing
func withThrottler(maxConcurrent int) Middleware {
	// Create a buffered channel with capacity of maxConcurrent to act as a semaphore
	throttler := make(chan struct{}, maxConcurrent)

	return func(next Handler) Handler {
		return func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			if message == nil {
				return tgbotapi.MessageConfig{}, errors.New("message is nil")
			}

			// Try to acquire a slot or wait for context cancellation
			select {
			case throttler <- struct{}{}: // Acquire slot
				// Ensure we release the slot after processing
				defer func() { <-throttler }()
				// Process the message
				return next(ctx, message)
			case <-ctx.Done():
				// Context was cancelled while waiting for a slot
				return tgbotapi.MessageConfig{}, fmt.Errorf("context cancelled while waiting for throttler: %w", ctx.Err())
			}
		}
	}
}
