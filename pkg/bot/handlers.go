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

func (s *ServiceImpl) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	slog.Info("Received message",
		slog.Int64("chat_id", message.Chat.ID),
		slog.String("text", message.Text),
	)

	if message.Text == "" {
		return
	}

	// Handle /start command
	if message.Text == "/start" {
		msg := tgbotapi.NewMessage(message.Chat.ID, s.Messages.GetMessage(message.From.LanguageCode, i18n.StartMessage))
		if _, err := s.Bot.Send(msg); err != nil {
			slog.Error("Failed to send start message",
				slog.Any("error", err),
				slog.Int64("chat_id", message.Chat.ID),
			)
		}
		return
	}

	// Send typing action
	typing := tgbotapi.NewChatAction(message.Chat.ID, tgbotapi.ChatTyping)
	if _, err := s.Bot.Send(typing); err != nil {
		slog.Error("Failed to send typing action",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
	}

	request := &core.PetAdviceRequest{
		ChatID:  fmt.Sprintf("%d", message.Chat.ID),
		Message: message.Text,
	}

	if message.From == nil {
		slog.Warn("Message from is nil",
			slog.Any("message", message),
		)
		return
	}

	request.UserID = fmt.Sprintf("%d", message.From.ID)

	response, err := s.AISvc.GetPetAdvice(ctx, request)
	if err != nil {
		slog.Error("Failed to get AI response",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
		switch {
		case errors.Is(err, core.ErrRateLimit):
			s.sendRateLimitMessage(message.Chat.ID, message.From.LanguageCode)
		case errors.Is(err, core.ErrGlobalLimit):
			s.sendGlobalLimitMessage(message.Chat.ID, message.From.LanguageCode)
		default:
			s.sendErrorMessage(message.Chat.ID, message.From.LanguageCode)
		}
		return
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

	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send message",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
		return
	}
}

func (s *ServiceImpl) sendErrorMessage(chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, s.Messages.GetMessage(lang, i18n.ErrorMessage))
	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send error message",
			slog.Any("error", err),
			slog.Int64("chat_id", chatID),
		)
	}
}

func (s *ServiceImpl) sendRateLimitMessage(chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, s.Messages.GetMessage(lang, i18n.RateLimitMessage))
	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send rate limit message",
			slog.Any("error", err),
			slog.Int64("chat_id", chatID),
		)
	}
}

func (s *ServiceImpl) sendGlobalLimitMessage(chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, s.Messages.GetMessage(lang, i18n.GlobalLimitMessage))
	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send global limit message",
			slog.Any("error", err),
			slog.Int64("chat_id", chatID),
		)
	}
}
