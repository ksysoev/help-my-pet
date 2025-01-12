package bot

import "github.com/ksysoev/help-my-pet/pkg/ratelimit"

// Config holds the configuration for the Telegram bot
type Config struct {
	RateLimit     *ratelimit.Config `mapstructure:"rate_limit"`
	TelegramToken string            `mapstructure:"telegram_token"`
}
