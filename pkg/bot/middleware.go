package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// Middleware represents a function that wraps a Handler with additional functionality
type Middleware func(next Handler) Handler

func UseMiddleware(handler Handler, middlewares ...Middleware) Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

// WithErrorHandling middleware handles errors from the message handler
func WithErrorHandling(getMessage func(lang string, msgType i18n.Message) string) Middleware {
	return func(next Handler) Handler {
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
		}
	}
}

// WithThrottler creates a middleware that limits the number of concurrent message processing
func WithThrottler(maxConcurrent int) Middleware {
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

type requestState struct {
	cancel    context.CancelFunc
	messageID int
}

// WithRequestReducer creates middleware that reduces multiple concurrent requests from the same chat to a single active request
func WithRequestReducer() Middleware {
	var mu sync.RWMutex
	activeRequests := make(map[int64]requestState)

	return func(next Handler) Handler {
		return func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			if message == nil {
				return tgbotapi.MessageConfig{}, errors.New("message is nil")
			}

			chatID := message.Chat.ID

			// Create new context and cancel function for this request
			reqCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			// Cancel any existing request for this chat
			mu.Lock()
			if existing, exists := activeRequests[chatID]; exists {
				existing.cancel()
				delete(activeRequests, chatID)
			}
			activeRequests[chatID] = requestState{
				cancel:    cancel,
				messageID: message.MessageID,
			}
			mu.Unlock()

			// Cleanup when context is done
			go func() {
				<-reqCtx.Done()
				mu.Lock()
				if state, exists := activeRequests[chatID]; exists && state.messageID == message.MessageID {
					delete(activeRequests, chatID)
				}
				mu.Unlock()
			}()

			return next(reqCtx, message)
		}
	}
}

// withREDMetrics wraps a Handler to measure and log request duration and error occurrence for performance monitoring.
func withREDMetrics() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			start := time.Now()
			resp, err := next(ctx, message)

			slog.InfoContext(ctx, "Message processing time", slog.Duration("duration", time.Since(start)), slog.Bool("error", err != nil))

			return resp, err
		}
	}
}
