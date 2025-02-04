package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/help-my-pet/pkg/bot"
	"github.com/ksysoev/help-my-pet/pkg/prov/anthropic"
	"github.com/ksysoev/help-my-pet/pkg/repo/memory"
	"github.com/spf13/viper"
)

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Config struct {
	Bot       bot.Config             `mapstructure:"bot"`
	AI        anthropic.Config       `mapstructure:"ai"`
	Redis     RedisConfig            `mapstructure:"redis"`
	RateLimit memory.RateLimitConfig `mapstructure:"rate_limit"`
}

// initConfig initializes the configuration by reading from the specified config file.
func initConfig(arg *args) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("ai.model", "claude-2")
	v.SetDefault("ai.max_tokens", 1000)
	v.SetDefault("redis.url", "redis://localhost:6379")
	v.SetDefault("redis.db", 0)

	if arg.ConfigPath != "" {
		v.SetConfigFile(arg.ConfigPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if cfg.Bot.TelegramToken == "" {
		return nil, fmt.Errorf("telegram token is required")
	}
	if cfg.AI.APIKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}

	slog.Debug("Config loaded", slog.Any("config", cfg))

	return &cfg, nil
}
