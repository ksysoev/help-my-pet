package anthropic

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

	// Get formatted system prompt with parser instructions
	parser, err := NewResponseParser()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize response parser: %w", err)
	}
	formattedSystemPrompt := strings.Replace(systemPrompt, "{format_instructions}", parser.FormatInstructions(), 1)

	llm, err := newAnthropicModel(cfg.APIKey, cfg.Model, cfg.MaxTokens, formattedSystemPrompt)
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
func (p *Provider) Call(ctx context.Context, prompt string) (*core.Response, error) {
	fullPrompt := fmt.Sprintf("%s\n\nQuestion: %s", p.parser.FormatInstructions(), prompt)

	slog.Debug("Anthropic LLM full prompt", slog.String("prompt", fullPrompt))

	response, err := p.llm.Call(ctx, fullPrompt)
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
