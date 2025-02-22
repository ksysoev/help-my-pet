package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core/conversation"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIService_ProcessMessage(t *testing.T) {
	tests := []struct {
		request        *message.UserMessage
		response       *message.LLMResult
		expectedResult *message.Response
		setupMocks     func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conversation *conversation.Conversation)
		name           string
		errorContains  string
		wantErr        bool
	}{
		{
			name: "successful response with follow-up questions",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			response: &message.LLMResult{
				Text: "Cats need a balanced diet...",
				Questions: []message.Question{
					{Text: "How old is your cat?"},
					{
						Text:    "Is your cat indoor or outdoor?",
						Answers: []string{"Indoor", "Outdoor"},
					},
				},
			},
			expectedResult: &message.Response{
				Message: "Cats need a balanced diet...\n\nHow old is your cat?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockProfileRepo.EXPECT().GetCurrentProfile(context.Background(), "user123").Return(nil, ErrProfileNotFound)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
				expectedPrompt := "\nCurrent question: What food is good for cats?"
				mockLLM.EXPECT().
					Analyze(context.Background(), expectedPrompt, []*message.Image(nil)).
					Return(&message.LLMResult{
						Text: "Cats need a balanced diet...",
						Questions: []message.Question{
							{Text: "How old is your cat?"},
							{
								Text:    "Is your cat indoor or outdoor?",
								Answers: []string{"Indoor", "Outdoor"},
							},
						},
					}, nil)
				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful response without questions",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			response: &message.LLMResult{
				Text:      "Cats need a balanced diet...",
				Questions: []message.Question{},
			},
			expectedResult: &message.Response{
				Message: "Cats need a balanced diet...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockProfileRepo.EXPECT().GetCurrentProfile(context.Background(), "user123").Return(nil, ErrProfileNotFound)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
				expectedPrompt := "\nCurrent question: What food is good for cats?"
				mockLLM.EXPECT().
					Analyze(context.Background(), expectedPrompt, []*message.Image(nil)).
					Return(&message.LLMResult{
						Text:      "Cats need a balanced diet...",
						Questions: []message.Question{},
					}, nil)
				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty question",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "",
			},
			response: &message.LLMResult{
				Text:      "I understand you have a pet-related question...",
				Questions: []message.Question{},
			},
			expectedResult: &message.Response{
				Message: "I understand you have a pet-related question...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockProfileRepo.EXPECT().GetCurrentProfile(context.Background(), "user123").Return(nil, ErrProfileNotFound)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
				expectedPrompt := "\nCurrent question: "
				mockLLM.EXPECT().
					Analyze(context.Background(), expectedPrompt, []*message.Image(nil)).
					Return(&message.LLMResult{
						Text:      "I understand you have a pet-related question...",
						Questions: []message.Question{},
					}, nil)
				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "llm error",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockProfileRepo.EXPECT().GetCurrentProfile(context.Background(), "user123").Return(nil, ErrProfileNotFound)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				// Expect save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
				expectedPrompt := "\nCurrent question: What food is good for cats?"
				mockLLM.EXPECT().
					Analyze(context.Background(), expectedPrompt, []*message.Image(nil)).
					Return(nil, fmt.Errorf("llm error"))
			},
			wantErr:       true,
			errorContains: "failed to get AI response",
		},
		{
			name: "repository FindOrCreate error",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(nil, fmt.Errorf("db error"))
			},
			wantErr:       true,
			errorContains: "failed to get conversation",
		},
		{
			name: "rate limit exceeded",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(false, nil)
			},
			wantErr:       true,
			errorContains: "rate limit exceeded for user",
		},
		{
			name: "rate limit check error",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(false, fmt.Errorf("rate limit check failed"))
			},
			wantErr:       true,
			errorContains: "failed to check rate limit",
		},
		{
			name: "repository Save error",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What food is good for cats?",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				// Expect first save to fail
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(fmt.Errorf("save error"))
			},
			wantErr:       true,
			errorContains: "failed to save conversation",
		},
		{
			name: "with conversation history",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "What about dogs?",
			},
			response: &message.LLMResult{
				Text:      "Dogs need different food...",
				Questions: []message.Question{},
			},
			expectedResult: &message.Response{
				Message: "Dogs need different food...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
				mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(nil)
				mockProfileRepo.EXPECT().GetCurrentProfile(context.Background(), "user123").Return(nil, ErrProfileNotFound)
				// Add previous conversation
				conv.AddMessage("user", "What food is good for cats?")
				conv.AddMessage("assistant", "Cats need a balanced diet...")

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)

				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)

				expectedPrompt := "Previous conversation:\nuser: What food is good for cats?\nassistant: Cats need a balanced diet...\n\nCurrent question: What about dogs?"
				mockLLM.EXPECT().
					Analyze(context.Background(), expectedPrompt, []*message.Image(nil)).
					Return(&message.LLMResult{
						Text:      "Dogs need different food...",
						Questions: []message.Question{},
						Media:     "test media",
					}, nil)

				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := NewMockLLM(t)
			mockRepo := NewMockConversationRepository(t)
			mockProfileRepo := NewMockPetProfileRepository(t)
			conv := conversation.NewConversation("test-chat")
			mockRateLimiter := NewMockRateLimiter(t)

			tt.setupMocks(t, mockLLM, mockRepo, mockProfileRepo, mockRateLimiter, conv)
			svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

			got, err := svc.ProcessMessage(context.Background(), tt.request)
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

