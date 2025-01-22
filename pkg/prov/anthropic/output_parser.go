package anthropic

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/tmc/langchaingo/outputparser"
)

var (
	ErrInvalidJSON = errors.New("invalid JSON format in response")
	ErrEmptyText   = errors.New("response text is empty")
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
			Name: "questions",
			Description: `Array of follow-up questions. Each question should be an object with the following structure:
{
  "text": "The question text (required)",
  "answers": ["array", "of", "predefined", "answer", "options"] (optional)
}
Example:
[
  {
    "text": "How old is your cat?",
    "answers": []
  },
  {
    "text": "Is your cat indoor or outdoor?",
    "answers": ["Indoor", "Outdoor"]
  }
]`,
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

// sanitizeJSONStringLiterals processes a JSON string, escaping newlines only within string literals
func sanitizeJSONStringLiterals(input string) (string, error) {
	var result strings.Builder
	var inString bool
	var escaped bool

	for i := 0; i < len(input); i++ {
		char := input[i]

		// Handle escape sequences
		if escaped {
			result.WriteByte(char)
			escaped = false
			continue
		}

		// Check for escape character
		if char == '\\' {
			result.WriteByte(char)
			escaped = true
			continue
		}

		// Handle string boundaries
		if char == '"' {
			inString = !inString
			result.WriteByte(char)
			continue
		}

		// Handle newlines and tabs
		if inString && (char == '\n' || char == '\t') {
			if char == '\n' {
				result.WriteString("\\n")
			} else {
				result.WriteString("\\t")
			}
			continue
		}

		// Write all other characters as-is
		result.WriteByte(char)
	}

	if inString {
		return "", fmt.Errorf("unterminated string in JSON")
	}

	return result.String(), nil
}

// extractJSON attempts to extract JSON content from the text
func extractJSON(text string) string {
	// Try to find JSON between markdown code blocks
	if strings.HasPrefix(text, "```json") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSuffix(text, "```")
		return strings.TrimSpace(text)
	}

	// Try to find JSON between curly braces if no code block
	if !strings.HasPrefix(text, "{") {
		start := strings.Index(text, "{")
		if start != -1 {
			end := strings.LastIndex(text, "}")
			if end != -1 && end > start {
				return strings.TrimSpace(text[start : end+1])
			}
		}
	}

	return strings.TrimSpace(text)
}

// Parse parses the LLM output into our Response struct
func (p *ResponseParser) Parse(text string) (*core.Response, error) {
	if text == "" {
		return nil, ErrEmptyText
	}

	// Extract JSON content
	text = extractJSON(text)

	// Sanitize string literals
	sanitized, err := sanitizeJSONStringLiterals(text)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	var response core.Response
	decoder := json.NewDecoder(strings.NewReader(sanitized))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		slog.Error("failed to parse response",
			slog.Any("error", err),
			slog.String("original", text),
			slog.String("sanitized", sanitized))
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if response.Text == "" {
		return nil, ErrEmptyText
	}

	return &response, nil
}
