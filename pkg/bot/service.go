package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/ratelimit"
)

// RateLimiter defines the interface for rate limiting functionality
type RateLimiter interface {
	IsAllowed(ctx context.Context, userID int64) (bool, error)
	RecordAccess(ctx context.Context, userID int64) error
}

type AIProvider interface {
	GetPetAdvice(ctx context.Context, chatID string, question string) (string, error)
	Start(ctx context.Context) (string, error)
}

type ServiceImpl struct {
	Bot         BotAPI
	AISvc       AIProvider
	rateLimiter RateLimiter
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

	var limiter RateLimiter
	if cfg.RateLimit != nil {
		limiter = ratelimit.NewRateLimiter(cfg.RateLimit)
	}

	return &ServiceImpl{
		Bot:         bot,
		AISvc:       aiSvc,
		rateLimiter: limiter,
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

	// Skip rate limiting for /start command or if rate limiter is not configured
	if message.Text != "/start" && s.rateLimiter != nil && message.From != nil {
		allowed, err := s.rateLimiter.IsAllowed(ctx, message.From.ID)
		if err != nil {
			slog.Error("Rate limit exceeded",
				slog.Any("error", err),
				slog.Int64("user_id", message.From.ID),
			)
			s.sendRateLimitExceededMessage(message.Chat.ID)
			return
		}
		if !allowed {
			s.sendRateLimitExceededMessage(message.Chat.ID)
			return
		}
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
	response, err := s.AISvc.GetPetAdvice(ctx, fmt.Sprintf("%d", message.Chat.ID), message.Text)
	if err != nil {
		slog.Error("Failed to get AI response",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
		s.sendErrorMessage(message.Chat.ID)
		return
	}

	// Send response
	msg := tgbotapi.NewMessage(message.Chat.ID, response)
	msg.ReplyToMessageID = message.MessageID

	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send message",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
		return
	}

	// Record the access after successful response
	if message.Text != "/start" && s.rateLimiter != nil && message.From != nil {
		if err := s.rateLimiter.RecordAccess(ctx, message.From.ID); err != nil {
			slog.Error("Failed to record rate limit access",
				slog.Any("error", err),
				slog.Int64("user_id", message.From.ID),
			)
		}
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

func (s *ServiceImpl) sendRateLimitExceededMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Rate limit exceeded. Please try again later.")
	if _, err := s.Bot.Send(msg); err != nil {
		slog.Error("Failed to send rate limit message",
			slog.Any("error", err),
			slog.Int64("chat_id", chatID),
		)
	}
}