func TestAIService_ProcessMessage_Questionnaire(t *testing.T) {
	tests := []struct {
		request        *message.UserMessage
		expectedResult *message.Response
		setupMocks     func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation)
		name           string
		errorContains  string
		wantErr        bool
	}{
		{
			name: "successful questionnaire response with next question",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "2 years old",
			},
			expectedResult: &message.Response{
				Message: "Is your cat indoor or outdoor?",
				Answers: []string{"Indoor", "Outdoor"},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				// Setup conversation in questioning state
				questions := []message.Question{
					{Text: "How old is your cat?"},
					{
						Text:    "Is your cat indoor or outdoor?",
						Answers: []string{"Indoor", "Outdoor"},
					},
				}
				err := conv.StartFollowUpQuestions("Cats need a balanced diet...", questions)
				require.NoError(t, err)

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
				// Expect second save after updating questionnaire state
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful questionnaire completion",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "Indoor",
			},
			expectedResult: &message.Response{
				Message: "Based on your answers, here's my advice...",
				Answers: []string{},
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				// Setup conversation in questioning state with last question
				questions := []message.Question{
					{Text: "How old is your cat?"},
					{
						Text:    "Is your cat indoor or outdoor?",
						Answers: []string{"Indoor", "Outdoor"},
					},
				}

				err := conv.StartFollowUpQuestions("Cats need a balanced diet...", questions)
				require.NoError(t, err)

				_, err = conv.AddQuestionAnswer("2 years old")
				require.NoError(t, err)

				mockProfileRepo.EXPECT().GetCurrentProfile(context.Background(), "user123").Return(nil, ErrProfileNotFound)

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)

				// Expect first save after adding user message
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)

				expectedPrompt := "\nFollow-up information:\nQuestion: How old is your cat?\nAnswer: 2 years old\nQuestion: Is your cat indoor or outdoor?\nAnswer: Indoor\n"
				mockLLM.EXPECT().
					Report(context.Background(), expectedPrompt).
					Return(&message.LLMResult{
						Text: "Based on your answers, here's my advice...",
					}, nil)

				// Expect second save after LLM response
				mockRepo.EXPECT().
					Save(context.Background(), conv).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error saving conversation in questionnaire",
			request: &message.UserMessage{
				UserID: "user123",
				ChatID: "test-chat",
				Text:   "2 years old",
			},
			setupMocks: func(t *testing.T, mockLLM *MockLLM, mockRepo *MockConversationRepository, mockProfileRepo *MockPetProfileRepository, mockRateLimiter *MockRateLimiter, conv *conversation.Conversation) {
				questions := []message.Question{
					{Text: "How old is your cat?"},
					{Text: "Is your cat indoor or outdoor?"},
				}

				err := conv.StartFollowUpQuestions("Cats need a balanced diet...", questions)
				require.NoError(t, err)

				mockRepo.EXPECT().
					FindOrCreate(context.Background(), "test-chat").
					Return(conv, nil)
				// Expect save after adding user message to fail
				mockRepo.EXPECT().
					Save(context.Background(), conv).
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
			mockProfileRepo := NewMockPetProfileRepository(t)
			conv := conversation.NewConversation("test-chat")
			mockRateLimiter := NewMockRateLimiter(t)

			tt.setupMocks(t, mockLLM, mockRepo, mockProfileRepo, mockRateLimiter, conv)
			svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

			got, err := svc.ProcessMessage(context.Background(), tt.request)
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

func TestAIService_ProcessMessage_RateLimiterRecordError(t *testing.T) {
	mockLLM := NewMockLLM(t)
	mockRepo := NewMockConversationRepository(t)
	mockProfileRepo := NewMockPetProfileRepository(t)
	mockRateLimiter := NewMockRateLimiter(t)
	conv := conversation.NewConversation("test-chat")

	mockRepo.EXPECT().
		FindOrCreate(context.Background(), "test-chat").
		Return(conv, nil)

	mockRateLimiter.On("IsNewQuestionAllowed", context.Background(), "user123").Return(true, nil)
	mockRateLimiter.On("RecordNewQuestion", context.Background(), "user123").Return(fmt.Errorf("record error"))

	svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

	request := &message.UserMessage{
		UserID: "user123",
		ChatID: "test-chat",
		Text:   "test question",
	}

	_, err := svc.ProcessMessage(context.Background(), request)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record rate limit")
}

func TestAIService_ProcessMessage_ContextCancellation(t *testing.T) {
	mockLLM := NewMockLLM(t)
	mockRepo := NewMockConversationRepository(t)
	mockProfileRepo := NewMockPetProfileRepository(t)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context before the call
	cancel()

	mockRateLimiter := NewMockRateLimiter(t)
	mockRateLimiter.On("IsNewQuestionAllowed", ctx, "user123").Return(true, nil)
	mockRateLimiter.On("RecordNewQuestion", ctx, "user123").Return(nil)
	mockProfileRepo.EXPECT().GetCurrentProfile(ctx, "user123").Return(nil, ErrProfileNotFound)

	conv := conversation.NewConversation("test-chat")
	mockRepo.EXPECT().
		FindOrCreate(ctx, "test-chat").
		Return(conv, nil)
	// Expect save after adding user message
	mockRepo.EXPECT().
		Save(ctx, conv).
		Return(nil)

	mockLLM.EXPECT().
		Analyze(ctx, "\nCurrent question: test question", []*message.Image(nil)).
		Return(nil, context.Canceled)

	svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)

	request := &message.UserMessage{
		UserID: "user123",
		ChatID: "test-chat",
		Text:   "test question",
	}

	_, err := svc.ProcessMessage(ctx, request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get AI response")
}

func TestNewAIService(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockLLM := NewMockLLM(t)
		mockRepo := NewMockConversationRepository(t)
		mockProfileRepo := NewMockPetProfileRepository(t)
		mockRateLimiter := NewMockRateLimiter(t)
		svc := NewAIService(mockLLM, mockRepo, mockProfileRepo, mockRateLimiter)
		require.NotNil(t, svc)
		assert.Equal(t, mockLLM, svc.llm)
		assert.Equal(t, mockRepo, svc.repo)
		assert.Equal(t, mockProfileRepo, svc.profileRepo)
		assert.Equal(t, mockRateLimiter, svc.rateLimiter)
	})
}
