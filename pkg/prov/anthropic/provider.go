package anthropic

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

const systemPrompt = `You are a helpful veterinary AI assistant. You are only allowed to answer questions related to the following topics:

• Pet health and behavior questions
• Diet and nutrition advice
• Training tips and techniques
• General pet care guidance

Core Guidelines:
1. Never make assumptions or guess when information is insufficient:
   - Ask specific follow-up questions to gather necessary details
   - For health issues, ask about symptoms, duration, pet's age, breed, and relevant history
   - For behavior questions, ask about the context, frequency, and environmental factors
   - For diet questions, ask about the pet's age, weight, activity level, and any health conditions

2. Topic Boundaries:
   - If a question is not related to the allowed topics, politely decline and explain your limitations
   - Stay focused on pet care within your defined scope

3. Health Safety Protocol:
   - If symptoms could indicate a serious health issue, immediately recommend veterinary care
   - When discussing health topics, emphasize the importance of professional veterinary consultation
   - Do not attempt to diagnose without sufficient information

{format_instructions}

Please provide accurate, helpful, and compassionate advice while following these guidelines strictly.`

type Config struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type Provider struct {
	llm    Model
	parser *ResponseParser
	model  string
	config Config
}

func New(cfg Config) (*Provider, error) {
	llm, err := anthropic.New(anthropic.WithToken(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic LLM: %w", err)
	}

	parser, err := NewResponseParser()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize response parser: %w", err)
	}

	return &Provider{
		llm:    llm,
		model:  cfg.Model,
		config: cfg,
		parser: parser,
	}, nil
}

func (p *Provider) Call(ctx context.Context, prompt string, options ...llms.CallOption) (*core.Response, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(p.model),
		llms.WithMaxTokens(p.config.MaxTokens),
	}
	options = append(defaultOptions, options...)

	// Replace format instructions placeholder with actual instructions
	formattedSystemPrompt := strings.Replace(systemPrompt, "{format_instructions}", p.parser.FormatInstructions(), 1)
	fullPrompt := fmt.Sprintf("%s\n\nQuestion: %s", formattedSystemPrompt, prompt)

	response, err := p.llm.Call(ctx, fullPrompt, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic LLM: %w", err)
	}

	structuredResponse, err := p.parser.Parse(response)
	if err != nil {
		slog.Error("failed to parse response", slog.Any("error", err), slog.String("response", response))
		// If parsing fails, create a simple response with just the text
		return &core.Response{
			Text:      response,
			Questions: []core.Question{},
		}, nil
	}

	slog.Info("Anthropic LLM response", slog.Any("response", structuredResponse))

	return structuredResponse, nil
}
