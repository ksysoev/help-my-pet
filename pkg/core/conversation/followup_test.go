package conversation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFollowUpQuestionnaireState(t *testing.T) {
	tests := []struct {
		name        string
		initPrompt  string
		questions   []Question
		expectedLen int
	}{
		{
			name:        "empty questions",
			initPrompt:  "Initial Prompt",
			questions:   []Question{},
			expectedLen: 0,
		},
		{
			name:       "single question",
			initPrompt: "Initial Prompt",
			questions: []Question{
				{Text: "What is your name?"},
			},
			expectedLen: 1,
		},
		{
			name:       "multiple questions",
			initPrompt: "Initial Prompt",
			questions: []Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			require := require.New(t)
			assert := assert.New(t)

			// Act
			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)

			// Assert
			require.NotNil(state, "state should not be nil")
			assert.Equal(tt.initPrompt, state.InitialPrompt, "initial prompt should match")
			assert.Equal(tt.expectedLen, len(state.QAPairs), "number of QAPairs should match the expected length")

			// Validate QAPairs
			for i, question := range tt.questions {
				assert.Equal(question.Text, state.QAPairs[i].Question.Text, "Question text should match")
				assert.Empty(state.QAPairs[i].Answer, "Answer should be empty initially")
			}
		})
	}
}

func TestFollowUpQuestionnaire_GetCurrentQuestion(t *testing.T) {
	tests := []struct {
		name        string
		initPrompt  string
		questions   []Question
		expectedLen int
	}{
		{
			name:       "single question",
			initPrompt: "Initial Prompt",
			questions: []Question{
				{Text: "What is your name?"},
			},
			expectedLen: 1,
		},
		{
			name:       "multiple questions",
			initPrompt: "Initial Prompt",
			questions: []Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			require := require.New(t)
			assert := assert.New(t)

			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)

			// Act
			question, err := state.GetCurrentQuestion()

			// Assert
			require.NoError(err, "error should be nil")
			require.NotNil(question, "question should not be nil")
			assert.Equal(tt.questions[0].Text, question.Text, "question text should match")
		})
	}
}

func TestFollowUpQuestionnaire_GetResults(t *testing.T) {
	tests := []struct {
		name          string
		initPrompt    string
		questions     []Question
		answers       []string
		expectError   error
		expectedPairs []QuestionAnswer
	}{
		{
			name:          "empty questionnaire",
			initPrompt:    "Initial Prompt",
			questions:     []Question{},
			answers:       []string{},
			expectError:   nil,
			expectedPairs: []QuestionAnswer{},
		},
		{
			name:       "complete questionnaire",
			initPrompt: "Initial Prompt",
			questions: []Question{
				{Text: "What is your name?"},
				{Text: "How old are you?"},
			},
			answers:     []string{"John", "30"},
			expectError: nil,
			expectedPairs: []QuestionAnswer{
				{Question: Question{Text: "What is your name?"}, Answer: "John"},
				{Question: Question{Text: "How old are you?"}, Answer: "30"},
			},
		},
		{
			name:       "incomplete questionnaire",
			initPrompt: "Initial Prompt",
			questions: []Question{
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
			assert := assert.New(t)

			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)
			for _, answer := range tt.answers {
				_, err := state.ProcessAnswer(answer)
				assert.NoError(err, "unexpected error while processing answer")
			}

			// Act
			results, err := state.GetResults()

			// Assert
			if tt.expectError != nil {
				assert.ErrorIs(err, tt.expectError, "error should match the expected error")
				assert.Nil(results, "results should be nil when there's an error")
			} else {
				assert.NoError(err, "error should be nil")
				assert.Equal(tt.expectedPairs, results, "results should match the expected pairs")
			}
		})
	}
}

func TestFollowUpQuestionnaire_ProcessAnswer(t *testing.T) {
	tests := []struct {
		name          string
		initPrompt    string
		questions     []Question
		answers       []string
		expectError   error
		expectedIndex int
	}{
		{
			name:          "empty questionnaire",
			initPrompt:    "Initial Prompt",
			questions:     []Question{},
			answers:       []string{},
			expectError:   ErrNoMoreQuestions,
			expectedIndex: 0,
		},
		{
			name:       "complete questionnaire",
			initPrompt: "Initial Prompt",
			questions: []Question{
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
			assert := assert.New(t)

			state := NewFollowUpQuestionnaireState(tt.initPrompt, tt.questions)

			// Act
			for _, answer := range tt.answers {
				_, err := state.ProcessAnswer(answer)
				assert.NoError(err, "unexpected error while processing answer")
			}

			// Assert
			_, err := state.ProcessAnswer("Extra Answer")
			assert.ErrorIs(err, tt.expectError, "error should match the expected error")
			assert.Equal(tt.expectedIndex, state.CurrentIndex, "current index should match the expected index")
		})
	}
}
