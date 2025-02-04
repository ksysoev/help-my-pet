package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

const (
	requestTimeout = 30 * time.Second
)

// BotAPI interface represents the Telegram bot API capabilities we use
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	StopReceivingUpdates()
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type AIProvider interface {
	ProcessMessage(ctx context.Context, request *message.UserMessage) (*message.Response, error)
	ProcessEditProfile(ctx context.Context, request *message.UserMessage) (*message.Response, error)
}

// Config holds the configuration for the Telegram bot
type Config struct {
	Messages      *i18n.Config `mapstructure:"messages"`
	TelegramToken string       `mapstructure:"telegram_token"`
}

type ServiceImpl struct {
	Bot      BotAPI
	AISvc    AIProvider
	Messages *i18n.Config
	handler  Handler
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

	s := &ServiceImpl{
		Bot:      bot,
		AISvc:    aiSvc,
		Messages: cfg.Messages,
	}

	s.handler = s.setupHandler()

	return s, nil
}

func (s *ServiceImpl) processMessage(ctx context.Context, message *tgbotapi.Message) {
	// Send typing action
	typing := tgbotapi.NewChatAction(message.Chat.ID, tgbotapi.ChatTyping)
	if _, err := s.Bot.Request(typing); err != nil {
		slog.ErrorContext(ctx, "Failed to send typing action",
			slog.Any("error", err),
		)
	}

	// Handle message with middleware
	msgConfig, err := s.handler.Handle(ctx, message)

	if errors.Is(err, context.Canceled) {
		slog.InfoContext(ctx, "Request cancelled",
			slog.Int64("chat_id", message.Chat.ID),
		)

		return
	} else if err != nil {
		slog.ErrorContext(ctx, "Unexpected error",
			slog.Any("error", err),
		)
		return
	}

	// Skip sending if message is empty
	if msgConfig.Text == "" {
		return
	}

	// Send response
	if _, err := s.Bot.Send(msgConfig); err != nil {
		slog.ErrorContext(ctx, "Failed to send message",
			slog.Any("error", err),
		)
	}
}

func (s *ServiceImpl) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting Telegram bot")

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

				reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)

				// nolint:staticcheck // don't want to have dependecy on cmd package here for now
				reqCtx = context.WithValue(reqCtx, "req_id", uuid.New().String())
				// nolint:staticcheck // don't want to have dependecy on cmd package here for now
				reqCtx = context.WithValue(reqCtx, "chat_id", fmt.Sprintf("%d", update.Message.Chat.ID))

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
				slog.InfoContext(ctx, "Graceful shutdown completed")
			case <-time.After(requestTimeout):
				slog.Warn("Graceful shutdown timed out after 30 seconds")
			}

			return nil
		}
	}
}
