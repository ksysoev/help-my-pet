package anthropic

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// LLMCaller defines the interface for LLM interactions
type LLMCaller interface {
	Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error)
}

// LLMAdapter adapts the langchaingo LLM to our core.LLM interface
type LLMAdapter struct {
	llm llms.LLM
}

func NewLLMAdapter(llm llms.LLM) LLMCaller {
	return &LLMAdapter{llm: llm}
}

func (a *LLMAdapter) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return a.llm.Call(ctx, prompt, options...)
}
