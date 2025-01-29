package middleware

import (
	"context"
	"errors"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type requestState struct {
	cancel    context.CancelFunc
	messageID int
}

// WithRequestReducer creates middleware that reduces multiple concurrent requests from the same chat to a single active request
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
