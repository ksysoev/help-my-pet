package middleware

import (
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WithMetrics wraps a Handler to record processing time and error occurrence metrics for each message processed.
// It logs the duration of message processing and whether an error occurred during execution.
// Returns a Middleware that measures and logs performance metrics for the wrapped Handler.
func WithMetrics() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			start := time.Now()
			resp, err := next.Handle(ctx, message)

			slog.InfoContext(ctx, "Message processing time", slog.Duration("duration", time.Since(start)), slog.Bool("error", err != nil))

			return resp, err
		})
	}
}
