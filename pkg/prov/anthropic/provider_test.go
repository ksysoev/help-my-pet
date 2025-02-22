package anthropic

import (
	"context"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/stretchr/testify/assert"
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

func TestProvider_Anylyze(t *testing.T) {
	ctx := context.Background()
	config := Config{
		APIKey:    "test-api-key",
		Model:     "claude-2",
		MaxTokens: 1000,
	}

	type testCase struct {
		wantResult *message.LLMResult
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
			wantResult: &message.LLMResult{
				Text: "test response",
				Questions: []message.Question{
					{Text: "follow up?"},
				},
			},
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)

				p := &Provider{
					llm:    mockModel,
					config: config,
				}

				mockModel.EXPECT().Analyze(ctx, p.systemInfo()+"test prompt", []*message.Image(nil)).
					Return(`{"text": "test response", "questions": [{"text": "follow up?"}]}`, nil)

				return p
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

				formatInstructions := parser.FormatInstructions()

				p := &Provider{
					llm:    mockModel,
					config: config,
					parser: parser,
				}

				mockModel.EXPECT().Call(ctx, formatInstructions, p.systemInfo()+"test prompt", []*message.Image(nil)).
					Return("test response", nil)

				return p
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

				formatInstructions := parser.FormatInstructions()

				p := &Provider{
					llm:    mockModel,
					config: config,
					parser: parser,
				}

				mockModel.EXPECT().Call(ctx, formatInstructions, p.systemInfo()+"test prompt", []*message.Image(nil)).
					Return("", assert.AnError)

				return p
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupMock(t)
			result, err := provider.Analyze(ctx, tt.prompt, nil)

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
