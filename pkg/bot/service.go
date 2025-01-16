package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core"
)

type AIProvider interface {
	GetPetAdvice(ctx context.Context, userID string, question string) (*core.PetAdviceResponse, error)
	Start(ctx context.Context) (string, error)
}

type ServiceImpl struct {
	Bot   BotAPI
	AISvc AIProvider
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

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	return &ServiceImpl{
		Bot:   bot,
		AISvc: aiSvc,
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
		response, err := s.AISvc.Start(ctx)
		if err != nil {
			slog.Error("Failed to get start message",
				slog.Any("error", err),
				slog.Int64("chat_id", message.Chat.ID),
			)
			s.sendErrorMessage(message.Chat.ID)
			return
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, response)
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
	response, err := s.AISvc.GetPetAdvice(ctx, fmt.Sprintf("%d", userID), message.Text)
	if err != nil {
		slog.Error("Failed to get AI response",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
		s.sendErrorMessage(message.Chat.ID)
		return
	}

	// Create message with buttons if available
	msg := tgbotapi.NewMessage(message.Chat.ID, response.Message)
	msg.ReplyToMessageID = message.MessageID

	// Add reply keyboard if there are answers
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
	}

	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send message",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
		return
	}

}

func (s *ServiceImpl) sendErrorMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Sorry, I encountered an error while processing your request. Please try again later.")
	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send error message",
			slog.Any("error", err),
			slog.Int64("chat_id", chatID),
		)
	}
}
