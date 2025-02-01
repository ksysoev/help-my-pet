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
		request        *UserMessage
		response       *Response
		expectedResult *PetAdviceResponse
		setupMocks     func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation)
		name           string
		errorContains  string
		wantErr        bool
	}{
		{
			name: "successful response with follow-up questions",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
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
			expectedResult: &PetAdviceResponse{
				Message: "Cats need a balanced diet...\n\nHow old is your cat?",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
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
				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful response without questions",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			response: &Response{
				Text:      "Cats need a balanced diet...",
				Questions: []Question{},
			},
			expectedResult: &PetAdviceResponse{
				Message: "Cats need a balanced diet...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
				mockLLM.EXPECT().
					Call(context.Background(), "What food is good for cats?").
					Return(&Response{
						Text:      "Cats need a balanced diet...",
						Questions: []Question{},
					}, nil)
				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty question",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "",
			},
			response: &Response{
				Text:      "I understand you have a pet-related question...",
				Questions: []Question{},
			},
			expectedResult: &PetAdviceResponse{
				Message: "I understand you have a pet-related question...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
				mockLLM.EXPECT().
					Call(context.Background(), "").
					Return(&Response{
						Text:      "I understand you have a pet-related question...",
						Questions: []Question{},
					}, nil)
				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "llm error",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				// Expect save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
				mockLLM.EXPECT().
					Call(context.Background(), "What food is good for cats?").
					Return(nil, fmt.Errorf("llm error"))
			},
			wantErr:       true,
			errorContains: "failed to get AI response",
		},
		{
			name: "repository FindOrCreate error",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(nil, fmt.Errorf("db error"))
			},
			wantErr:       true,
			errorContains: "failed to get conversation",
		},
		{
			name: "rate limit exceeded",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(false, nil)
			},
			wantErr:       true,
			errorContains: "rate limit exceeded for user",
		},
		{
			name: "rate limit check error",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(false, fmt.Errorf("rate limit check failed"))
			},
			wantErr:       true,
			errorContains: "failed to check rate limit",
		},
		{
			name: "repository Save error",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				// Expect first save to fail
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(fmt.Errorf("save error"))
			},
			wantErr:       true,
			errorContains: "failed to save conversation",
		},
		{
			name: "with conversation history",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What about dogs?",
			},
			response: &Response{
				Text:      "Dogs need different food...",
				Questions: []Question{},
			},
			expectedResult: &PetAdviceResponse{
				Message: "Dogs need different food...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				// Add previous conversation
				conversation.AddMessage("user", "What food is good for cats?")
				conversation.AddMessage("assistant", "Cats need a balanced diet...")

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)

				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)

				expectedPrompt := "Previous conversation:\nuser: What food is good for cats?\nassistant: Cats need a balanced diet...\n\nCurrent question: What about dogs?"
				mockLLM.EXPECT().
					Call(context.Background(), expectedPrompt).
					Return(&Response{
						Text:      "Dogs need different food...",
						Questions: []Question{},
					}, nil)

				// Expect second save after LLM response
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
			mockRateLimiter := NewMockRateLimiter(t)
			mockProfileRepo := NewMockPetProfileRepository(t)

			tt.setupMocks(t, mockLLM, mockRepo, mockRateLimiter, conversation)
			svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

			got, err := svc.GetPetAdvice(context.Background(), tt.request)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, got)

			// No need to verify questions storage since they're now stored in Questionnaire struct
		})
	}
}

