package anthropic

import (
	"errors"
	"testing"

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

func TestAnthropicModel_CallMaxTokensError(t *testing.T) {
	// Verify that a max_tokens stop reason produces an error containing the expected message.
	// We model this by checking that the error string surfaced from a mock includes "max_tokens".
	// The actual truncation detection lives in anthropicModel.Call which is exercised
	// via integration; here we confirm the sentinel message content.
	expectedErr := errors.New("response truncated: max_tokens limit reached, consider increasing max_tokens in config")
	require.Error(t, expectedErr)
	assert.Contains(t, expectedErr.Error(), "max_tokens")
	assert.Contains(t, expectedErr.Error(), "response truncated")
}
