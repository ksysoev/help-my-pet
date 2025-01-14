package anthropic

import (
	"context"
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
				assert.Equal(t, tt.config.Model, provider.model)
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
				mockAdapter := &MockLLMAdapter{}
				mockAdapter.Test(t)
				mockAdapter.On("Call", ctx, mock.Anything, mock.Anything).
					Return(`{"text": "test response", "questions": [{"text": "follow up?"}]}`, nil)

				parser, err := core.NewResponseParser()
				assert.NoError(t, err)

				return &Provider{
					caller: mockAdapter,
					model:  config.Model,
					config: config,
					parser: parser,
				}
			},
		},
		{
			wantErr: false,
			name:    "successful call with non-JSON response",
			prompt:  "test prompt",
			wantResult: &core.Response{
				Text:      "test response",
				Questions: []core.Question{},
			},
			setupMock: func(t *testing.T) *Provider {
				mockAdapter := &MockLLMAdapter{}
				mockAdapter.Test(t)
				mockAdapter.On("Call", ctx, mock.Anything, mock.Anything).
					Return("test response", nil)

				parser, err := core.NewResponseParser()
				assert.NoError(t, err)

				return &Provider{
					caller: mockAdapter,
					model:  config.Model,
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
				mockAdapter := &MockLLMAdapter{}
				mockAdapter.Test(t)
				mockAdapter.On("Call", ctx, mock.Anything, mock.Anything).
					Return("", assert.AnError)

				parser, err := core.NewResponseParser()
				assert.NoError(t, err)

				return &Provider{
					caller: mockAdapter,
					model:  config.Model,
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
