package anthropic

import (
	"context"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

type Config struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

type Provider struct {
	llm   core.LLM
	model string
}

func New(cfg Config) (*Provider, error) {
	llm, err := anthropic.New(anthropic.WithToken(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic LLM: %w", err)
	}

	return &Provider{
		llm:   llm,
		model: cfg.Model,
	}, nil
}

func (p *Provider) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	options = append(options, llms.WithModel(p.model))
	return p.llm.Call(ctx, prompt, options...)
}
