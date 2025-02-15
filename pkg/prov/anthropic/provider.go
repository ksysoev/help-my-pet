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
	parser *ResponseParser
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

// Call performs a request to the underlying language model (LLM) with the provided context and prompt.
// It combines internal format instructions and system information with the user prompt to create the full input.
// If the request succeeds, the LLM response is parsed into a structured format.
// ctx is the request context, prompt is the user's input.
// Returns a structured LLMResult containing the response and any follow-up questions, or an error if the call fails or the response cannot be parsed.
func (p *Provider) Call(ctx context.Context, prompt string, imgs []*message.Image) (*message.LLMResult, error) {
	formatInstructions := p.parser.FormatInstructions()

	slog.DebugContext(ctx, "Anthropic LLM call",
		slog.String("format_instructions", formatInstructions),
		slog.String("question", prompt))

	if len(imgs) > 0 {
		imgDescription, err := p.llm.DescribeImages(ctx, imgs)
		if err != nil {
			return nil, fmt.Errorf("failed to describe images: %w", err)
		}

		prompt = fmt.Sprintf("Image Analysis:\n%s\n\n%s", imgDescription, prompt)
	}

	response, err := p.llm.Call(ctx, formatInstructions, p.systemInfo()+prompt, []*message.Image{})
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

// systemInfo retrieves and formats basic system information, including the current date in YYYY-MM-DD format.
// It generates a string containing the system details to be included in LLM calls.
// Returns a string representing the system information.
func (p *Provider) systemInfo() string {
	return fmt.Sprintf("System Information:\n Current date in format YYYY-MM-DD: %s\n\n", time.Now().Format("2006-01-02"))
}
