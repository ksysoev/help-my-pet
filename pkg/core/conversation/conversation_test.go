package conversation

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConversation(t *testing.T) {
	id := "test-GetID"
	conv := NewConversation(id)

	assert.Equal(t, id, conv.GetID())
	assert.Empty(t, conv.Messages)
	assert.Equal(t, StateNormal, conv.GetState())
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
			conv := NewConversation("test-GetID")
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
	tests := []struct {
		name     string
		messages []struct {
			role    string
			content string
		}
		skip       int
		wantOutput string
	}{
		{
			name: "retrieve full history",
			messages: []struct {
				role    string
				content string
			}{
				{"user", "Hello"},
				{"assistant", "Hi there!"},
				{"user", "How are you?"},
			},
			skip:       0,
			wantOutput: "Previous conversation:\nuser: Hello\nassistant: Hi there!\nuser: How are you?\n",
		},
		{
			name: "skip one message",
			messages: []struct {
				role    string
				content string
			}{
				{"user", "Hello"},
				{"assistant", "Hi there!"},
				{"user", "How are you?"},
			},
			skip:       1,
			wantOutput: "Previous conversation:\nuser: Hello\nassistant: Hi there!\n",
		},
		{
			name: "skip all messages",
			messages: []struct {
				role    string
				content string
			}{
				{"user", "Hello"},
				{"assistant", "Hi there!"},
			},
			skip:       2,
			wantOutput: "",
		},
		{
			name:       "no messages",
			messages:   []struct{ role, content string }{},
			skip:       0,
			wantOutput: "",
		},
		{
			name: "skip more messages than exist",
			messages: []struct {
				role    string
				content string
			}{
				{"user", "Hello"},
			},
			skip:       5,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConversation("test-GetID")

			for _, msg := range tt.messages {
				conv.AddMessage(msg.role, msg.content)
			}

			output := conv.History(tt.skip)
			assert.Equal(t, tt.wantOutput, output)
		})
	}
}

func TestConversation_MessageHistory(t *testing.T) {
	conv := NewConversation("test-GetID")

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
		wantQuestion *message.Question
		name         string
		wantErr      bool
	}{
		{
			name: "get first question",
			setupConv: func() *Conversation {
				conv := NewConversation("test-GetID")
				questions := []message.Question{
					{Text: "What type of pet do you have?", Answers: []string{"Dog", "Cat"}},
					{Text: "How old is your pet?"},
				}

				err := conv.StartFollowUpQuestions("Initial prompt", questions)
				require.NoError(t, err)

				return conv
			},
			wantErr:      false,
			wantQuestion: &message.Question{Text: "What type of pet do you have?", Answers: []string{"Dog", "Cat"}},
		},
		{
			name: "no questionnaire started",
			setupConv: func() *Conversation {
				return NewConversation("test-GetID")
			},
			wantErr:      true,
			wantQuestion: nil,
		},
		{
			name: "all questions answered",
			setupConv: func() *Conversation {
				conv := NewConversation("test-GetID")
				questions := []message.Question{
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
				conv := NewConversation("test-GetID")
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
				conv := NewConversation("test-GetID")
				questions := []message.Question{{Text: "What type of pet do you have?"}}
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
				conv := NewConversation("test-GetID")
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
				conv := NewConversation("test-GetID")
				questions := []message.Question{
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
				return NewConversation("test-GetID")
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
						Question: message.Question{Text: "What type of pet do you have?"},
						Answer:   "Dog",
					},
					{
						Question: message.Question{Text: "How old is your pet?"},
						Answer:   "2 years",
					},
				}
				assert.Equal(t, expectedQA, answers)
			}
		})
	}
}

func TestConversationUnmarshal_NormalState(t *testing.T) {
	data, err := json.Marshal(struct {
		ID            string
		State         ConversationState
		Messages      []Message
		Questionnaire json.RawMessage `json:"questionnaire"`
	}{
		ID:    "test-GetID",
		State: StateNormal,
		Messages: []Message{
			{Content: "hello"},
		},
	})
	require.NoError(t, err)

	conv, err := Unmarshal(data)
	assert.NoError(t, err)
	assert.Equal(t, "test-GetID", conv.GetID())
	assert.Equal(t, StateNormal, conv.GetState())
	assert.Equal(t, 1, len(conv.Messages))
	assert.Nil(t, conv.Questionnaire)
}

