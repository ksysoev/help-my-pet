package anthropic

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

// Config defines the configuration settings required for interacting with the language model API.
// It includes parameters for API authentication, model selection, and token usage limits.
// APIKey specifies the API key used for authenticating with the language model provider.
// Model identifies the specific language model to interact with, such as "claude-2".
// MaxTokens sets the maximum number of tokens allowed per request, controlling output length and cost.
type Config struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

// Provider encapsulates the LLM model, response parser, and configuration for handling language model interactions.
// It facilitates seamless interaction with the underlying LLM by combining format instructions, user inputs, and system details.
// The response parser is responsible for converting the raw LLM output into a structured format as defined by the ResponseParser.
// Config contains essential settings such as API keys, model type, and token limits, which are used to initialize the provider.
type Provider struct {
	llm    Model
	config Config
}

// New initializes and returns a new Provider instance based on the provided configuration.
// It validates the Config parameters, such as the presence of an API key, and sets up required components.
// cfg provides the configuration needed to initialize the provider, including API key, model, and max tokens.
// Returns a Provider instance for LLM interactions or an error if initialization fails due to invalid inputs or setup issues.
func New(cfg Config) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	llm, err := newAnthropicModel(cfg.APIKey, cfg.Model, cfg.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic model: %w", err)
	}

	return &Provider{
		llm:    llm,
		config: cfg,
	}, nil
}

func (p *Provider) Analyze(ctx context.Context, request string, imgs []*message.Image) (*message.LLMResult, error) {
	slog.DebugContext(ctx, "Anthropic LLM call", slog.String("question", request))

	parser := newAssistantResponseParser(analyzeOutput)

	systemPrompt := analyzePrompt + parser.FormatInstructions()

	response, err := p.llm.Call(ctx, systemPrompt, p.systemInfo()+request, imgs)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	slog.Debug("Anthropic LLM response", slog.Any("response", response))

	result, err := parser.Parse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result, nil
}

func (p *Provider) Report(ctx context.Context, request string) (*message.LLMResult, error) {
	slog.DebugContext(ctx, "Anthropic LLM call", slog.String("question", request))

	parser := newAssistantResponseParser(reportOutput)

	systemPrompt := reportPrompt + parser.FormatInstructions()

	response, err := p.llm.Call(ctx, systemPrompt, p.systemInfo()+request, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	slog.Debug("Anthropic LLM response", slog.Any("response", response))

	result, err := parser.Parse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result, nil
}

// systemInfo retrieves and formats basic system information, including the current date in YYYY-MM-DD format.
// It generates a string containing the system details to be included in LLM calls.
// Returns a string representing the system information.
func (p *Provider) systemInfo() string {
	return fmt.Sprintf("System Information:\n Current date in format YYYY-MM-DD: %s\n\n", time.Now().Format("2006-01-02"))
}
