package bot

import (
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"github.com/ksysoev/help-my-pet/pkg/repo/memory"
)

// Config holds the configuration for the Telegram bot
type Config struct {
	Messages      *i18n.Config            `mapstructure:"messages"`
	RateLimit     *memory.RateLimitConfig `mapstructure:"rate_limit"`
	TelegramToken string                  `mapstructure:"telegram_token"`
}
