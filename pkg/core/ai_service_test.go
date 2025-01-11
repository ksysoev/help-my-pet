package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAIService_GetPetAdvice(t *testing.T) {
	tests := []struct {
		err      error
		name     string
		query    string
		response string
		wantErr  bool
	}{
		{
			name:     "successful response",
			query:    "What food is good for cats?",
			response: "Cats need a balanced diet...",
			err:      nil,
			wantErr:  false,
		},
		{
			name:     "empty question",
			query:    "",
			response: "I understand you have a pet-related question...",
			err:      nil,
			wantErr:  false,
		},
		{
			name:     "llm error",
			query:    "What food is good for cats?",
			response: "",
			err:      fmt.Errorf("llm error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := NewMockLLM(t)
			expectedPrompt := fmt.Sprintf(`You are a helpful veterinary AI assistant. Please provide accurate, helpful, and compassionate advice for the following pet-related question. If the question involves a serious medical condition, always recommend consulting with a veterinarian.

Question: %s

Please provide a clear and informative response:`, tt.query)

			// Setup mock expectations
			mockLLM.EXPECT().
				Call(context.Background(), expectedPrompt, mock.Anything, mock.Anything).
				Return(tt.response, tt.err)

			svc := &AIService{
				llm:   mockLLM,
				model: "test-model",
			}

			got, err := svc.GetPetAdvice(context.Background(), tt.query)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.response, got)
		})
	}
}

func TestNewAIService(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockLLM := NewMockLLM(t)
		svc := NewAIService(mockLLM, "test-model")
		require.NotNil(t, svc)
		assert.Equal(t, "test-model", svc.model)
		assert.Equal(t, mockLLM, svc.llm)
	})
}

func TestAIService_GetPetAdvice_ContextCancellation(t *testing.T) {
	mockLLM := NewMockLLM(t)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context before the call
	cancel()

	expectedPrompt := fmt.Sprintf(`You are a helpful veterinary AI assistant. Please provide accurate, helpful, and compassionate advice for the following pet-related question. If the question involves a serious medical condition, always recommend consulting with a veterinarian.

Question: %s

Please provide a clear and informative response:`, "test question")

	mockLLM.EXPECT().
		Call(ctx, expectedPrompt, mock.Anything, mock.Anything).
		Return("", context.Canceled)

	svc := &AIService{
		llm:   mockLLM,
		model: "test-model",
	}

	_, err := svc.GetPetAdvice(ctx, "test question")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get AI response")
}
