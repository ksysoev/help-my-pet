package core

import (
	"context"
	"encoding/json"

	"github.com/tmc/langchaingo/llms"
)

// Question represents a follow-up question with optional predefined answers
type Question struct {
	Text    string   `json:"text"`
	Answers []string `json:"answers,omitempty"`
}

// Response represents a structured response from the LLM
type Response struct {
	Text      string     `json:"text"`      // Main response text
	Questions []Question `json:"questions"` // Optional follow-up questions
}

// LLM interface represents the language model capabilities
type LLM interface {
	// Call sends a prompt to the LLM and returns a structured response
	Call(ctx context.Context, prompt string, options ...llms.CallOption) (*Response, error)
}

// ParseResponse attempts to parse a JSON string into a Response struct
func ParseResponse(jsonStr string) (*Response, error) {
	var response Response
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, err
	}
	return &response, nil
}
