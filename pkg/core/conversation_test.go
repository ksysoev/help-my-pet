package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConversation(t *testing.T) {
	id := "test-id"
	conv := NewConversation(id)

	assert.Equal(t, id, conv.ID)
	assert.Empty(t, conv.Messages)
	assert.Equal(t, StateNormal, conv.State)
	assert.Nil(t, conv.Questionnaire)
}

func TestConversation_AddMessage(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		content        string
		messageCount   int
		expectedLength int
	}{
		{
			name:           "add user message",
			role:           "user",
			content:        "Hello",
			messageCount:   1,
			expectedLength: 1,
		},
		{
			name:           "add assistant message",
			role:           "assistant",
			content:        "Hi there!",
			messageCount:   1,
			expectedLength: 1,
		},
		{
			name:           "exceed max history",
			role:           "user",
			content:        "Message",
			messageCount:   MaxMessageHistory + 2,
			expectedLength: MaxMessageHistory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConversation("test-id")
			beforeAdd := time.Now()
			time.Sleep(time.Millisecond) // Ensure time difference

			if tt.name == "exceed max history" {
				// For max history test, add messages sequentially
				for i := 0; i < tt.messageCount; i++ {
					conv.AddMessage(tt.role, fmt.Sprintf("Message %d", i))
				}
				// Check if the first message in the history is the correct one
				assert.Equal(t, fmt.Sprintf("Message %d", tt.messageCount-MaxMessageHistory), conv.Messages[0].Content)
			} else {
				// For other tests, just add one message
				conv.AddMessage(tt.role, tt.content)
				assert.Equal(t, tt.content, conv.Messages[0].Content)
			}

			assert.Len(t, conv.Messages, tt.expectedLength)
			assert.Equal(t, tt.role, conv.Messages[0].Role)
			assert.NotZero(t, conv.Messages[0].Timestamp)
			assert.True(t, conv.Messages[0].Timestamp.After(beforeAdd))
		})
	}
}

func TestConversation_GetContext(t *testing.T) {
	conv := NewConversation("test-id")
	messages := []struct {
		role    string
		content string
	}{
		{"user", "Hello"},
		{"assistant", "Hi there!"},
		{"user", "How are you?"},
	}

	for _, msg := range messages {
		conv.AddMessage(msg.role, msg.content)
	}

	context := conv.GetContext()
	assert.Len(t, context, len(messages))

	for i, msg := range messages {
		assert.Equal(t, msg.role, context[i].Role)
		assert.Equal(t, msg.content, context[i].Content)
	}
}

func TestConversation_StartQuestionnaire(t *testing.T) {
	conv := NewConversation("test-id")
	initialPrompt := "Let me help you with that."
	questions := []Question{
		{Text: "What type of pet do you have?", Answers: []string{"Dog", "Cat", "Bird"}},
		{Text: "How old is your pet?"},
	}

	conv.StartQuestionnaire(initialPrompt, questions)

	assert.Equal(t, StateQuestioning, conv.State)
	assert.NotNil(t, conv.Questionnaire)
	assert.Equal(t, initialPrompt, conv.Questionnaire.InitialPrompt)
	assert.Len(t, conv.Questionnaire.QAPairs, len(questions))
	for i, q := range questions {
		assert.Equal(t, q, conv.Questionnaire.QAPairs[i].Question)
		assert.Empty(t, conv.Questionnaire.QAPairs[i].Answer)
	}
	assert.Equal(t, 0, conv.Questionnaire.CurrentIndex)
}

func TestConversation_MessageHistory(t *testing.T) {
	conv := NewConversation("test-id")

	// Add total of 8 messages (0 to 7)
	for i := 0; i < MaxMessageHistory+3; i++ {
		conv.AddMessage("user", fmt.Sprintf("Message %d", i))
	}

	// Should have last 5 messages (3,4,5,6,7)
	assert.Len(t, conv.Messages, MaxMessageHistory)

	// First message should be 3
	assert.Equal(t, "Message 3", conv.Messages[0].Content)
	// Last message should be 7
	assert.Equal(t, "Message 7", conv.Messages[MaxMessageHistory-1].Content)
}

func TestConversation_GetCurrentQuestion(t *testing.T) {
	tests := []struct {
		setupConv    func() *Conversation
		wantQuestion *Question
		name         string
		wantErr      bool
	}{
		{
			name: "get first question",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{
					{Text: "What type of pet do you have?", Answers: []string{"Dog", "Cat"}},
					{Text: "How old is your pet?"},
				}
				conv.StartQuestionnaire("Initial prompt", questions)
				return conv
			},
			wantErr:      false,
			wantQuestion: &Question{Text: "What type of pet do you have?", Answers: []string{"Dog", "Cat"}},
		},
		{
			name: "no questionnaire started",
			setupConv: func() *Conversation {
				return NewConversation("test-id")
			},
			wantErr:      true,
			wantQuestion: nil,
		},
		{
			name: "all questions answered",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{
					{Text: "What type of pet do you have?"},
				}
				conv.StartQuestionnaire("Initial prompt", questions)
				_, err := conv.AddQuestionAnswer("Dog")
				require.NoError(t, err)
				return conv
			},
			wantErr:      true,
			wantQuestion: nil,
		},
		{
			name: "questioning state but nil questionnaire",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				conv.State = StateQuestioning // Set state but don't initialize questionnaire
				return conv
			},
			wantErr:      true,
			wantQuestion: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := tt.setupConv()
			question, err := conv.GetCurrentQuestion()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, question)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantQuestion, question)
			}
		})
	}
}

