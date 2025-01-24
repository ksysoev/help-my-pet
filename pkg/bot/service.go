package bot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

const (
	requestTimeout = 30 * time.Second
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

func (s *ServiceImpl) processMessage(ctx context.Context, message *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Send typing action
	typing := tgbotapi.NewChatAction(message.Chat.ID, tgbotapi.ChatTyping)
	if _, err := s.Bot.Request(typing); err != nil {
		slog.Error("Failed to send typing action",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
	}

	// Handle message with middleware
	handler := s.setupHandler()
	msgConfig, err := handler(ctx, message)
	if err != nil {
		slog.Error("Unexpected error",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
		return
	}

	// Skip sending if message is empty
	if msgConfig.Text == "" {
		return
	}

	// Send response
	if _, err := s.Bot.Send(msgConfig); err != nil {
		slog.Error("Failed to send message",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
	}
}

func (s *ServiceImpl) Run(ctx context.Context) error {
	slog.Info("Starting Telegram bot")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := s.Bot.GetUpdatesChan(updateConfig)

	var wg sync.WaitGroup

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			wg.Add(1)

			go func() {
				defer wg.Done()

				reqCtx, cancel := context.WithTimeout(context.Background(), requestTimeout)

				defer cancel()

				s.processMessage(reqCtx, update.Message)
			}()

		case <-ctx.Done():
			slog.Info("Starting graceful shutdown")
			s.Bot.StopReceivingUpdates()

			// Wait for ongoing message processors with a timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				slog.Info("Graceful shutdown completed")
			case <-time.After(requestTimeout):
				slog.Warn("Graceful shutdown timed out after 30 seconds")
			}

			return nil
		}
	}
}
