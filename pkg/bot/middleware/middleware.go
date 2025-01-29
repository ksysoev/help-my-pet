package middleware

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Middleware defines a function that wraps a Handler with additional behavior or processing logic.
// It takes a Handler as input and returns a new, modified Handler that incorporates the middleware's functionality.
type Middleware func(next Handler) Handler

// Handler defines the interface for processing incoming messages in a bot framework.
// It accepts a context for request-scoped values and cancellation signals, and the message to be handled.
// Returns a configured message response and an error if the processing fails.
type Handler interface {
	Handle(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error)
}

// HandlerFunc processes an incoming Telegram message within a given context and generates a response configuration.
// It takes a context for controlling execution and a pointer to the incoming Telegram message as input parameters.
// Returns a MessageConfig containing the response to be sent and an error if message handling fails.
type HandlerFunc func(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error)

// Handle executes the HandlerFunc with the provided context and Telegram message.
// It processes the incoming message and generates a response configuration.
// Returns MessageConfig containing the response settings and error if the handler execution fails.
func (h HandlerFunc) Handle(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	return h(ctx, message)
}

// Use composes a new Handler by wrapping the provided handler with the given middlewares in the specified order.
// It iterates over the middlewares and applies each one sequentially, returning the final wrapped handler.
// Accepts handler, the base Handler to wrap, and middlewares, a variadic list of Middleware functions to apply.
// Returns a Handler, which represents the wrapped composition of the specified handler and middlewares.
func Use(handler Handler, middlewares ...Middleware) Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
