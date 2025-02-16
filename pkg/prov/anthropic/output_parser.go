package anthropic

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

var (
	ErrInvalidJSON = errors.New("invalid JSON format in response")
	ErrEmptyText   = errors.New("response text is empty")
)

// ResponseParser is a custom output parser for our Response type
type ResponseParser struct{}

// NewResponseParser creates a new ResponseParser instance
func NewResponseParser() (*ResponseParser, error) {
	return &ResponseParser{}, nil
}

// FormatInstructions returns the format instructions for the LLM
func (p *ResponseParser) FormatInstructions() string {
	return `Return your response in JSON format with this structure:
{
  "thoughs": "Detailed description of your thoughts and reasoning behind the advice",
  "text": "The main response text providing pet care advice",
  "questions": [
    {
      "text": "Any follow-up questions to gather more information",
      "answers": ["Optional", "Array", "Of", "Predefined", "Answers"]
    }
  ],
  "media": "Optional detailed description of any media content provided if provided(photo, video, documents), this information may be used for future queries"
}

Note:
- The "thoughts" field you should use to explain in details your thought process and reasoning behing your response
- The "text" field is required and must contain your main advice or response
- The "questions" array is optional and can be empty if no follow-up questions are needed
- Each question must have a "text" field
- The "answers" field in questions is optional
- The "media" field is optional and can be used to save detailed media information for use in future queries. Focus on information for veterinarians and pet owners.

Example with no questions:
{
  "thoughts": "Based on the symptoms described, it sounds like your cat may have hairballs. Try brushing them daily and consider specialized hairball control food.",
  "text": "Based on the symptoms described, it sounds like your cat may have hairballs. Try brushing them daily and consider specialized hairball control food.",
  "questions": [],
  "media": "Size of hairballs is about 1 inch in diameter. It doesn't contain any blood or foreign objects."'
}

Example with questions:
{
  "thoughts": "Anxiety in dogs can be triggered by various factors. To provide proper advice, I need more information. Based on the symptoms described, your dog seems to be anxious in new environments or around loud noises.",
  "text": "To provide proper advice for your dog's anxiety, I need some more information.",
  "questions": [
    {
      "text": "How often does your dog show these symptoms?",
      "answers": ["Daily", "Weekly", "Only in specific situations"]
    },
    {
      "text": "Have you noticed any specific triggers?",
      "answers": []
    }
  ],
  "media": "Your dog seems to be anxious in new environments or around loud noises."
}`
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
func (p *ResponseParser) Parse(text string) (*message.LLMResult, error) {
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

	var response message.LLMResult
	decoder := json.NewDecoder(strings.NewReader(sanitized))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		slog.Error("failed to parse response",
			slog.Any("error", err),
			slog.String("original", text))
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if response.Text == "" {
		return nil, ErrEmptyText
	}

	return &response, nil
}
