package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/help-my-pet/pkg/bot"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/spf13/viper"
)

type Config struct {
	Bot struct {
		TelegramToken string `mapstructure:"telegram_token"`
	} `mapstructure:"bot"`
	AI struct {
		AnthropicKey string `mapstructure:"anthropic_key"`
		Model        string `mapstructure:"model"`
	} `mapstructure:"ai"`
}

// initConfig initializes the configuration by reading from the specified config file.
func initConfig(arg *args) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("ai.model", "claude-2")

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
	if cfg.AI.AnthropicKey == "" {
		return nil, fmt.Errorf("anthropic key is required")
	}

	slog.Debug("Config loaded", slog.Any("config", cfg))

	return &cfg, nil
}

type BotServiceFactory func(token string, aiSvc *core.AIService) BotService

type BotService interface {
	Run(ctx context.Context) error
}

// defaultBotServiceFactory creates a real bot service
func defaultBotServiceFactory(token string, aiSvc *core.AIService) BotService {
	return bot.NewService(token, aiSvc)
}

// runBotFunc is the function type for running the bot
type runBotFunc func(ctx context.Context, cfg *Config, factory BotServiceFactory) error

// runBot is the default implementation
var runBot runBotFunc = func(ctx context.Context, cfg *Config, factory BotServiceFactory) error {
	if factory == nil {
		factory = defaultBotServiceFactory
	}

	aiService := core.NewAIService(cfg.AI.AnthropicKey, cfg.AI.Model)
	botService := factory(cfg.Bot.TelegramToken, aiService)

	return botService.Run(ctx)
}
