package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
				Call(context.Background(), expectedPrompt).
				Return(tt.response, tt.err)

			mockRepo := NewMockConversationRepository(t)
			conversation := NewConversation("test-chat")

			mockRepo.EXPECT().
				FindOrCreate(context.Background(), "test-chat").
				Return(conversation, nil)

			if !tt.wantErr {
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			}

			svc := &AIService{
				llm:  mockLLM,
				repo: mockRepo,
			}

			got, err := svc.GetPetAdvice(context.Background(), "test-chat", tt.query)
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
		mockRepo := NewMockConversationRepository(t)
		svc := NewAIService(mockLLM, mockRepo)
		require.NotNil(t, svc)
		assert.Equal(t, mockLLM, svc.llm)
		assert.Equal(t, mockRepo, svc.repo)
	})
}

func TestAIService_Start(t *testing.T) {
	t.Run("successful start", func(t *testing.T) {
		mockLLM := NewMockLLM(t)
		mockRepo := NewMockConversationRepository(t)
		svc := NewAIService(mockLLM, mockRepo)

		response, err := svc.Start(context.Background())
		require.NoError(t, err)
		assert.Contains(t, response, "Welcome to Help My Pet Bot!")
		assert.Contains(t, response, "I'm your personal pet care assistant")
		assert.Contains(t, response, "To get started, just ask me any question about your pet!")
	})

	t.Run("with cancelled context", func(t *testing.T) {
		mockLLM := NewMockLLM(t)
		mockRepo := NewMockConversationRepository(t)
		svc := NewAIService(mockLLM, mockRepo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		response, err := svc.Start(ctx)
		require.NoError(t, err)
		assert.Contains(t, response, "Welcome to Help My Pet Bot!")
	})
}

func TestAIService_GetPetAdvice_ContextCancellation(t *testing.T) {
	mockLLM := NewMockLLM(t)
	mockRepo := NewMockConversationRepository(t)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context before the call
	cancel()

	expectedPrompt := fmt.Sprintf(`You are a helpful veterinary AI assistant. Please provide accurate, helpful, and compassionate advice for the following pet-related question. If the question involves a serious medical condition, always recommend consulting with a veterinarian.

Question: %s

Please provide a clear and informative response:`, "test question")

	mockLLM.EXPECT().
		Call(ctx, expectedPrompt).
		Return("", context.Canceled)

	conversation := NewConversation("test-chat")
	mockRepo.EXPECT().
		FindOrCreate(ctx, "test-chat").
		Return(conversation, nil)

	svc := &AIService{
		llm:  mockLLM,
		repo: mockRepo,
	}

	_, err := svc.GetPetAdvice(ctx, "test-chat", "test question")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get AI response")
}
