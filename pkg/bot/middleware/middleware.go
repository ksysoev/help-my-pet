package middleware

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Middleware represents a function that wraps a Handler with additional functionality
type Middleware func(next Handler) Handler

type Handler interface {
	Handle(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error)
}

type HandlerFunc func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error)

func (h HandlerFunc) Handle(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	return h(ctx, message)
}

func Use(handler Handler, middlewares ...Middleware) Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
