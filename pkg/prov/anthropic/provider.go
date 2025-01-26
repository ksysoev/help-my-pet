package anthropic

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/tmc/langchaingo/llms"
)

// anthropicModel adapts the Anthropic API to the Model interface
type anthropicModel struct {
	client     *anthropic.Client
	modelID    string
	maxTokens  int
	systemText string
}

func newAnthropicModel(apiKey string, modelID string, maxTokens int, systemPrompt string) (*anthropicModel, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	if client == nil {
		return nil, fmt.Errorf("failed to create Anthropic client")
	}

	return &anthropicModel{
		client:     client,
		modelID:    modelID,
		maxTokens:  maxTokens,
		systemText: systemPrompt,
	}, nil
}

func (m *anthropicModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	fullPrompt := strings.Replace(systemPrompt, "{format_instructions}", m.systemText, 1) + "\n\nQuestion: " + prompt

	message, err := m.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(m.modelID),
		MaxTokens: anthropic.F(int64(m.maxTokens)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(fullPrompt)),
		}),
	})
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic API")
	}

	return message.Content[0].Text, nil
}

func (m *anthropicModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return nil, fmt.Errorf("GenerateContent not implemented")
}

const systemPrompt = `You are a helpful veterinary AI assistant. You MUST detect the language of the user's question and respond in the SAME language. You are only allowed to answer questions related to the following topics:

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
   - If symptoms could indicate a serious health issue, recommend veterinary care
   - When discussing health topics, recomend professional veterinary consultation
   - Do not attempt to diagnose without sufficient information

{format_instructions}

Please provide accurate, helpful, and compassionate advice while following these guidelines strictly.`

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
func (p *Provider) Call(ctx context.Context, prompt string, options ...llms.CallOption) (*core.Response, error) {
	// Replace format instructions placeholder with actual instructions
	formattedSystemPrompt := strings.Replace(systemPrompt, "{format_instructions}", p.parser.FormatInstructions(), 1)
	fullPrompt := fmt.Sprintf("%s\n\nQuestion: %s", formattedSystemPrompt, prompt)

	slog.Debug("Anthropic LLM full prompt", slog.String("prompt", fullPrompt))

	response, err := p.llm.Call(ctx, fullPrompt, options...)
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
