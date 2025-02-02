package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

func (s *ServiceImpl) HandleCommand(_ context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	switch msg.Command() {
	case "start":
		return tgbotapi.NewMessage(msg.Chat.ID, s.Messages.GetMessage(msg.From.LanguageCode, i18n.StartMessage)), nil
	case "terms":
		msg := tgbotapi.NewMessage(msg.Chat.ID, termsContent)
		msg.ParseMode = "HTML"
		return msg, nil
	default:
		return tgbotapi.NewMessage(msg.Chat.ID, "Unknown command"), nil
	}
}
