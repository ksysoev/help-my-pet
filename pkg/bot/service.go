package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AIProvider interface {
	GetPetAdvice(ctx context.Context, question string) (string, error)
}

type Service struct {
	bot   *tgbotapi.BotAPI
	aiSvc AIProvider
}

func NewService(token string, aiSvc AIProvider) *Service {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Telegram bot: %v", err))
	}

	return &Service{
		bot:   bot,
		aiSvc: aiSvc,
	}
}

func (s *Service) Run(ctx context.Context) error {
	slog.Info("Starting Telegram bot")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := s.bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			go s.handleMessage(ctx, update.Message)

		case <-ctx.Done():
			slog.Info("Shutting down bot")
			s.bot.StopReceivingUpdates()
			return nil
		}
	}
}

func (s *Service) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	slog.Info("Received message",
		slog.Int64("chat_id", message.Chat.ID),
		slog.String("text", message.Text),
	)

	if message.Text == "" {
		return
	}

	// Send typing action
	typing := tgbotapi.NewChatAction(message.Chat.ID, tgbotapi.ChatTyping)
	s.bot.Send(typing)

	// Get AI response
	response, err := s.aiSvc.GetPetAdvice(ctx, message.Text)
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

	if _, err := s.bot.Send(msg); err != nil {
		slog.Error("Failed to send message",
			slog.Any("error", err),
			slog.Int64("chat_id", message.Chat.ID),
		)
	}
}

func (s *Service) sendErrorMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Sorry, I encountered an error while processing your request. Please try again later.")
	s.bot.Send(msg)
}
