package anthropic

import (
	"context"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

const systemPrompt = `You are a helpful veterinary AI assistant. Please provide accurate, helpful, and compassionate advice for pet-related questions. If the question involves a serious medical condition, always recommend consulting with a veterinarian.`

type Config struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type Provider struct {
	llm    core.LLM
	model  string
	config Config
}

func New(cfg Config) (*Provider, error) {
	llm, err := anthropic.New(anthropic.WithToken(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic LLM: %w", err)
	}

	return &Provider{
		llm:    llm,
		model:  cfg.Model,
		config: cfg,
	}, nil
}

func (p *Provider) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(p.model),
		llms.WithMaxTokens(p.config.MaxTokens),
	}
	options = append(defaultOptions, options...)

	fullPrompt := fmt.Sprintf("%s\n\nQuestion: %s", systemPrompt, prompt)
	return p.llm.Call(ctx, fullPrompt, options...)
}
