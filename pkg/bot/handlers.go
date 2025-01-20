package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

func (s *ServiceImpl) handleMessage(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	slog.Info("Received message",
		slog.Int64("chat_id", message.Chat.ID),
		slog.String("text", message.Text),
	)

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

	request := &core.PetAdviceRequest{
		ChatID:  fmt.Sprintf("%d", message.Chat.ID),
		Message: message.Text,
		UserID:  fmt.Sprintf("%d", message.From.ID),
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
	msg.ReplyToMessageID = message.MessageID

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
