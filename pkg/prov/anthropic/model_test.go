package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnthropicModel(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		modelID   string
		maxTokens int
		thinking  bool
		wantErr   bool
	}{
		{
			name:      "creates model with thinking enabled",
			apiKey:    "test-key",
			modelID:   "claude-sonnet-4-6",
			maxTokens: 16000,
			thinking:  true,
			wantErr:   false,
		},
		{
			name:      "creates model with thinking disabled",
			apiKey:    "test-key",
			modelID:   "claude-haiku-4-5",
			maxTokens: 4096,
			thinking:  false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := newAnthropicModel(tt.apiKey, tt.modelID, tt.maxTokens, tt.thinking)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, model)
			} else {
				require.NoError(t, err)
				require.NotNil(t, model)
				assert.Equal(t, tt.modelID, model.modelID)
				assert.Equal(t, tt.maxTokens, model.maxTokens)
				assert.Equal(t, tt.thinking, model.thinking)
			}
		})
	}
}

// fakeAPIResponse is the minimal JSON body the Anthropic Messages API returns.
type fakeAPIResponse struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	Model        string        `json:"model"`
	StopReason   string        `json:"stop_reason"`
	StopSequence *string       `json:"stop_sequence"`
	Content      []interface{} `json:"content"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// newFakeServer starts an httptest server that always returns the provided response body.
func newFakeServer(t *testing.T, resp fakeAPIResponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestAnthropicModel_CallMaxTokensError(t *testing.T) {
	// Build a fake server that returns stop_reason = "max_tokens" to trigger
	// the truncation-detection branch in anthropicModel.Call.
	srv := newFakeServer(t, fakeAPIResponse{
		ID:         "msg_test",
		Type:       "message",
		Role:       "assistant",
		Model:      "claude-sonnet-4-6",
		StopReason: "max_tokens",
		Content: []interface{}{
			map[string]string{"type": "text", "text": `{"truncated`},
		},
	})
	defer srv.Close()

	client := anthropicsdk.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(srv.URL),
	)

	model := &anthropicModel{
		client:    client,
		modelID:   "claude-sonnet-4-6",
		maxTokens: 100,
		thinking:  false,
	}

	_, err := model.Call(context.Background(), "system prompt", "user question", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max_tokens")
	assert.Contains(t, err.Error(), "response truncated")
}
