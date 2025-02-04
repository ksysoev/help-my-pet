package message

import "github.com/ksysoev/help-my-pet/pkg/core/conversation"

// LLMResult represents a structured response from the LLM
type LLMResult struct {
	Text      string                  `json:"text"`      // Main response text
	Questions []conversation.Question `json:"questions"` // Optional follow-up questions
}
