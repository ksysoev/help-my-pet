package core

import (
	"context"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
)

// Response represents a structured response from the LLM
type Response struct {
	Text      string                  `json:"text"`      // Main response text
	Questions []conversation.Question `json:"questions"` // Optional follow-up questions
}

// LLM interface represents the language model capabilities
type LLM interface {
	// Call sends a prompt to the LLM and returns a structured response
	Call(ctx context.Context, prompt string) (*Response, error)
}
