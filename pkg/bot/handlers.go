package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/middleware"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

// Handler represents a function that handles a Telegram message
type Handler interface {
	Handle(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error)
}

// setupHandler sets up the message handler with all middleware
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

func (s *ServiceImpl) Handle(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	slog.DebugContext(ctx, "Received message", slog.String("text", message.Text))

	if message.Text == "" {
		return tgbotapi.MessageConfig{}, nil
	}

	// Handle /start command
	if message.Text == "/start" {
		return tgbotapi.NewMessage(message.Chat.ID, s.Messages.GetMessage(message.From.LanguageCode, i18n.StartMessage)), nil
	}

	if message.From == nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("message from is nil")
	}

	request, err := core.NewUserMessage(
		fmt.Sprintf("%d", message.From.ID),
		fmt.Sprintf("%d", message.Chat.ID),
		message.Text,
	)

	if errors.Is(err, core.ErrTextTooLong) {
		return tgbotapi.NewMessage(message.Chat.ID, s.Messages.GetMessage(message.From.LanguageCode, i18n.MessageTooLong)), nil
	} else if err != nil {
		return tgbotapi.NewMessage(message.Chat.ID, s.Messages.GetMessage(message.From.LanguageCode, i18n.ErrorMessage)), nil
	}

	response, err := s.AISvc.GetPetAdvice(ctx, request)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrRateLimit):
			return tgbotapi.NewMessage(message.Chat.ID, s.Messages.GetMessage(message.From.LanguageCode, i18n.RateLimitMessage)), nil
		case errors.Is(err, core.ErrGlobalLimit):
			return tgbotapi.NewMessage(message.Chat.ID, s.Messages.GetMessage(message.From.LanguageCode, i18n.GlobalLimitMessage)), nil
		default:
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to get AI response: %w", err)
		}
	}

	// Create message with buttons if available
	msg := tgbotapi.NewMessage(message.Chat.ID, response.Message)

	// Handle keyboard markup based on answers
	if len(response.Answers) > 0 {
		keyboard := make([][]tgbotapi.KeyboardButton, len(response.Answers))
		for i, answer := range response.Answers {
			keyboard[i] = []tgbotapi.KeyboardButton{
				{Text: answer},
			}
		}
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:        keyboard,
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
		}
	} else {
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
			Selective:      false,
		}
	}

	return msg, nil
}
