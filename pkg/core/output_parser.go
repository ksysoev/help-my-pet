package core

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/tmc/langchaingo/outputparser"
)

// ResponseParser is a custom output parser for our Response type
type ResponseParser struct {
	parser outputparser.Structured
}

// NewResponseParser creates a new ResponseParser instance
func NewResponseParser() (*ResponseParser, error) {
	// Define the schema for our response
	schema := []outputparser.ResponseSchema{
		{
			Name:        "text",
			Description: "The main response text providing pet care advice",
		},
		{
			Name:        "questions",
			Description: "Array of follow-up questions with optional predefined answers",
		},
	}

	// Initialize the Structured parser with our schema
	parser := outputparser.NewStructured(schema)

	return &ResponseParser{
		parser: parser,
	}, nil
}

// FormatInstructions returns the format instructions for the LLM
func (p *ResponseParser) FormatInstructions() string {
	return p.parser.GetFormatInstructions()
}

// Parse parses the LLM output into our Response struct
func (p *ResponseParser) Parse(text string) (*Response, error) {
	var response Response
	if strings.HasPrefix(text, "```json") && strings.HasSuffix(text, "```") {
		// Extract JSON from markdown code block
		text = strings.TrimPrefix(text, "```json\n")
		text = strings.TrimSuffix(text, "\n```")
	}

	if err := json.Unmarshal([]byte(text), &response); err != nil {
		slog.Error("failed to parse response", slog.Any("error", err), slog.String("response", text))

		return &Response{
			Text:      text,
			Questions: []Question{},
		}, nil
	}

	// If inner parsing fails or text is not JSON, return the outer response
	return &response, nil
}
