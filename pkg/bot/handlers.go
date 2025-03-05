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
		middleware.WithErrorHandling(),
		middleware.WithLocalization(),
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
	if msg.Video != nil || msg.Audio != nil || msg.Voice != nil || msg.Document != nil {
		return tgbotapi.NewMessage(
			msg.Chat.ID,
			i18n.GetLocale(ctx).Sprintf("Sorry, I cannot process videos, audio, or documents. Please send your question as text only."),
		), nil
	}

	if len(msg.Photo) > 0 {
		resp, err := s.handlePhoto(ctx, msg)

		if err != nil {
			return s.handleProcessingError(ctx, err, msg)
		}

		return resp, nil
	}

	if msg.Text == "" {
		return tgbotapi.MessageConfig{}, nil
	}

	if msg.Command() != "" {
		resp, err := s.HandleCommand(ctx, msg)
		if err != nil {
			return s.handleProcessingError(ctx, err, msg)
		}

		return resp, nil
	}

	request, err := message.NewUserMessage(
		fmt.Sprintf("%d", msg.From.ID),
		fmt.Sprintf("%d", msg.Chat.ID),
		msg.Text,
	)

	if errors.Is(err, message.ErrTextTooLong) {
		return tgbotapi.NewMessage(
			msg.Chat.ID,
			i18n.GetLocale(ctx).Sprintf("I apologize, but your message is too long for me to process. Please try to make it shorter and more concise."),
		), nil
	} else if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("failed to create user message: %w", err)
	}

	response, err := s.AISvc.ProcessMessage(ctx, request)
	if err != nil {
		return s.handleProcessingError(ctx, err, msg)
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
func (s *ServiceImpl) handleProcessingError(ctx context.Context, err error, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	switch {
	case errors.Is(err, core.ErrRateLimit):
		return tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf("You have reached the maximum number of requests per hour. Please try again later.")), nil
	case errors.Is(err, core.ErrGlobalLimit):
		return tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf("We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.")), nil
	default:
		return tgbotapi.MessageConfig{}, fmt.Errorf("failed to get AI response: %w", err)
	}
}

// HandleRemovingBot resets the conversation context for a user in a chat upon bot removal.
// It ensures the conversation state is cleared for the given userID and chatID.
// Returns error if the reset operation fails, including details about the failure.
func (s *ServiceImpl) HandleRemovingBot(ctx context.Context, userID, chatID string) error {
	if err := s.AISvc.ResetUserConversation(ctx, userID, chatID); err != nil {
		return fmt.Errorf("failed to reset user conversation: %w", err)
	}

	return nil
}
