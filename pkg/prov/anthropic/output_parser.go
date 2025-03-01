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

// newAssistantResponseParser initializes and returns a ResponseParser for structuring assistant LLM responses.
// It uses the provided prompt as the format instructions for parsing.
// Returns a ResponseParser configured to parse responses into the message.LLMResult type.
func newAssistantResponseParser(prompt string) *ResponseParser[message.LLMResult] {
	return NewResponseParser[message.LLMResult](prompt)
}

// ResponseParser is a generic type that encapsulates functionality for parsing responses formatted by an LLM.
// It utilizes the provided format instructions to guide the parsing process and ensure the output conforms to a structured type.
type ResponseParser[T any] struct {
	instructions string
}

// NewResponseParser creates a new instance of ResponseParser with the provided format instructions.
// It returns a pointer to the initialized ResponseParser or an error if initialization fails.
// inst specifies the format instructions to guide the LLM response parsing.
// Returns a generic ResponseParser initialized with the given instructions and an error if inst is invalid or empty.
func NewResponseParser[T any](inst string) *ResponseParser[T] {
	return &ResponseParser[T]{
		instructions: inst,
	}
}

// FormatInstructions retrieves the format instructions used to guide the parsing process of the response.
// It returns a string containing the predefined format instructions stored within the ResponseParser.
func (p *ResponseParser[T]) FormatInstructions() string {
	return p.instructions
}

// Parse extracts and decodes JSON content from the provided text into the generic type T.
// It sanitizes input by escaping invalid string literals and removes surrounding markdown formatting if present.
// Returns a pointer to the parsed object of type T on success, or an error if the input is empty,
// the JSON is invalid, or the sanitization process fails.
func (p *ResponseParser[T]) Parse(text string) (*T, error) {
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

	var response T
	decoder := json.NewDecoder(strings.NewReader(sanitized))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		slog.Error("failed to parse response",
			slog.Any("error", err),
			slog.String("original", text))
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return &response, nil
}

// sanitizeJSONStringLiterals escapes invalid newline and tab characters within JSON string literals.
// It traverses the input string, ensuring proper escapes for special characters and handles escaped sequences correctly.
// Returns the sanitized JSON string with proper escaping, or an error if an unterminated string literal is detected.
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

// extractJSON extracts a JSON string from the provided text, handling both markdown code blocks and inline JSON structures.
// It removes surrounding markdown syntax such as "```json" and trims spaces around the detected JSON content.
// Returns the extracted JSON content as a string and ensures well-formed JSON is identified.
// Returns an empty string if no JSON-like structure is found in the input.
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

// cleanMarkdownTextFormat removes specific Markdown formatting from the input string to produce plain text output.
// It specifically strips bold syntax (`**`) from the text.
// Accepts text the input string containing Markdown-formatted text.
// Returns the cleaned string with bold syntax removed.
// Does not return an error as it performs a simple substitution.
func cleanMarkdownTextFormat(text string) string {
	// Remove bold markdown syntax
	text = strings.ReplaceAll(text, "**", "")

	return text
}
