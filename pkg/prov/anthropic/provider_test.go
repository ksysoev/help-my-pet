package anthropic

import (
	"context"
	"fmt"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tmc/langchaingo/llms"
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

	tests := []struct {
		name       string
		prompt     string
		setupMock  func(t *testing.T) *Provider
		wantResult string
		wantErr    bool
	}{
		{
			name:   "successful call",
			prompt: "test prompt",
			setupMock: func(t *testing.T) *Provider {
				mockLLM := core.NewMockLLM(t)
				expectedPrompt := fmt.Sprintf("%s\n\nQuestion: %s", systemPrompt, "test prompt")
				mockLLM.EXPECT().
					Call(ctx, expectedPrompt,
						mock.MatchedBy(func(opt llms.CallOption) bool { return true }),
						mock.MatchedBy(func(opt llms.CallOption) bool { return true })).
					Return("test response", nil)

				return &Provider{
					llm:    mockLLM,
					model:  config.Model,
					config: config,
				}
			},
			wantResult: "test response",
			wantErr:    false,
		},
		{
			name:   "error from LLM",
			prompt: "test prompt",
			setupMock: func(t *testing.T) *Provider {
				mockLLM := core.NewMockLLM(t)
				expectedPrompt := fmt.Sprintf("%s\n\nQuestion: %s", systemPrompt, "test prompt")
				mockLLM.EXPECT().
					Call(ctx, expectedPrompt,
						mock.MatchedBy(func(opt llms.CallOption) bool { return true }),
						mock.MatchedBy(func(opt llms.CallOption) bool { return true })).
					Return("", assert.AnError)

				return &Provider{
					llm:    mockLLM,
					model:  config.Model,
					config: config,
				}
			},
			wantResult: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupMock(t)
			result, err := provider.Call(ctx, tt.prompt)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
