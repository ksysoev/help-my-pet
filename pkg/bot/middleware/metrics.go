package middleware

import (
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WithREDMetrics wraps a Handler to measure and log request duration and error occurrence for performance monitoring.
func WithREDMetrics() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			start := time.Now()
			resp, err := next.Handle(ctx, message)

			slog.InfoContext(ctx, "Message processing time", slog.Duration("duration", time.Since(start)), slog.Bool("error", err != nil))

			return resp, err
		})
	}
}
