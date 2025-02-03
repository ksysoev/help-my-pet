package conversation

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
			content:        "Text",
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
					conv.AddMessage(tt.role, fmt.Sprintf("Text %d", i))
				}
				// Check if the first message in the history is the correct one
				assert.Equal(t, fmt.Sprintf("Text %d", tt.messageCount-MaxMessageHistory), conv.Messages[0].Content)
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

func TestConversation_MessageHistory(t *testing.T) {
	conv := NewConversation("test-id")

	// Add total of 8 messages (0 to 7)
	for i := 0; i < MaxMessageHistory+3; i++ {
		conv.AddMessage("user", fmt.Sprintf("Text %d", i))
	}

	// Should have last 5 messages (3,4,5,6,7)
	assert.Len(t, conv.Messages, MaxMessageHistory)

	// First message should be 3
	assert.Equal(t, "Text 3", conv.Messages[0].Content)
	// Last message should be 7
	assert.Equal(t, "Text 7", conv.Messages[MaxMessageHistory-1].Content)
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

				err := conv.StartFollowUpQuestions("Initial prompt", questions)
				require.NoError(t, err)

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

				err := conv.StartFollowUpQuestions("Initial prompt", questions)
				require.NoError(t, err)

				_, err = conv.AddQuestionAnswer("Dog")
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
				conv.State = StateFollowUpQuestioning // Set state but don't initialize questionnaire
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
				err := conv.StartFollowUpQuestions("Initial prompt", questions)
				require.NoError(t, err)
				_, err = conv.AddQuestionAnswer("Dog")
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
				conv.State = StateFollowUpQuestioning
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

				err := conv.StartFollowUpQuestions("Initial prompt", questions)
				require.NoError(t, err)

				_, err = conv.AddQuestionAnswer("Dog")
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
