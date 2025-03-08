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
// MediaModel specifies the model used for media analysis, such as "haiku".
// MaxTokens sets the maximum number of tokens allowed per request, controlling output length and cost.
type Config struct {
	APIKey     string `mapstructure:"api_key"`
	Model      string `mapstructure:"model"`
	MediaModel string `mapstructure:"media_model"`
	MaxTokens  int    `mapstructure:"max_tokens"`
}

// Provider encapsulates the LLM model, response parser, and configuration for handling language model interactions.
// It facilitates seamless interaction with the underlying LLM by combining format instructions, user inputs, and system details.
// The response parser is responsible for converting the raw LLM output into a structured format as defined by the ResponseParser.
// Config contains essential settings such as API keys, model type, and token limits, which are used to initialize the provider.
type Provider struct {
	llm        Model
	mediaModel Model
	config     Config
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

	mediaModel, err := newAnthropicModel(cfg.APIKey, cfg.MediaModel, cfg.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic media model: %w", err)
	}

	return &Provider{
		llm:        llm,
		mediaModel: mediaModel,
		config:     cfg,
	}, nil
}

// Analyze processes a request and associated images using the LLM, returning a formatted result or an error.
// It sends the request combined with system prompts to the LLM, parses the response, and handles errors if the API call or parsing fails.
// ctx is the context for managing request lifecycle; request is the input query; imgs represents associated images to be analyzed.
// Returns a structured LLMResult containing the analysis or an error if the LLM call or response parsing fails.
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

// Report generates a formatted analysis based on the provided request using the LLM and parses the response into a structured result.
// It sends a system prompt combined with the user's input to the LLM and handles errors during the API call or parsing process.
// ctx is the context for managing the request lifecycle; request is the input query to be analyzed.
// Returns a structured LLMResult containing the analysis or an error if the LLM call or response parsing fails.
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
