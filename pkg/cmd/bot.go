package cmd

import (
	"context"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/bot"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/prov/anthropic"
)

// BotRunner handles the initialization and running of the Telegram bot.
type BotRunner struct {
	botService    *bot.Service
	llmProvider   core.LLM
	createService func(cfg *bot.Config, aiSvc bot.AIProvider) (*bot.Service, error)
}

// NewBotRunner creates a new BotRunner with default implementations.
func NewBotRunner() *BotRunner {
	return &BotRunner{
		createService: bot.NewService,
	}
}

// WithBotService sets a custom bot service for testing.
func (r *BotRunner) WithBotService(service *bot.Service) *BotRunner {
	r.botService = service
	return r
}

// WithLLMProvider sets a custom LLM provider for testing.
func (r *BotRunner) WithLLMProvider(provider core.LLM) *BotRunner {
	r.llmProvider = provider
	return r
}

// RunBot initializes and runs the Telegram bot with the provided configuration.
// It sets up the AI provider, creates necessary services, and starts the bot.
func (r *BotRunner) RunBot(ctx context.Context, cfg *Config) error {
	var botService *bot.Service
	if r.botService != nil {
		botService = r.botService
	} else {
		var llmProvider core.LLM
		var err error

		if r.llmProvider != nil {
			llmProvider = r.llmProvider
		} else {
			llmProvider, err = anthropic.New(cfg.AI)
			if err != nil {
				return fmt.Errorf("failed to initialize Anthropic provider: %w", err)
			}
		}

		aiService := core.NewAIService(llmProvider)
		botService, err = r.createService(&cfg.Bot, aiService)
		if err != nil {
			return fmt.Errorf("failed to create bot service: %w", err)
		}
	}

	return botService.Run(ctx)
}
