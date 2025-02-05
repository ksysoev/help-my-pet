package middleware

import (
	"context"
	"errors"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// requestState holds the context cancel function and message id for a specific request to manage concurrency.
type requestState struct {
	cancel    context.CancelFunc
	messageID int
}

// WithRequestReducer limits concurrent message handling per chat by canceling previous requests for the same chat.
// It ensures only the latest request for a given chat id is processed, canceling any prior requests automatically.
// Returns a Middleware that wraps a Handler to provide this functionality.
// It returns an error if nil message is passed to the Handler.
func WithRequestReducer() Middleware {
	var mu sync.RWMutex
	activeRequests := make(map[int64]requestState)

	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
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

			return next.Handle(reqCtx, message)
		})
	}
}