func TestConversationUnmarshal_CompletedState(t *testing.T) {
	data, err := json.Marshal(struct {
		ID            string
		State         ConversationState
		Messages      []Message
		Questionnaire json.RawMessage `json:"questionnaire"`
	}{
		ID:    "test-GetID",
		State: StateCompleted,
		Messages: []Message{
			{Content: "finished"},
		},
	})
	require.NoError(t, err)

	conv, err := Unmarshal(data)
	assert.NoError(t, err)
	assert.Equal(t, "test-GetID", conv.GetID())
	assert.Equal(t, StateNormal, conv.GetState()) // completed falls back to normal
	assert.Equal(t, 1, len(conv.Messages))
	assert.Nil(t, conv.Questionnaire)
}

func TestConversationUnmarshal_ProfileState(t *testing.T) {

	mockQuestionnaire, err := json.Marshal(struct {
		SomeField string
	}{
		SomeField: "test",
	})
	require.NoError(t, err)

	data, err := json.Marshal(struct {
		ID            string
		State         ConversationState
		Messages      []Message
		Questionnaire json.RawMessage `json:"questionnaire"`
	}{
		ID:            "test-GetID",
		State:         StatePetProfileQuestioning,
		Messages:      []Message{},
		Questionnaire: mockQuestionnaire,
	})
	require.NoError(t, err)

	conv, err := Unmarshal(data)
	assert.NoError(t, err)
	assert.Equal(t, "test-GetID", conv.GetID())
	assert.Equal(t, StatePetProfileQuestioning, conv.GetState())
	assert.NotNil(t, conv.Questionnaire)
}

func TestConversationUnmarshal_FollowUpState(t *testing.T) {
	mockQuestionnaire, err := json.Marshal(struct {
		InitialPrompt string
		Questions     []message.Question
	}{
		InitialPrompt: "some prompt",
		Questions: []message.Question{
			{Text: "Q1"},
		},
	})
	require.NoError(t, err)

	data, err := json.Marshal(struct {
		ID            string
		State         ConversationState
		Messages      []Message
		Questionnaire json.RawMessage `json:"questionnaire"`
	}{
		ID:            "test-GetID",
		State:         StateFollowUpQuestioning,
		Messages:      []Message{},
		Questionnaire: mockQuestionnaire,
	})
	require.NoError(t, err)

	conv, err := Unmarshal(data)
	assert.NoError(t, err)
	assert.Equal(t, "test-GetID", conv.GetID())
	assert.Equal(t, StateFollowUpQuestioning, conv.GetState())
	assert.NotNil(t, conv.Questionnaire)
}

func TestConversationUnmarshal_InvalidJSON(t *testing.T) {
	_, err := Unmarshal([]byte("invalid json"))
	assert.Error(t, err)
}

func TestConversationUnmarshal_InvalidJSONProfileState(t *testing.T) {
	_, err := Unmarshal([]byte(`{"GetID":"test-GetID","state":"pet_profile_questioning","messages":[],"questionnaire":"invalid json"}`))
	assert.Error(t, err)
}

func TestConversationUnmarshal_InvalidJSONFollowUpState(t *testing.T) {
	_, err := Unmarshal([]byte(`{"GetID":"test-GetID","state":"follow_up_questioning","messages":[],"questionnaire":"invalid json"}`))
	assert.Error(t, err)
}

func TestConversationUnmarshal_UnknownState(t *testing.T) {
	data, err := json.Marshal(struct {
		ID            string
		State         ConversationState
		Messages      []Message
		Questionnaire json.RawMessage `json:"questionnaire"`
	}{
		ID:    "test-GetID",
		State: "unknown_state",
	})
	require.NoError(t, err)

	conv, err := Unmarshal(data)
	assert.Error(t, err)
	assert.Nil(t, conv)
}

func TestConversationReset(t *testing.T) {
	conv := NewConversation("test-GetID")
	conv.State = StatePetProfileQuestioning
	conv.Questionnaire = &PetProfileStateImpl{}

	conv.CancelQuestionnaire()

	assert.Equal(t, StateNormal, conv.State)
	assert.Nil(t, conv.Questionnaire)
}
