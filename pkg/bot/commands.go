package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/middleware"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

func (s *ServiceImpl) HandleCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	switch msg.Command() {
	case "start":
		startMsg := middleware.GetLocalizer(ctx).Sprintf(`Welcome to Help My Pet Bot! 🐾

I'm your personal pet care assistant, ready to help you take better care of your furry friend. I can assist you with:

• Pet health and behavior questions
• Diet and nutrition advice
• Training tips and techniques
• General pet care guidance

Simply type your question or concern about your pet, and I'll provide helpful, informative answers based on reliable veterinary knowledge. Remember, while I can offer guidance, for serious medical conditions, always consult with a veterinarian.

To get started, just ask me any question about your pet!`)

		return tgbotapi.NewMessage(msg.Chat.ID, startMsg), nil
	case "terms":
		msg := tgbotapi.NewMessage(msg.Chat.ID, middleware.GetLocalizer(ctx).Sprintf(termsContent))
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
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to process edit profile request: %w", err)
		}

		return tgbotapi.NewMessage(msg.Chat.ID, resp.Message), nil
	default:
		return tgbotapi.NewMessage(msg.Chat.ID, "Unknown command"), nil
	}
}