func TestConversation_AddQuestionAnswer_AdditionalCases(t *testing.T) {
	tests := []struct {
		name      string
		setupConv func() *Conversation
		answer    string
		wantErr   bool
	}{
		{
			name: "attempt to answer after completion",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{{Text: "What type of pet do you have?"}}
				conv.StartQuestionnaire("Initial prompt", questions)
				_, err := conv.AddQuestionAnswer("Dog")
				require.NoError(t, err)
				return conv
			},
			answer:  "Another answer",
			wantErr: true,
		},
		{
			name: "attempt to answer with empty questionnaire",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				conv.State = StateQuestioning
				return conv
			},
			answer:  "Answer",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := tt.setupConv()
			_, err := conv.AddQuestionAnswer(tt.answer)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConversation_AddQuestionAnswer(t *testing.T) {
	tests := []struct {
		setupConv     func() *Conversation
		name          string
		answer        string
		wantState     ConversationState
		wantNextIndex int
		wantComplete  bool
		wantErr       bool
	}{
		{
			name: "add first answer",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{
					{Text: "What type of pet do you have?"},
					{Text: "How old is your pet?"},
				}
				conv.StartQuestionnaire("Initial prompt", questions)
				return conv
			},
			answer:        "Dog",
			wantComplete:  false,
			wantErr:       false,
			wantState:     StateQuestioning,
			wantNextIndex: 1,
		},
		{
			name: "complete questionnaire",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{{Text: "What type of pet do you have?"}}
				conv.StartQuestionnaire("Initial prompt", questions)
				return conv
			},
			answer:        "Dog",
			wantComplete:  true,
			wantErr:       false,
			wantState:     StateNormal,
			wantNextIndex: 1,
		},
		{
			name: "answer without questionnaire",
			setupConv: func() *Conversation {
				return NewConversation("test-id")
			},
			answer:       "Dog",
			wantComplete: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := tt.setupConv()
			complete, err := conv.AddQuestionAnswer(tt.answer)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantComplete, complete)
				assert.Equal(t, tt.wantState, conv.State)
				assert.Equal(t, tt.wantNextIndex, conv.Questionnaire.CurrentIndex)
				assert.Equal(t, tt.answer, conv.Questionnaire.QAPairs[tt.wantNextIndex-1].Answer)
			}
		})
	}
}

func TestConversation_GetQuestionnaireResult_AdditionalCases(t *testing.T) {
	tests := []struct {
		setupConv     func() *Conversation
		name          string
		wantAnswers   []string
		wantQuestions []Question
		wantErr       bool
	}{
		{
			name: "partial answers",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{
					{Text: "What type of pet do you have?"},
					{Text: "How old is your pet?"},
				}
				conv.StartQuestionnaire("Initial prompt", questions)
				_, err := conv.AddQuestionAnswer("Dog")
				require.NoError(t, err)
				return conv
			},
			wantErr:     false,
			wantAnswers: []string{"Dog", ""},
			wantQuestions: []Question{
				{Text: "What type of pet do you have?"},
				{Text: "How old is your pet?"},
			},
		},
		{
			name: "questionnaire exists but no answers",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{
					{Text: "What type of pet do you have?"},
				}
				conv.StartQuestionnaire("Initial prompt", questions)
				return conv
			},
			wantErr:     false,
			wantAnswers: []string{""},
			wantQuestions: []Question{
				{Text: "What type of pet do you have?"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := tt.setupConv()
			answers, err := conv.GetQuestionnaireResult()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, answers)
			} else {
				assert.NoError(t, err)
				expectedQA := make([]QuestionAnswer, len(tt.wantAnswers))
				for i, ans := range tt.wantAnswers {
					expectedQA[i] = QuestionAnswer{
						Question: tt.wantQuestions[i],
						Answer:   ans,
					}
				}
				assert.Equal(t, expectedQA, answers)
			}
		})
	}
}

func TestConversation_GetQuestionnaireResult(t *testing.T) {
	tests := []struct {
		setupConv func() *Conversation
		name      string
		wantErr   bool
	}{
		{
			name: "get complete questionnaire result",
			setupConv: func() *Conversation {
				conv := NewConversation("test-id")
				questions := []Question{
					{Text: "What type of pet do you have?"},
					{Text: "How old is your pet?"},
				}
				conv.StartQuestionnaire("Initial prompt", questions)
				_, err := conv.AddQuestionAnswer("Dog")
				require.NoError(t, err)

				_, err = conv.AddQuestionAnswer("2 years")
				require.NoError(t, err)

				return conv
			},
			wantErr: false,
		},
		{
			name: "no questionnaire data",
			setupConv: func() *Conversation {
				return NewConversation("test-id")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := tt.setupConv()
			answers, err := conv.GetQuestionnaireResult()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, answers)
			} else {
				assert.NoError(t, err)
				expectedQA := []QuestionAnswer{
					{
						Question: Question{Text: "What type of pet do you have?"},
						Answer:   "Dog",
					},
					{
						Question: Question{Text: "How old is your pet?"},
						Answer:   "2 years",
					},
				}
				assert.Equal(t, expectedQA, answers)
			}
		})
	}
}
