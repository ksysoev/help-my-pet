package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/help-my-pet/pkg/bot"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/prov/anthropic"
	"github.com/spf13/viper"
)

type Config struct {
	Bot struct {
		TelegramToken string `mapstructure:"telegram_token"`
	} `mapstructure:"bot"`
	AI struct {
		Model     string `mapstructure:"model"`
		Anthropic struct {
			APIKey    string `mapstructure:"api_key"`
			MaxTokens int    `mapstructure:"max_tokens"`
		} `mapstructure:"anthropic"`
	} `mapstructure:"ai"`
}

// initConfig initializes the configuration by reading from the specified config file.
func initConfig(arg *args) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("ai.model", "claude-2")
	v.SetDefault("ai.anthropic.max_tokens", 1000)

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
	if cfg.AI.Anthropic.APIKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}

	slog.Debug("Config loaded", slog.Any("config", cfg))

	return &cfg, nil
}

// runBotFunc is the function type for running the bot
type runBotFunc func(ctx context.Context, cfg *Config) error

// runBot is the default implementation
var runBot runBotFunc = func(ctx context.Context, cfg *Config) error {
	llmProvider, err := anthropic.New(anthropic.Config{
		APIKey:    cfg.AI.Anthropic.APIKey,
		Model:     cfg.AI.Model,
		MaxTokens: cfg.AI.Anthropic.MaxTokens,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Anthropic provider: %w", err)
	}

	aiService := core.NewAIService(llmProvider)
	botService := bot.NewService(cfg.Bot.TelegramToken, aiService)

	return botService.Run(ctx)
}
