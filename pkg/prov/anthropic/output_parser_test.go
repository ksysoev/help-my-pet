package anthropic_test

import (
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/prov/anthropic"
	"github.com/stretchr/testify/assert"
)

func TestResponseParser_Parse(t *testing.T) {
	tests := []struct {
		expected *core.Response
		input    string
		name     string
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
			expected: &core.Response{
				Text: "Your cat needs regular grooming.",
				Questions: []core.Question{
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
			expected: &core.Response{
				Text:      "Feed your cat twice daily.",
				Questions: []core.Question{},
			},
			wantErr: false,
		},
		{
			name:  "invalid JSON response",
			input: "This is not JSON",
			expected: &core.Response{
				Text:      "This is not JSON",
				Questions: []core.Question{},
			},
			wantErr: false,
		},
		{
			name: "empty questions array",
			input: `{
				"text": "Simple advice without questions.",
				"questions": []
			}`,
			expected: &core.Response{
				Text:      "Simple advice without questions.",
				Questions: []core.Question{},
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
			expected: &core.Response{
				Text: "Here's your pet advice.",
				Questions: []core.Question{
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