func TestAIService_GetPetAdvice_Questionnaire(t *testing.T) {
	tests := []struct {
		request        *UserMessage
		expectedResult *PetAdviceResponse
		setupMocks     func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation)
		name           string
		errorContains  string
		wantErr        bool
	}{
		{
			name: "successful questionnaire response with next question",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "2 years old",
			},
			expectedResult: &PetAdviceResponse{
				Message: "Is your cat indoor or outdoor?",
				Answers: []string{"Indoor", "Outdoor"},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				// Setup conversation in questioning state
				questions := []Question{
					{Text: "How old is your cat?"},
					{
						Text:    "Is your cat indoor or outdoor?",
						Answers: []string{"Indoor", "Outdoor"},
					},
				}
				conversation.StartQuestionnaire("Cats need a balanced diet...", questions)

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
				// Expect second save after updating questionnaire state
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful questionnaire completion",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "Indoor",
			},
			expectedResult: &PetAdviceResponse{
				Message: "Based on your answers, here's my advice...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				// Setup conversation in questioning state with last question
				questions := []Question{
					{Text: "How old is your cat?"},
					{
						Text:    "Is your cat indoor or outdoor?",
						Answers: []string{"Indoor", "Outdoor"},
					},
				}
				conversation.StartQuestionnaire("Cats need a balanced diet...", questions)
				_, err := conversation.AddQuestionAnswer("2 years old")
				require.NoError(t, err)

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)

				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)

				expectedPrompt := "Previous conversation:\n\nFollow-up information:\nQuestion: How old is your cat?\nAnswer: 2 years old\nQuestion: Is your cat indoor or outdoor?\nAnswer: Indoor\n"
				mockLLM.EXPECT().
					Call(context.Background(), expectedPrompt).
					Return(&Response{
						Text: "Based on your answers, here's my advice...",
					}, nil)

				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error adding question answer",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "2 years old",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				// Setup conversation in questioning state with no questions
				questions := []Question{}
				conversation.StartQuestionnaire("Cats need a balanced diet...", questions)

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
			},
			wantErr:       true,
			errorContains: "failed to add question answer: no more questions to answer",
		},
		{
			name: "error saving conversation in questionnaire",
			request: &UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "2 years old",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockRateLimiter *MockRateLimiter, conversation *Conversation) {
				questions := []Question{
					{Text: "How old is your cat?"},
					{Text: "Is your cat indoor or outdoor?"},
				}
				conversation.StartQuestionnaire("Cats need a balanced diet...", questions)

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conversation, nil)
				// Expect save after adding user message to fail
				mockRepo.EXPECT().
					Save(context.Background(), conversation).
					Return(fmt.Errorf("save error"))
			},
			wantErr:       true,
			errorContains: "failed to save conversation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := NewMockLLM(t)
			mockRepo := NewMockConversationRepository(t)
			conversation := NewConversation("test-chat")
			mockRateLimiter := NewMockRateLimiter(t)
			mockProfileRepo := NewMockPetProfileRepository(t)

			tt.setupMocks(t, mockLLM, mockRepo, mockRateLimiter, conversation)
			svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

			got, err := svc.GetPetAdvice(context.Background(), tt.request)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, got)
		})
	}
}

func TestAIService_GetPetAdvice_RateLimiterRecordError(t *testing.T) {
	mockLLM := NewMockLLM(t)
	mockRepo := NewMockConversationRepository(t)
	mockRateLimiter := NewMockRateLimiter(t)
	mockProfileRepo := NewMockPetProfileRepository(t)
	conversation := NewConversation("test-chat")

	mockRepo.EXPECT().
		FindOrCreate(context.Background(), "test-chat").
		Return(conversation, nil)

	mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
	mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(fmt.Errorf("record error"))

	svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

	request := &UserMessage{
		UserID: "user123",
		ChatID: "test-chat",
		Text:   "test question",
	}

	_, err := svc.GetPetAdvice(context.Background(), request)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record rate limit")
}

func TestAIService_GetPetAdvice_ContextCancellation(t *testing.T) {
	mockLLM := NewMockLLM(t)
	mockRepo := NewMockConversationRepository(t)
	mockProfileRepo := NewMockPetProfileRepository(t)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context before the call
	cancel()

	expectedPrompt := "test question"

	mockRateLimiter := NewMockRateLimiter(t)
	mockRateLimiter.On("IsNewQuestionAllowed", ctx, "user123").Return(true, nil)
	mockRateLimiter.On("RecordNewQuestion", ctx, "user123").Return(nil)

	conversation := NewConversation("test-chat")
	mockRepo.EXPECT().
		FindOrCreate(ctx, "test-chat").
		Return(conversation, nil)
	// Expect save after adding user message
	mockRepo.EXPECT().
		Save(ctx, conversation).
		Return(nil)

	mockLLM.EXPECT().
		Call(ctx, expectedPrompt).
		Return(nil, context.Canceled)

	mockProfileRepo.EXPECT().GetProfiles("user123").Return(nil, nil)

	svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

	request := &UserMessage{
		UserID: "user123",
		ChatID: "test-chat",
		Text:   "test question",
	}

	_, err := svc.GetPetAdvice(ctx, request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get AI response")
}

func TestNewAIService(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockLLM := NewMockLLM(t)
		mockRepo := NewMockConversationRepository(t)
		mockRateLimiter := NewMockRateLimiter(t)
		mockProfileRepo := NewMockPetProfileRepository(t)
		svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)
		require.NotNil(t, svc)
		assert.Equal(t, mockLLM, svc.llm)
		assert.Equal(t, mockRepo, svc.repo)
		assert.Equal(t, mockRateLimiter, svc.rateLimiter)
	})
}
