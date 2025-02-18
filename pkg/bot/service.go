package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/ksysoev/help-my-pet/pkg/bot/media"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
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
	GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error)
}

type AIProvider interface {
	ProcessMessage(ctx context.Context, request *message.UserMessage) (*message.Response, error)
	ProcessEditProfile(ctx context.Context, request *message.UserMessage) (*message.Response, error)
	CancelQuestionnaire(ctx context.Context, chatID string) error
}

type httpClient interface {
	Get(url string) (*http.Response, error)
}

// Config holds the configuration for the Telegram bot
type Config struct {
	TelegramToken string `mapstructure:"telegram_token"`
}

type ServiceImpl struct {
	token      string
	Bot        BotAPI
	AISvc      AIProvider
	handler    Handler
	collector  *media.Collector
	httpClient httpClient
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

	s := &ServiceImpl{
		token:     cfg.TelegramToken,
		Bot:       bot,
		AISvc:     aiSvc,
		collector: media.NewCollector(),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	s.handler = s.setupHandler()

	return s, nil
}

func (s *ServiceImpl) processMessage(ctx context.Context, message *tgbotapi.Message) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)

	go func() {
		defer wg.Done()
		s.keepTyping(ctx, message.Chat.ID, 5*time.Second)
	}()

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
	cancel()

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

// sendTyping sends a "typing" action to the specified chat to indicate activity to the user.
// It takes a context for request scoping and chatID to identify the target chat.
// Returns an error if the request to the bot API fails.
func (s *ServiceImpl) sendTyping(ctx context.Context, chatID int64) {
	typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := s.Bot.Request(typing); err != nil {
		slog.ErrorContext(ctx, "Failed to send typing action",
			slog.Any("error", err),
		)
	}
}

// keepTyping continuously sends typing notifications to the specified chat at a given interval until the context is canceled.
// ctx is the context controlling the lifecycle of the typing notifications.
// chatID is the identifier of the chat to send typing notifications to.
// interval specifies the duration between consecutive typing notifications.
func (s *ServiceImpl) keepTyping(ctx context.Context, chatID int64, interval time.Duration) {
	s.sendTyping(ctx, chatID)

	go func() {
		t := time.NewTicker(interval)
		for {
			select {
			case <-ctx.Done():
				t.Stop()
				return
			case <-t.C:
				s.sendTyping(ctx, chatID)
			}
		}
	}()
}
