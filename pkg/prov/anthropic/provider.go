package anthropic

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ksysoev/help-my-pet/pkg/core"
)

// Config holds the configuration for the Anthropic provider
type Config struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

// Provider implements the core.LLM interface
type Provider struct {
	llm    Model
	parser *ResponseParser
	config Config
}

// New creates a new Anthropic provider instance
func New(cfg Config) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	parser, err := NewResponseParser()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize response parser: %w", err)
	}

	llm, err := newAnthropicModel(cfg.APIKey, cfg.Model, cfg.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic model: %w", err)
	}

	return &Provider{
		llm:    llm,
		config: cfg,
		parser: parser,
	}, nil
}

// Call sends a message to the Anthropic API and returns the structured response
func (p *Provider) Call(ctx context.Context, prompt string) (*core.LLMResult, error) {
	formatInstructions := p.parser.FormatInstructions()

	slog.DebugContext(ctx, "Anthropic LLM call",
		slog.String("format_instructions", formatInstructions),
		slog.String("question", prompt))

	response, err := p.llm.Call(ctx, formatInstructions, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	structuredResponse, err := p.parser.Parse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	slog.Debug("Anthropic LLM response", slog.Any("response", structuredResponse))

	return structuredResponse, nil
}
