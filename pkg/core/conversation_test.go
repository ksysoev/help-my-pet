package core

import (
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
	assert.NotZero(t, conv.CreatedAt)
	assert.Equal(t, conv.CreatedAt, conv.UpdatedAt)
	assert.Equal(t, StateNormal, conv.State)
	assert.Nil(t, conv.Questionnaire)
}

func TestConversation_AddMessage(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		content string
	}{
		{
			name:    "add user message",
			role:    "user",
			content: "Hello",
		},
		{
			name:    "add assistant message",
			role:    "assistant",
			content: "Hi there!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConversation("test-id")
			beforeAdd := time.Now()
			time.Sleep(time.Millisecond) // Ensure time difference

			conv.AddMessage(tt.role, tt.content)

			assert.Len(t, conv.Messages, 1)
			msg := conv.Messages[0]
			assert.Equal(t, tt.role, msg.Role)
			assert.Equal(t, tt.content, msg.Content)
			assert.NotZero(t, msg.Timestamp)
			assert.True(t, msg.Timestamp.After(beforeAdd))
			assert.True(t, conv.UpdatedAt.After(conv.CreatedAt))
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
	assert.Equal(t, questions, conv.Questionnaire.Questions)
	assert.Equal(t, 0, conv.Questionnaire.CurrentIndex)
	assert.Len(t, conv.Questionnaire.Answers, len(questions))
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
				assert.Equal(t, tt.answer, conv.Questionnaire.Answers[tt.wantNextIndex-1])
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
			prompt, answers, err := conv.GetQuestionnaireResult()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, prompt)
				assert.Nil(t, answers)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "Initial prompt", prompt)
				assert.Equal(t, []string{"Dog", "2 years"}, answers)
			}
		})
	}
}
