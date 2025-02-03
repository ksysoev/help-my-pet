package anthropic_test

import (
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
	"github.com/ksysoev/help-my-pet/pkg/prov/anthropic"
	"github.com/stretchr/testify/assert"
)

func TestResponseParser_Parse(t *testing.T) {
	tests := []struct {
		expected *core.LLMResult
		name     string
		input    string
		wantErr  bool
	}{
		{
			name: "valid JSON response",
			input: `{
				"text": "Your cat needs regular grooming.",
				"questions": [
					{
						"text": "How often do you brush your cat?",
						"answers": ["Daily", "Weekly", "Monthly"]
					}
				]
			}`,
			expected: &core.LLMResult{
				Text: "Your cat needs regular grooming.",
				Questions: []conversation.Question{
					{
						Text:    "How often do you brush your cat?",
						Answers: []string{"Daily", "Weekly", "Monthly"},
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "JSON response in markdown code block",
			input: "```json\n{\n\t\"text\": \"Feed your cat twice daily.\",\n\t\"questions\": []\n}\n```",
			expected: &core.LLMResult{
				Text:      "Feed your cat twice daily.",
				Questions: []conversation.Question{},
			},
			wantErr: false,
		},
		{
			name: "unescaped newlines in string literals",
			input: `{
				"text": "Line 1
Line 2",
				"questions": [
					{
						"text": "Question with
newline",
						"answers": ["Answer with
newline", "Normal answer"]
					}
				]
			}`,
			expected: &core.LLMResult{
				Text: "Line 1\nLine 2",
				Questions: []conversation.Question{
					{
						Text:    "Question with\nnewline",
						Answers: []string{"Answer with\nnewline", "Normal answer"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "properly escaped characters",
			input: `{
				"text": "Tab\there \"quoted\" text\\with\\backslashes",
				"questions": [
					{
						"text": "Question with \"quotes\"",
						"answers": ["Answer with \t tab"]
					}
				]
			}`,
			expected: &core.LLMResult{
				Text: "Tab\there \"quoted\" text\\with\\backslashes",
				Questions: []conversation.Question{
					{
						Text:    "Question with \"quotes\"",
						Answers: []string{"Answer with \t tab"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unterminated string",
			input: `{
				"text": "Unterminated string,
				"questions": []
			}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "mixed content with newlines",
			input: `
				{
					"text": "First line
					second line
					third line",
					"questions": [
						{
							"text": "Question spanning
							multiple lines?",
							"answers": []
						}
					]
				}
			`,
			expected: &core.LLMResult{
				Text: "First line\n\t\t\t\t\tsecond line\n\t\t\t\t\tthird line",
				Questions: []conversation.Question{
					{
						Text:    "Question spanning\n\t\t\t\t\t\t\tmultiple lines?",
						Answers: []string{},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON response",
			input:    "This is not JSON",
			expected: nil,
			wantErr:  true,
		},
		{
			name: "empty questions array",
			input: `{
				"text": "Simple advice without questions.",
				"questions": []
			}`,
			expected: &core.LLMResult{
				Text:      "Simple advice without questions.",
				Questions: []conversation.Question{},
			},
			wantErr: false,
		},
		{
			name: "multiple questions with and without answers",
			input: `{
				"text": "Here's your pet advice.",
				"questions": [
					{
						"text": "What type of pet do you have?",
						"answers": ["Dog", "Cat", "Bird", "Other"]
					},
					{
						"text": "Describe your pet's behavior",
						"answers": []
					}
				]
			}`,
			expected: &core.LLMResult{
				Text: "Here's your pet advice.",
				Questions: []conversation.Question{
					{
						Text:    "What type of pet do you have?",
						Answers: []string{"Dog", "Cat", "Bird", "Other"},
					},
					{
						Text:    "Describe your pet's behavior",
						Answers: []string{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "JSON with comments-like content",
			input: `{
				"text": "Text with // comment",
				"questions": [
					{
						"text": "Question with /* comment */",
						"answers": ["// Answer with comment"]
					}
				]
			}`,
			expected: &core.LLMResult{
				Text: "Text with // comment",
				Questions: []conversation.Question{
					{
						Text:    "Question with /* comment */",
						Answers: []string{"// Answer with comment"},
					},
				},
			},
			wantErr: false,
		},
	}

	parser, err := anthropic.NewResponseParser()
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResponseParser_FormatInstructions(t *testing.T) {
	parser, err := anthropic.NewResponseParser()
	assert.NoError(t, err)

	instructions := parser.FormatInstructions()
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "text")
	assert.Contains(t, instructions, "questions")
	assert.Contains(t, instructions, "The main response text providing pet care advice")
}

func TestNewResponseParser(t *testing.T) {
	parser, err := anthropic.NewResponseParser()
	assert.NoError(t, err)
	assert.NotNil(t, parser)
}
