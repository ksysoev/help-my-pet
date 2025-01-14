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
		setupMocks     func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation)
		name           string
		query          string
		response       *Response
		expectedPrompt string
		errorContains  string
		wantErr        bool
	}{
		{
			name:  "successful response with follow-up questions",
			query: "What food is good for cats?",
			response: &Response{
				Text: "Cats need a balanced diet...",
				Questions: []Question{
					{Text: "How old is your cat?"},
					{
						Text:    "Is your cat indoor or outdoor?",
						Answers: []string{"Indoor", "Outdoor"},
					},
				},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				mockLLM.EXPECT().
					Call(context.Background(), "What food is good for cats?").
					Return(&Response{
						Text: "Cats need a balanced diet...",
						Questions: []Question{
							{Text: "How old is your cat?"},
							{
								Text:    "Is your cat indoor or outdoor?",
								Answers: []string{"Indoor", "Outdoor"},
							},
						},
					}, nil)
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			expectedPrompt: "What food is good for cats?",
			wantErr:        false,
		},
		{
			name:  "successful response without questions",
			query: "What food is good for cats?",
			response: &Response{
				Text:      "Cats need a balanced diet...",
				Questions: []Question{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				mockLLM.EXPECT().
					Call(context.Background(), "What food is good for cats?").
					Return(&Response{
						Text:      "Cats need a balanced diet...",
						Questions: []Question{},
					}, nil)
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			expectedPrompt: "What food is good for cats?",
			wantErr:        false,
		},
		{
			name:  "empty question",
			query: "",
			response: &Response{
				Text:      "I understand you have a pet-related question...",
				Questions: []Question{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				mockLLM.EXPECT().
					Call(context.Background(), "").
					Return(&Response{
						Text:      "I understand you have a pet-related question...",
						Questions: []Question{},
					}, nil)
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			expectedPrompt: "",
			wantErr:        false,
		},
		{
			name:  "llm error",
			query: "What food is good for cats?",
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				mockLLM.EXPECT().
					Call(context.Background(), "What food is good for cats?").
					Return(nil, fmt.Errorf("llm error"))
			},
			expectedPrompt: "What food is good for cats?",
			wantErr:        true,
			errorContains:  "failed to get AI response",
		},
		{
			name:  "repository FindOrCreate error",
			query: "What food is good for cats?",
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(nil, fmt.Errorf("db error"))
			},
			wantErr:       true,
			errorContains: "failed to get conversation",
		},
		{
			name:  "repository Save error",
			query: "What food is good for cats?",
			response: &Response{
				Text:      "Cats need a balanced diet...",
				Questions: []Question{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				mockLLM.EXPECT().
					Call(context.Background(), "What food is good for cats?").
					Return(&Response{
						Text:      "Cats need a balanced diet...",
						Questions: []Question{},
					}, nil)
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(fmt.Errorf("save error"))
			},
			expectedPrompt: "What food is good for cats?",
			wantErr:        true,
			errorContains:  "failed to save conversation",
		},
		{
			name:  "with conversation history",
			query: "What about dogs?",
			response: &Response{
				Text:      "Dogs need different food...",
				Questions: []Question{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, conversation *Conversation) {
				// Add previous conversation
				conversation.AddMessage("user", "What food is good for cats?")
				conversation.AddMessage("assistant", "Cats need a balanced diet...")

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)

				expectedPrompt := "Previous conversation:\nuser: What food is good for cats?\nassistant: Cats need a balanced diet...\n\nCurrent question: What about dogs?"
				mockLLM.EXPECT().
					Call(context.Background(), expectedPrompt).
					Return(&Response{
						Text:      "Dogs need different food...",
						Questions: []Question{},
					}, nil)

				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := NewMockLLM(t)
			mockRepo := NewMockConversationRepository(t)
			conversation := NewConversation("test-chat")

			// Setup mocks based on test case
			tt.setupMocks(t, mockLLM, mockRepo, conversation)

			svc := &AIService{
				llm:  mockLLM,
				repo: mockRepo,
			}

			got, err := svc.GetPetAdvice(context.Background(), "test-chat", tt.query)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.response.Text, got)

			// Verify questions were stored in conversation if present
			if tt.response != nil && len(tt.response.Questions) > 0 {
				messages := conversation.GetContext()
				var foundQuestions bool
				for _, msg := range messages {
					if msg.Role == "assistant_questions" {
						foundQuestions = true
						for _, q := range tt.response.Questions {
							assert.Contains(t, msg.Content, q.Text)
							for _, answer := range q.Answers {
								assert.Contains(t, msg.Content, answer)
							}
						}
					}
				}
				assert.True(t, foundQuestions, "Questions should be stored in conversation")
			}
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

	expectedPrompt := "test question"

	mockLLM.EXPECT().
		Call(ctx, expectedPrompt).
		Return(nil, context.Canceled)

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
