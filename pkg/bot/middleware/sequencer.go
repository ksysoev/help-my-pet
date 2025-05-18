package middleware

import (
	"context"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WithRequestSequencer enforces sequential processing of requests for each user to ensure only one active request at a time.
// It maintains a queue per user to serialize requests and handles context cancellation gracefully.
// Returns Middleware that applies this sequencing behavior to a given Handler.
func WithRequestSequencer() Middleware {
	// Map to store request queues for each user
	var (
		userQueues = make(map[int64]chan struct{})
		mu         sync.Mutex
	)

	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
			userID := message.From.ID

			curReq := make(chan struct{})
			mu.Lock()
			lastReq, exists := userQueues[userID]
			if !exists {
				close(lastReq)
			} else {
				lastReq = make(chan struct{})
				close(lastReq)
			}
			userQueues[userID] = curReq
			mu.Unlock()

			defer func() {
				close(curReq)

				mu.Lock()
				defer mu.Unlock()

				if curReq == userQueues[userID] {
					delete(userQueues, userID)
				}
			}()

			select {
			case <-lastReq:

				return next.Handle(ctx, message)
			case <-ctx.Done():
				// Context was cancelled while waiting for our turn
				return tgbotapi.MessageConfig{}, fmt.Errorf("context cancelled while waiting for user's previous requests to complete: %w", ctx.Err())
			}
		})
	}
}
