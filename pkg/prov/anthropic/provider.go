package anthropic

import (
	"context"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

const systemPrompt = `You are a helpful veterinary AI assistant. You are only allowed to answer questions related to the following topics:

• Pet health and behavior questions
• Diet and nutrition advice
• Training tips and techniques
• General pet care guidance

If a question is not related to these topics, you must politely decline to answer and explain that you can only assist with pet-related questions within the allowed topics.

You must be vigilant about potential health risks:
1. If you detect any symptoms or situations that could indicate a serious health issue, you MUST strongly recommend consulting a veterinarian immediately.
2. When discussing health-related topics, always err on the side of caution and emphasize the importance of professional veterinary care.

Please provide accurate, helpful, and compassionate advice while staying strictly within these guidelines.`

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
