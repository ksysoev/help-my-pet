package anthropic

import (
	"context"
	"strings"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: Config{
				APIKey:    "test-api-key",
				Model:     "claude-2",
				MaxTokens: 1000,
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: Config{
				APIKey:    "",
				Model:     "claude-2",
				MaxTokens: 1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, tt.config.MaxTokens, provider.config.MaxTokens)
			}
		})
	}
}

func TestProvider_Call(t *testing.T) {
	ctx := context.Background()
	config := Config{
		APIKey:    "test-api-key",
		Model:     "claude-2",
		MaxTokens: 1000,
	}

	type testCase struct {
		wantResult *core.Response
		setupMock  func(t *testing.T) *Provider
		name       string
		prompt     string
		wantErr    bool
	}

	tests := []testCase{
		{
			wantErr: false,
			name:    "successful call with valid JSON response",
			prompt:  "test prompt",
			wantResult: &core.Response{
				Text: "test response",
				Questions: []core.Question{
					{Text: "follow up?"},
				},
			},
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)
				parser, err := NewResponseParser()
				assert.NoError(t, err)

				fullPrompt := strings.Replace(systemPrompt, "{format_instructions}", parser.FormatInstructions(), 1) + "\n\nQuestion: test prompt"

				mockModel.EXPECT().Call(ctx, fullPrompt, mock.Anything).
					Return(`{"text": "test response", "questions": [{"text": "follow up?"}]}`, nil)

				return &Provider{
					llm:    mockModel,
					config: config,
					parser: parser,
				}
			},
		},
		{
			wantErr:    true,
			name:       "invalid JSON response from LLM",
			prompt:     "test prompt",
			wantResult: nil,
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)
				parser, err := NewResponseParser()
				assert.NoError(t, err)

				fullPrompt := strings.Replace(systemPrompt, "{format_instructions}", parser.FormatInstructions(), 1) + "\n\nQuestion: test prompt"

				mockModel.EXPECT().Call(ctx, fullPrompt, mock.Anything).
					Return("test response", nil)

				return &Provider{
					llm:    mockModel,
					config: config,
					parser: parser,
				}
			},
		},
		{
			wantErr:    true,
			name:       "error from LLM",
			prompt:     "test prompt",
			wantResult: nil,
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)
				parser, err := NewResponseParser()
				assert.NoError(t, err)

				fullPrompt := strings.Replace(systemPrompt, "{format_instructions}", parser.FormatInstructions(), 1) + "\n\nQuestion: test prompt"

				mockModel.EXPECT().Call(ctx, fullPrompt, mock.Anything).
					Return("", assert.AnError)

				return &Provider{
					llm:    mockModel,
					config: config,
					parser: parser,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupMock(t)
			result, err := provider.Call(ctx, tt.prompt)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
