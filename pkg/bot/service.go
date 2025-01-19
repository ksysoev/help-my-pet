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

type AIProvider interface {
	GetPetAdvice(ctx context.Context, request *core.PetAdviceRequest) (*core.PetAdviceResponse, error)
}

type ServiceImpl struct {
	Bot      BotAPI
	AISvc    AIProvider
	Messages *i18n.Config
}

// NewService creates a new bot service with the given configuration and AI provider
func NewService(cfg *Config, aiSvc AIProvider) (*ServiceImpl, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if aiSvc == nil {
		return nil, fmt.Errorf("AI provider cannot be nil")
	}

	if cfg.TelegramToken == "" {
		return nil, fmt.Errorf("telegram token cannot be empty")
	}

	if cfg.Messages == nil {
		return nil, fmt.Errorf("messages config cannot be nil")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	return &ServiceImpl{
		Bot:      bot,
		AISvc:    aiSvc,
		Messages: cfg.Messages,
	}, nil
}

func (s *ServiceImpl) Run(ctx context.Context) error {
	slog.Info("Starting Telegram bot")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := s.Bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			go s.handleMessage(ctx, update.Message)

		case <-ctx.Done():
			slog.Info("Shutting down bot")
			s.Bot.StopReceivingUpdates()
			return nil
		}
	}
}

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

	// Get AI response
	userID := message.Chat.ID
	if message.From != nil {
		userID = message.From.ID
	}
	request := &core.PetAdviceRequest{
		UserID:  fmt.Sprintf("%d", userID),
		ChatID:  fmt.Sprintf("%d", userID),
		Message: message.Text,
	}
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
