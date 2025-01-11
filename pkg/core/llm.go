package core

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// LLM interface represents the language model capabilities
type LLM interface {
	Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error)
}
