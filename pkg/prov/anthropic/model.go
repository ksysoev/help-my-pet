package anthropic

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// Model defines the interface for LLM interactions
type Model interface {
	Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error)
	GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error)
}
