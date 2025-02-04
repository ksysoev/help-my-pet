package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

func (s *ServiceImpl) HandleCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	switch msg.Command() {
	case "start":
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.StartMessage)), nil
	case "terms":
		msg := tgbotapi.NewMessage(msg.Chat.ID, termsContent)
		msg.ParseMode = "HTML"
		return msg, nil
	case "editprofile":
		req, err := message.NewUserMessage(
			fmt.Sprintf("%d", msg.From.ID),
			fmt.Sprintf("%d", msg.Chat.ID),
			msg.Text,
		)
		if err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to create user message: %w", err)
		}

		resp, err := s.AISvc.ProcessEditProfile(ctx, req)
		if err != nil {

			slog.ErrorContext(ctx, "failed to process edit profile", slog.Any("error", err))
			return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.ErrorMessage)), nil
		}

		return tgbotapi.NewMessage(msg.Chat.ID, resp.Message), nil
	default:
		return tgbotapi.NewMessage(msg.Chat.ID, "Unknown command"), nil
	}
}
