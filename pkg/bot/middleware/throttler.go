package middleware

import (
	"context"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WithThrottler limits the number of concurrent handler executions by ensuring no more than maxConcurrent routines run.
// It uses a buffered channel as a semaphore to manage concurrency, blocking excess requests until a slot is available.
// Accepts maxConcurrent, the maximum number of concurrent executions allowed.
// Returns a Middleware that enforces the concurrency limit and an error if context is cancelled or message is nil.
func WithThrottler(maxConcurrent int) Middleware {
	// Create a buffered channel with capacity of maxConcurrent to act as a semaphore
	throttler := make(chan struct{}, maxConcurrent)

	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			if message == nil {
				return tgbotapi.MessageConfig{}, errors.New("message is nil")
			}

			// Try to acquire a slot or wait for context cancellation
			select {
			case throttler <- struct{}{}: // Acquire slot
				// Ensure we release the slot after processing
				defer func() { <-throttler }()
				// Process the message
				return next.Handle(ctx, message)
			case <-ctx.Done():
				// Context was cancelled while waiting for a slot
				return tgbotapi.MessageConfig{}, fmt.Errorf("context cancelled while waiting for throttler: %w", ctx.Err())
			}
		})
	}
}
