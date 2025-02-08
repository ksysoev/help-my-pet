package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/middleware"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// Handler defines the interface for processing and responding to incoming messages in a Telegram bot context.
// It handles a message by performing necessary processing and returns the configuration for the outgoing message or an error.
// ctx is the context for managing request lifecycle and cancellation.
// message is the incoming Telegram message to be processed.
// Returns a configured message object for sending a response and an error if processing fails.
type Handler interface {
	Handle(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error)
}

// setupHandler initializes and configures the request handler with specified middleware components.
// It applies middleware for request reduction, concurrency throttling, metric collection, and error handling,
// ensuring proper management of requests and enhanced error messages.
// Returns a Handler that processes messages with the applied middleware stack.
func (s *ServiceImpl) setupHandler() Handler {
	h := middleware.Use(
		s,
		middleware.WithRequestReducer(),
		middleware.WithThrottler(30),
		middleware.WithMetrics(),
		middleware.WithErrorHandling(s.Messages.GetMessage),
	)

	return h
}

// Handle processes an incoming Telegram message and generates an appropriate response.
// It validates the message, handles commands, constructs contextual responses using AI,
// and supports reply markup for user interactions.
// Accepts ctx, the request context for cancellation or deadlines, and msg, the Telegram
// message to process.
// Returns a configured Telegram message response (tgbotapi.MessageConfig) or an error
// if validation fails, the input is unsupported, or message handling encounters issues.
func (s *ServiceImpl) Handle(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	slog.DebugContext(ctx, "Received msg", slog.String("text", msg.Text))

	if msg.From == nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("msg from is nil")
	}

	// Validate if message contains unsupported media type like images, videos, etc.
	if msg.Photo != nil || msg.Video != nil || msg.Audio != nil || msg.Voice != nil || msg.Document != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.UnsupportedMediaType)), nil
	}

	if msg.Text == "" {
		return tgbotapi.MessageConfig{}, nil
	}

	if msg.Command() != "" {
		return s.HandleCommand(ctx, msg)
	}

	request, err := message.NewUserMessage(
		fmt.Sprintf("%d", msg.From.ID),
		fmt.Sprintf("%d", msg.Chat.ID),
		msg.Text,
	)

	if errors.Is(err, message.ErrTextTooLong) {
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.MessageTooLong)), nil
	} else if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.ErrorMessage)), nil
	}

	response, err := s.AISvc.ProcessMessage(ctx, request)
	if err != nil {
		return s.handleProcessingError(err, msg)
	}

	// Create msg with buttons if available
	resp := tgbotapi.NewMessage(msg.Chat.ID, response.Message)

	// Handle keyboard markup based on answers
	if len(response.Answers) > 0 {
		keyboard := make([][]tgbotapi.KeyboardButton, len(response.Answers))
		for i, answer := range response.Answers {
			keyboard[i] = []tgbotapi.KeyboardButton{
				{Text: answer},
			}
		}
		resp.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:        keyboard,
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
		}
	} else {
		resp.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
			Selective:      false,
		}
	}

	return resp, nil
}

// handleProcessingError maps specific processing errors to localized user-facing messages or provides a default error response.
// It accepts err, the error encountered during message handling, and msg, the user's incoming message for context.
// Returns a configured message with an appropriate response and an error if the failure is unrecognized or unexpected.
func (s *ServiceImpl) handleProcessingError(err error, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	switch {
	case errors.Is(err, core.ErrRateLimit):
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.RateLimitMessage)), nil
	case errors.Is(err, core.ErrGlobalLimit):
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.GlobalLimitMessage)), nil
	case errors.Is(err, message.ErrTextTooLong):
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.MessageTooLong)), nil
	case errors.Is(err, message.ErrFutureDate):
		return tgbotapi.NewMessage(msg.Chat.ID, "Provided date cannot be in the future. Please provide a valid date."), nil
	case errors.Is(err, message.ErrInvalidDates):
		return tgbotapi.NewMessage(msg.Chat.ID, "Please provide a date in the valid format YYYY-MM-DD (e.g., 2023-12-31)"), nil
	default:
		return tgbotapi.MessageConfig{}, fmt.Errorf("failed to get AI response: %w", err)
	}
}
