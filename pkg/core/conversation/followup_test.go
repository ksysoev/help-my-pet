package conversation

import (
	"testing"

	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFollowUpQuestionnaireState(t *testing.T) {
	tests := []struct {
		name        string
		initPrompt  string
		questions   []message.Question
		expectedLen int
	}{
		{
			name:        "empty questions",
			initPrompt:  "Initial Prompt",
			questions:   []message.Question{},
			expectedLen: 0,
		},
		{
			name:       "single question",
			initPrompt: "Initial Prompt",
			questions: []message.Question{
				{Text: "What is your name?"},
			},
			expectedLen: 1,
		},
		{
			name:       "multiple questions",
			initPrompt: "Initial Prompt",
			questions: []message.Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)

			// Assert
			require.NotNil(t, state, "state should not be nil")
			assert.Equal(t, tt.initPrompt, state.InitialPrompt, "initial prompt should match")
			assert.Equal(t, tt.expectedLen, len(state.QAPairs), "number of QAPairs should match the expected length")

			// Validate QAPairs
			for i, question := range tt.questions {
				assert.Equal(t, question.Text, state.QAPairs[i].Question.Text, "Question text should match")
				assert.Empty(t, state.QAPairs[i].Answer, "Answer should be empty initially")
			}
		})
	}
}

func TestFollowUpQuestionnaire_GetCurrentQuestion(t *testing.T) {
	tests := []struct {
		name        string
		initPrompt  string
		questions   []message.Question
		expectedLen int
	}{
		{
			name:       "single question",
			initPrompt: "Initial Prompt",
			questions: []message.Question{
				{Text: "What is your name?"},
			},
			expectedLen: 1,
		},
		{
			name:       "multiple questions",
			initPrompt: "Initial Prompt",
			questions: []message.Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)

			// Act
			question, err := state.GetCurrentQuestion()

			// Assert
			require.NoError(t, err, "error should be nil")
			require.NotNil(t, question, "question should not be nil")
			assert.Equal(t, tt.questions[0].Text, question.Text, "question text should match")
		})
	}
}

func TestFollowUpQuestionnaire_GetResults(t *testing.T) {
	tests := []struct {
		name          string
		initPrompt    string
		questions     []message.Question
		answers       []string
		expectError   error
		expectedPairs []QuestionAnswer
	}{
		{
			name:          "empty questionnaire",
			initPrompt:    "Initial Prompt",
			questions:     []message.Question{},
			answers:       []string{},
			expectError:   nil,
			expectedPairs: []QuestionAnswer{},
		},
		{
			name:       "complete questionnaire",
			initPrompt: "Initial Prompt",
			questions: []message.Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			answers:     []string{"John", "30"},
			expectError: nil,
			expectedPairs: []QuestionAnswer{
				{Question: message.Question{Text: "What is your name?"}, Answer: "John"},
				{Question: message.Question{Text: "How old are you?"}, Answer: "30"},
			},
		},
		{
			name:       "incomplete questionnaire",
			initPrompt: "Initial Prompt",
			questions: []message.Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			answers:       []string{"John"},
			expectError:   ErrQuestionnaireIncomplete,
			expectedPairs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)
			for _, answer := range tt.answers {
				_, err := state.ProcessAnswer(answer)
				assert.NoError(t, err, "unexpected error while processing answer")
			}

			// Act
			results, err := state.GetResults()

			// Assert
			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError, "error should match the expected error")
				assert.Nil(t, results, "results should be nil when there's an error")
			} else {
				assert.NoError(t, err, "error should be nil")
				assert.Equal(t, tt.expectedPairs, results, "results should match the expected pairs")
			}
		})
	}
}

func TestFollowUpQuestionnaire_ProcessAnswer(t *testing.T) {
	tests := []struct {
		name          string
		initPrompt    string
		questions     []message.Question
		answers       []string
		expectError   error
		expectedIndex int
	}{
		{
			name:          "empty questionnaire",
			initPrompt:    "Initial Prompt",
			questions:     []message.Question{},
			answers:       []string{},
			expectError:   ErrNoMoreQuestions,
			expectedIndex: 0,
		},
		{
			name:       "complete questionnaire",
			initPrompt: "Initial Prompt",
			questions: []message.Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			answers:       []string{"John"},
			expectError:   nil,
			expectedIndex: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)

			// Act
			for _, answer := range tt.answers {
				_, err := state.ProcessAnswer(answer)
				assert.NoError(t, err, "unexpected error while processing answer")
			}

			// Assert
			_, err := state.ProcessAnswer("Extra Answer")
			assert.ErrorIs(t, err, tt.expectError, "error should match the expected error")
			assert.Equal(t, tt.expectedIndex, state.CurrentIndex, "current index should match the expected index")
		})
	}
}

func TestFollowUpQuestionnaire_ProcessAnswer_TooLongAnswer(t *testing.T) {
	lognAnswer := `
	This is a very long answer that is longer than the maximum allowed length
	This is a very long answer that is longer than the maximum allowed length
	This is a very long answer that is longer than the maximum allowed length
`

	// Arrange
	state := NewFollowUpQuestionnaireState("Initial Prompt", []message.Question{{Text: "What is your name?"}})

	// Act
	done, err := state.ProcessAnswer(lognAnswer)

	// Assert
	assert.ErrorIs(t, err, message.ErrTextTooLong, "error should be ErrTextTooLong")
	assert.False(t, done, "done should be false")
	assert.Equal(t, 0, state.CurrentIndex, "current index should not be incremented")
}
