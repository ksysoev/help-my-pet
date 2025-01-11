package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// BotAPI interface represents the Telegram bot API capabilities we use
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	StopReceivingUpdates()
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}
