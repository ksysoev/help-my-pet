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

	// Handle /start command
	if msg.Text == "/start" {
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.StartMessage)), nil
	} else if msg.Text == "/terms" {
		msg := tgbotapi.NewMessage(msg.Chat.ID, termsContent)
		msg.ParseMode = "HTML"
		return msg, nil
	}

	request, err := core.NewUserMessage(
		fmt.Sprintf("%d", msg.From.ID),
		fmt.Sprintf("%d", msg.Chat.ID),
		msg.Text,
	)

	if errors.Is(err, core.ErrTextTooLong) {
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.MessageTooLong)), nil
	} else if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.ErrorMessage)), nil
	}

	response, err := s.AISvc.GetPetAdvice(ctx, request)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrRateLimit):
			return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.RateLimitMessage)), nil
		case errors.Is(err, core.ErrGlobalLimit):
			return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.GlobalLimitMessage)), nil
		default:
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to get AI response: %w", err)
		}
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
