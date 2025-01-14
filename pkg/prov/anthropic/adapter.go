package anthropic

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// LLMCaller defines the interface for LLM interactions
type LLMCaller interface {
	Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error)
}

// LLMAdapter adapts the langchaingo Model to our core.LLM interface
type LLMAdapter struct {
	model llms.Model
}

func NewLLMAdapter(model llms.Model) LLMCaller {
	return &LLMAdapter{model: model}
}

func (a *LLMAdapter) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	response, err := llms.GenerateFromSinglePrompt(ctx, a.model, prompt, options...)
	if err != nil {
		return "", err
	}
	return response, nil
}
