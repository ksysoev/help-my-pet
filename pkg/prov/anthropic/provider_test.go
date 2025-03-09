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
		images     []*message.Image
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

				mockModel.EXPECT().Call(ctx, analyzePrompt+analyzeOutput, p.systemInfo()+"test prompt", []*message.Image(nil)).
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

				p := &Provider{
					llm:    mockModel,
					config: config,
				}

				mockModel.EXPECT().Call(ctx, analyzePrompt+analyzeOutput, p.systemInfo()+"test prompt", []*message.Image(nil)).
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

				p := &Provider{
					llm:    mockModel,
					config: config,
				}

				mockModel.EXPECT().Call(ctx, analyzePrompt+analyzeOutput, p.systemInfo()+"test prompt", []*message.Image(nil)).
					Return("", assert.AnError)

				return p
			},
		},
		{
			wantErr: false,
			name:    "successful call with media description",
			prompt:  "test prompt",
			wantResult: &message.LLMResult{
				Text:  "test response",
				Media: "media description",
			},
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)
				mockMediaModel := NewMockModel(t)

				p := &Provider{
					llm:        mockModel,
					mediaModel: mockMediaModel,
					config:     config,
				}

				mockMediaModel.EXPECT().Call(ctx, mediaExtractionPrompt, mediaOutputFormat, []*message.Image{
					{
						MIME: "image/jpeg",
						Data: "base64-encoded-image",
					},
				}).Return("media description", nil)

				mockModel.EXPECT().Call(ctx, analyzePrompt+analyzeOutput, p.systemInfo()+"test prompt\n\nMedia content:\nmedia description", []*message.Image(nil)).
					Return(`{"text": "test response"}`, nil)

				return p
			},
			images: []*message.Image{
				{
					MIME: "image/jpeg",
					Data: "base64-encoded-image",
				},
			},
		},
		{
			wantErr:    true,
			name:       "error from media model",
			prompt:     "test prompt",
			wantResult: nil,
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)
				mockMediaModel := NewMockModel(t)

				p := &Provider{
					llm:        mockModel,
					mediaModel: mockMediaModel,
					config:     config,
				}

				mockMediaModel.EXPECT().Call(ctx, mediaExtractionPrompt, mediaOutputFormat, []*message.Image{
					{
						MIME: "image/jpeg",
						Data: "base64-encoded-image",
					},
				}).Return("", assert.AnError)

				return p
			},
			images: []*message.Image{
				{
					MIME: "image/jpeg",
					Data: "base64-encoded-image",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupMock(t)
			result, err := provider.Analyze(ctx, tt.prompt, tt.images)

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

func TestProvider_Report(t *testing.T) {
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
		request    string
		wantErr    bool
	}

	tests := []testCase{
		{
			wantErr: false,
			name:    "successful call with valid JSON response",
			request: "report request",
			wantResult: &message.LLMResult{
				Text: "test report response",
				Questions: []message.Question{
					{Text: "additional details?"},
				},
			},
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)

				p := &Provider{
					llm:    mockModel,
					config: config,
				}

				mockModel.EXPECT().Call(ctx, reportPrompt+reportOutput, p.systemInfo()+"report request", []*message.Image(nil)).
					Return(`{"text": "test report response", "questions": [{"text": "additional details?"}]}`, nil)

				return p
			},
		},
		{
			wantErr:    true,
			name:       "invalid JSON response from LLM",
			request:    "report request",
			wantResult: nil,
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)

				p := &Provider{
					llm:    mockModel,
					config: config,
				}

				mockModel.EXPECT().Call(ctx, reportPrompt+reportOutput, p.systemInfo()+"report request", []*message.Image(nil)).
					Return("invalid response", nil)

				return p
			},
		},
		{
			wantErr:    true,
			name:       "error from LLM",
			request:    "report request",
			wantResult: nil,
			setupMock: func(t *testing.T) *Provider {
				mockModel := NewMockModel(t)

				p := &Provider{
					llm:    mockModel,
					config: config,
				}

				mockModel.EXPECT().Call(ctx, reportPrompt+reportOutput, p.systemInfo()+"report request", []*message.Image(nil)).
					Return("", assert.AnError)

				return p
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupMock(t)
			result, err := provider.Report(ctx, tt.request)

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
